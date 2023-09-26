package cache

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_enriching "github.com/topfreegames/podium/leaderboard/v2/mocks"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"time"
)

var _ = Describe("Enricher with cache tests", func() {
	tenantID := "tenantID"
	leaderboardID := "leaderboardID"
	It("should update cache on success", func() {

		ctrl := gomock.NewController(GinkgoT())
		enricher := mock_enriching.NewMockEnricher(ctrl)
		cache := mock_enriching.NewMockEnricherCache(ctrl)

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
		cache.EXPECT().
			Get(gomock.Any(), tenantID, leaderboardID, members).
			Return(nil, false, nil)

		enricher.EXPECT().
			Enrich(gomock.Any(), tenantID, leaderboardID, members).
			Return(expectedResult, nil)

		cache.EXPECT().
			Set(gomock.Any(), tenantID, leaderboardID, expectedResult, 24*time.Hour).
			Return(nil)

		wrapper := NewCachedEnricher(cache, enricher)
		res, err := wrapper.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal(expectedResult))
	})

	It("should not fail on cache set error", func() {
		ctrl := gomock.NewController(GinkgoT())
		enricher := mock_enriching.NewMockEnricher(ctrl)
		cache := mock_enriching.NewMockEnricherCache(ctrl)

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

		cache.EXPECT().
			Get(gomock.Any(), tenantID, leaderboardID, members).
			Return(nil, false, nil)

		cache.EXPECT().
			Set(gomock.Any(), tenantID, leaderboardID, expectedResult, 24*time.Hour).
			Return(errors.New("error"))

		enricher.EXPECT().
			Enrich(gomock.Any(), tenantID, leaderboardID, members).
			Return(expectedResult, nil)

		wrapper := NewCachedEnricher(cache, enricher)
		res, err := wrapper.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal(expectedResult))
	})

	It("should return cached data if all members are cached", func() {
		ctrl := gomock.NewController(GinkgoT())
		enricher := mock_enriching.NewMockEnricher(ctrl)
		cache := mock_enriching.NewMockEnricherCache(ctrl)

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

		cache.EXPECT().
			Get(gomock.Any(), tenantID, leaderboardID, members).
			Return(map[string]map[string]string{
				"publicID": {
					"key": "value",
				},
			}, true, nil)

		wrapped := NewCachedEnricher(cache, enricher)
		res, err := wrapped.Enrich(context.Background(), tenantID, leaderboardID, members)

		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal(expectedResult))
	})
})
