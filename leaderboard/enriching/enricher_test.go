package enriching

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/v2/model"
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

	It("should not enrich if no webhook url is configured and cloud save service is not enabled", func() {

		enrich := NewEnricher()

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
			{
				PublicID: "publicID2",
			},
		}

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(BeNil())
		Expect(res).To(Equal(members))
	})

	It("should fail if cloud save service call fails (status != 200)", func() {
		mux.HandleFunc(cloudSaveEndpoint, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("{}"))
		})

		enrich := NewEnricher(
			WithCloudSaveUrl(server.URL),
			WithCloudSaveEnabled(map[string]bool{
				tenantID: true,
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

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(HaveOccurred())
		Expect(res).To(BeNil())
	})

	It("should fail if cloud save service returns invalid json", func() {
		mux.HandleFunc(cloudSaveEndpoint, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("invalid"))
		})

		enrich := NewEnricher(
			WithCloudSaveUrl(server.URL),
			WithCloudSaveEnabled(map[string]bool{
				tenantID: true,
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

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(HaveOccurred())
		Expect(res).To(BeNil())
	})

	It("should succeed with enrichment from cloud save service", func() {
		mux.HandleFunc(cloudSaveEndpoint, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("{\"documents\": [{\"accountId\": \"publicID\", \"data\": {\"key\": \"value\"}}]}"))
		})

		enrich := NewEnricher(
			WithCloudSaveUrl(server.URL),
			WithCloudSaveEnabled(map[string]bool{
				tenantID: true,
			}),
		)

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

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(BeNil())
		Expect(res).To(Equal(expectedResult))
	})

	It("should return error if webhook call fails", func() {
		mux.HandleFunc(enrichWebhookEndpoint, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("{}"))
		})

		enrich := NewEnricher(
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

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(HaveOccurred())
		Expect(res).To(BeNil())
	})

	It("should fail if webhook returns invalid json", func() {
		mux.HandleFunc(enrichWebhookEndpoint, func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("invalid"))
		})

		enrich := NewEnricher(
			WithWebhookUrls(map[string]string{
				tenantID: server.URL,
			}),
		)

		members := []*model.Member{
			{
				PublicID: "publicID",
			},
		}

		res, err := enrich.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ErrEnrichmentCall))
		Expect(res).To(BeNil())
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
