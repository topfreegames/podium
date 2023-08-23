//go:build unit

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

	It("should return correct error if tenantID is not configured", func() {
		enrich := &enricherImpl{
			config: EnrichmentConfig{},
			logger: zap.NewNop(),
		}

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
			{
				PublicID: "publicID2",
			},
		}

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)
		Expect(err).To(MatchError(ErrNotConfigured))
		Expect(res).To(BeNil())
	})

	It("should return error if webhook call fails", func() {
		ctrl := gomock.NewController(GinkgoT())
		cache := mock_enriching.NewMockEnricherCache(ctrl)
		mux.HandleFunc(enrichURL, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("{}"))
		})

		enrich := &enricherImpl{
			config: EnrichmentConfig{
				WebhookUrls: map[string]string{
					tenantID: server.URL,
				},
			},
			cache:  cache,
			logger: zap.NewNop(),
			client: &http.Client{},
		}

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
			{
				PublicID: "publicID2",
			},
		}

		cache.EXPECT().Get(gomock.Any(), tenantID, leaderboardID, members).Return(nil, false, nil)

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(HaveOccurred())
		Expect(res).To(BeNil())
	})

	It("should fail if webhook returns invalid json", func() {
		ctrl := gomock.NewController(GinkgoT())
		cache := mock_enriching.NewMockEnricherCache(ctrl)
		mux.HandleFunc(enrichURL, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("invalid"))
		})
		enrich := &enricherImpl{
			config: EnrichmentConfig{
				WebhookUrls: map[string]string{
					tenantID: server.URL,
				},
			},
			cache:  cache,
			logger: zap.NewNop(),
			client: &http.Client{},
		}

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
		}

		cache.EXPECT().Get(gomock.Any(), tenantID, leaderboardID, members).Return(nil, false, nil)

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ErrEnrichmentCall))
		Expect(res).To(BeNil())
	})

	It("should return members with metadata if call succeeds", func() {
		ctrl := gomock.NewController(GinkgoT())
		cache := mock_enriching.NewMockEnricherCache(ctrl)

		mux.HandleFunc(enrichURL, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("{\"members\": [{ \"id\": \"publicID\", \"metadata\": { \"key\": \"value\" } }]}"))
		})
		enrich := &enricherImpl{
			config: EnrichmentConfig{
				WebhookUrls: map[string]string{
					tenantID: server.URL,
				},
			},
			cache:  cache,
			logger: zap.NewNop(),
			client: &http.Client{},
		}

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
		}

		expectedResult := []*model.Member{
			{
				PublicID: "publicID",
				Metadata: map[string]string{
					"key": "value",
				},
			},
		}

		cache.EXPECT().Get(gomock.Any(), tenantID, leaderboardID, members).Return(nil, false, nil)
		cache.EXPECT().Set(gomock.Any(), tenantID, leaderboardID, expectedResult).Return(nil)

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal(expectedResult))
	})
})

var _ = Describe("test url builder", func() {
	correctUrl := "http://localhost:8080" + enrichURL
	It("should work if base url has no http", func() {
		baseUrl := "localhost:8080"
		res, err := buildUrl(baseUrl)

		Expect(res).To(Equal(correctUrl))
		Expect(err).NotTo(HaveOccurred())
	})

	It("should work if base url ends with slash", func() {
		baseUrl := "http://localhost:8080/"
		res, err := buildUrl(baseUrl)

		Expect(res).To(Equal(correctUrl))
		Expect(err).NotTo(HaveOccurred())
	})

	It("should work if base url does not end with slash", func() {
		baseUrl := "http://localhost:8080"
		res, err := buildUrl(baseUrl)

		Expect(res).To(Equal(correctUrl))
		Expect(err).NotTo(HaveOccurred())
	})
})
