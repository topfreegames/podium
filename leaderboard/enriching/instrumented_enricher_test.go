package enriching

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	extmocks "github.com/topfreegames/extensions/middleware/mocks"
	mock_enriching "github.com/topfreegames/podium/leaderboard/v2/mocks"
	"github.com/topfreegames/podium/leaderboard/v2/model"
)

var _ = Describe("Instrumented enricher", func() {
	var ctrl *gomock.Controller
	ctx := context.Background()
	tenantID := "tenantID"
	leaderboardID := "leaderboardID"
	members := []*model.Member{
		{
			PublicID: "publicID",
		},
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should send metrics if tenant ID is configured", func() {
		mockReporter := extmocks.NewMockMetricsReporter(ctrl)
		impl := mock_enriching.NewMockEnricher(ctrl)

		enricher := NewInstrumentedEnricher(impl, mockReporter)

		impl.EXPECT().Enrich(gomock.Any(), tenantID, leaderboardID, members).Return(members, nil)
		mockReporter.EXPECT().Increment(enrichmentCalls)
		mockReporter.EXPECT().Timing(enrichmentTimingMilli, gomock.Any())

		_, _ = enricher.Enrich(ctx, tenantID, leaderboardID, members)

	})

	It("should send error metric if enrichment returns error", func() {
		mockReporter := extmocks.NewMockMetricsReporter(ctrl)
		impl := mock_enriching.NewMockEnricher(ctrl)

		enricher := NewInstrumentedEnricher(impl, mockReporter)

		impl.EXPECT().Enrich(gomock.Any(), tenantID, leaderboardID, members).Return(nil, errors.New("error"))
		mockReporter.EXPECT().Increment(enrichmentCalls)
		mockReporter.EXPECT().Increment(enrichmentErrors)
		mockReporter.EXPECT().Timing(enrichmentTimingMilli, gomock.Any())

		_, _ = enricher.Enrich(ctx, tenantID, leaderboardID, members)
	})
})
