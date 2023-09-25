package cache

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_extensions "github.com/topfreegames/extensions/middleware/mocks"
	mock_enriching "github.com/topfreegames/podium/leaderboard/v2/mocks"
	"github.com/topfreegames/podium/leaderboard/v2/model"
)

var _ = Describe("Instrumented enrich cache Get tests", func() {
	tenantID := "tenant-id"
	leaderboardID := "leaderboard-id"
	members := []*model.Member{
		{
			PublicID: "member1",
		},
	}

	result := map[string]map[string]string{
		"member1": {
			"field1": "value1",
		},
	}

	It("should send metrics when Get is called successfully with hit", func() {
		ctrl := gomock.NewController(GinkgoT())
		impl := mock_enriching.NewMockEnricherCache(ctrl)
		metricsReporter := mock_extensions.NewMockMetricsReporter(ctrl)

		impl.EXPECT().Get(gomock.Any(), tenantID, leaderboardID, members).Return(result, true, nil)
		metricsReporter.EXPECT().Increment(enrichmentCacheGets).Return(nil)
		metricsReporter.EXPECT().Increment(enrichmentCacheHits).Return(nil)
		metricsReporter.EXPECT().Timing(enrichmentCacheGetTimingMilli, gomock.Any()).Return(nil)

		instrumentedCache := NewInstrumentedCache(impl, metricsReporter)
		res, hit, err := instrumentedCache.Get(context.Background(), tenantID, leaderboardID, members)

		Expect(res).To(Equal(result))
		Expect(hit).To(BeTrue())
		Expect(err).To(BeNil())
	})

	It("should send metrics when Get is called successfully with miss", func() {
		ctrl := gomock.NewController(GinkgoT())
		impl := mock_enriching.NewMockEnricherCache(ctrl)
		metricsReporter := mock_extensions.NewMockMetricsReporter(ctrl)

		impl.EXPECT().Get(gomock.Any(), tenantID, leaderboardID, members).Return(nil, false, nil)
		metricsReporter.EXPECT().Increment(enrichmentCacheGets).Return(nil)
		metricsReporter.EXPECT().Timing(enrichmentCacheGetTimingMilli, gomock.Any()).Return(nil)

		instrumentedCache := NewInstrumentedCache(impl, metricsReporter)
		res, hit, err := instrumentedCache.Get(context.Background(), tenantID, leaderboardID, members)
		Expect(res).To(BeNil())
		Expect(hit).To(BeFalse())
		Expect(err).To(BeNil())
	})

	It("should send metrics when Get is called with error", func() {
		ctrl := gomock.NewController(GinkgoT())
		impl := mock_enriching.NewMockEnricherCache(ctrl)
		metricsReporter := mock_extensions.NewMockMetricsReporter(ctrl)

		impl.EXPECT().Get(gomock.Any(), tenantID, leaderboardID, members).Return(nil, false, errors.New("error"))
		metricsReporter.EXPECT().Increment(enrichmentCacheGets).Return(nil)
		metricsReporter.EXPECT().Increment(enrichmentCacheGetErrors).Return(nil)
		metricsReporter.EXPECT().Timing(enrichmentCacheGetTimingMilli, gomock.Any()).Return(nil)

		instrumentedCache := NewInstrumentedCache(impl, metricsReporter)
		res, hit, err := instrumentedCache.Get(context.Background(), tenantID, leaderboardID, members)

		Expect(res).To(BeNil())
		Expect(hit).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("Instrumented enrich cache Set tests", func() {
	tenantID := "tenant-id"
	leaderboardID := "leaderboard-id"
	members := []*model.Member{
		{
			PublicID: "member1",
			Metadata: map[string]string{
				"field1": "value1",
			},
		},
	}

	It("should send metrics when Set is called successfully", func() {
		ctrl := gomock.NewController(GinkgoT())
		impl := mock_enriching.NewMockEnricherCache(ctrl)
		metricsReporter := mock_extensions.NewMockMetricsReporter(ctrl)

		impl.EXPECT().Set(gomock.Any(), tenantID, leaderboardID, members, gomock.Any()).Return(nil)
		metricsReporter.EXPECT().Increment(enrichmentCacheSets).Return(nil)
		metricsReporter.EXPECT().Timing(enrichmentCacheSetTimingMilli, gomock.Any()).Return(nil)

		instrumentedCache := NewInstrumentedCache(impl, metricsReporter)
		err := instrumentedCache.Set(context.Background(), tenantID, leaderboardID, members, 0)

		Expect(err).To(BeNil())
	})

	It("should send metrics when Set is called with error", func() {
		ctrl := gomock.NewController(GinkgoT())
		impl := mock_enriching.NewMockEnricherCache(ctrl)
		metricsReporter := mock_extensions.NewMockMetricsReporter(ctrl)

		impl.EXPECT().Set(gomock.Any(), tenantID, leaderboardID, members, gomock.Any()).Return(errors.New("error"))
		metricsReporter.EXPECT().Increment(enrichmentCacheSets).Return(nil)
		metricsReporter.EXPECT().Increment(enrichmentCacheSetErrors).Return(nil)
		metricsReporter.EXPECT().Timing(enrichmentCacheSetTimingMilli, gomock.Any()).Return(nil)

		instrumentedCache := NewInstrumentedCache(impl, metricsReporter)
		err := instrumentedCache.Set(context.Background(), tenantID, leaderboardID, members, 0)

		Expect(err).To(HaveOccurred())
	})
})
