package enriching

import (
	"context"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_enriching "github.com/topfreegames/podium/leaderboard/v2/mocks"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Enricher tests", func() {
	var mux *http.ServeMux
	var server *httptest.Server

	BeforeEach(func() {
		mux = http.NewServeMux()
		server = httptest.NewServer(mux)

	})

	AfterEach(func() {
		server.Close()
	})

	leaderboardID := "leaderboardID"
	tenantID := "tenantID"

	It("should not enrich if no webhook url is configured and cloud save service is disabled", func() {
		cache := mock_enriching.NewMockEnricherCache(gomock.NewController(GinkgoT()))

		enrich := NewEnricher(zap.NewNop(), cache,
			WithCloudSaveDisabled(map[string]bool{tenantID: true}),
		)

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
			{
				PublicID: "publicID2",
			},
		}

		cache.EXPECT().
			Get(gomock.Any(), tenantID, leaderboardID, members).
			Return(nil, false, nil)

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(BeNil())
		Expect(res).To(Equal(members))
	})

	It("should fail if cloud save service call fails (status != 200)", func() {
		mux.HandleFunc(cloudSaveEndpoint, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("{}"))
		})

		cache := mock_enriching.NewMockEnricherCache(gomock.NewController(GinkgoT()))
		enrich := NewEnricher(zap.NewNop(), cache, WithCloudSaveUrl(server.URL))

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
			{
				PublicID: "publicID2",
			},
		}

		cache.EXPECT().
			Get(gomock.Any(), tenantID, leaderboardID, members).
			Return(nil, false, nil)

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(HaveOccurred())
		Expect(res).To(BeNil())
	})

	It("should fail if cloud save service returns invalid json", func() {
		mux.HandleFunc(cloudSaveEndpoint, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("invalid"))
		})

		cache := mock_enriching.NewMockEnricherCache(gomock.NewController(GinkgoT()))
		enrich := NewEnricher(zap.NewNop(), cache, WithCloudSaveUrl(server.URL))

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
			{
				PublicID: "publicID2",
			},
		}

		cache.EXPECT().
			Get(gomock.Any(), tenantID, leaderboardID, members).
			Return(nil, false, nil)

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(HaveOccurred())
		Expect(res).To(BeNil())
	})

	It("should succeed with enrichment from cloud save service", func() {
		mux.HandleFunc(cloudSaveEndpoint, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("{\"documents\": [{\"accountId\": \"publicID\", \"data\": {\"key\": \"value\"}}]}"))
		})

		cache := mock_enriching.NewMockEnricherCache(gomock.NewController(GinkgoT()))
		enrich := NewEnricher(zap.NewNop(), cache, WithCloudSaveUrl(server.URL))

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
		}
		cache.EXPECT().
			Get(gomock.Any(), tenantID, leaderboardID, members).
			Return(nil, false, nil)

		expectedResult := []*model.Member{
			{
				PublicID: "publicID",
				Metadata: map[string]string{
					"key": "value",
				},
			},
		}

		cache.EXPECT().
			Set(gomock.Any(), tenantID, leaderboardID, expectedResult, gomock.Any()).
			Return(nil)

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(BeNil())
		Expect(res).To(Equal(expectedResult))
	})

	It("should return error if webhook call fails", func() {
		mux.HandleFunc(enrichWebhookEndpoint, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("{}"))
		})

		cache := mock_enriching.NewMockEnricherCache(gomock.NewController(GinkgoT()))
		enrich := NewEnricher(zap.NewNop(), cache,
			WithWebhookUrls(map[string]string{
				tenantID: server.URL,
			}),
		)

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
			{
				PublicID: "publicID2",
			},
		}

		cache.EXPECT().
			Get(gomock.Any(), tenantID, leaderboardID, members).
			Return(nil, false, nil)

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(HaveOccurred())
		Expect(res).To(BeNil())
	})

	It("should fail if webhook returns invalid json", func() {
		mux.HandleFunc(enrichWebhookEndpoint, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("invalid"))
		})

		cache := mock_enriching.NewMockEnricherCache(gomock.NewController(GinkgoT()))
		enrich := NewEnricher(zap.NewNop(), cache,
			WithWebhookUrls(map[string]string{
				tenantID: server.URL,
			}),
		)

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
		}

		cache.EXPECT().
			Get(gomock.Any(), tenantID, leaderboardID, members).
			Return(nil, false, nil)

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ErrEnrichmentCall))
		Expect(res).To(BeNil())
	})

	It("should return members with metadata if call succeeds and update cache", func() {
		mux.HandleFunc(enrichWebhookEndpoint, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("{\"members\": [{ \"id\": \"publicID\", \"metadata\": { \"key\": \"value\" } }]}"))
		})

		cache := mock_enriching.NewMockEnricherCache(gomock.NewController(GinkgoT()))
		enrich := NewEnricher(zap.NewNop(), cache,
			WithWebhookUrls(map[string]string{
				tenantID: server.URL,
			}),
		)

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
		}

		cache.EXPECT().
			Get(gomock.Any(), tenantID, leaderboardID, members).
			Return(nil, false, nil)

		expectedResult := []*model.Member{
			{
				PublicID: "publicID",
				Metadata: map[string]string{
					"key": "value",
				},
			},
		}

		cache.EXPECT().
			Set(gomock.Any(), tenantID, leaderboardID, expectedResult, gomock.Any()).
			Return(nil)

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal(expectedResult))
	})

	It("should return cached data if all members are cached", func() {
		cache := mock_enriching.NewMockEnricherCache(gomock.NewController(GinkgoT()))
		enrich := NewEnricher(zap.NewNop(), cache)

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
		}

		cache.EXPECT().
			Get(gomock.Any(), tenantID, leaderboardID, members).
			Return(
				map[string]map[string]string{
					"publicID": {
						"key": "value",
					},
				},
				true, nil)

		expectedResult := []*model.Member{
			{
				PublicID: "publicID",
				Metadata: map[string]string{
					"key": "value",
				},
			},
		}

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal(expectedResult))
	})

	It("should ignore cache if one or more members are not cached", func() {
		mux.HandleFunc(enrichWebhookEndpoint, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("{\"members\": [{ \"id\": \"publicID\", \"metadata\": { \"key\": \"value\" } }, { \"id\": \"publicID2\", \"metadata\": { \"key2\": \"value2\" } }]}"))
		})

		cache := mock_enriching.NewMockEnricherCache(gomock.NewController(GinkgoT()))
		enrich := NewEnricher(zap.NewNop(), cache,
			WithWebhookUrls(map[string]string{
				tenantID: server.URL,
			}),
		)

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
			{
				PublicID: "publicID2",
			},
		}

		cache.EXPECT().
			Get(gomock.Any(), tenantID, leaderboardID, members).
			Return(map[string]map[string]string{
				"publicID": {
					"key": "value",
				},
				"publicID2": {
					"key2": "value2",
				},
			}, true, nil)

		expectedResult := []*model.Member{
			{
				PublicID: "publicID",
				Metadata: map[string]string{
					"key": "value",
				},
			},
			{
				PublicID: "publicID2",
				Metadata: map[string]string{
					"key2": "value2",
				},
			},
		}

		cache.EXPECT().
			Set(gomock.Any(), tenantID, leaderboardID, expectedResult, gomock.Any()).
			Return(nil)

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal(expectedResult))
	})
})

var _ = Describe("test url builder", func() {
	correctUrl := "http://localhost:8080" + enrichWebhookEndpoint
	It("should work if base url has no http", func() {
		baseUrl := "localhost:8080"
		res, err := buildUrl(baseUrl, enrichWebhookEndpoint)

		Expect(res).To(Equal(correctUrl))
		Expect(err).NotTo(HaveOccurred())
	})

	It("should work if base url ends with slash", func() {
		baseUrl := "http://localhost:8080/"
		res, err := buildUrl(baseUrl, enrichWebhookEndpoint)

		Expect(res).To(Equal(correctUrl))
		Expect(err).NotTo(HaveOccurred())
	})

	It("should work if base url does not end with slash", func() {
		baseUrl := "http://localhost:8080"
		res, err := buildUrl(baseUrl, enrichWebhookEndpoint)

		Expect(res).To(Equal(correctUrl))
		Expect(err).NotTo(HaveOccurred())
	})
})
