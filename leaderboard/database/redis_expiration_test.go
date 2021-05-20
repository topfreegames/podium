package database_test

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/database/redis"
)

var _ = Describe("Redis Expiration Database", func() {
	var ctrl *gomock.Controller
	var mock *redis.MockRedis
	var redisExpiration database.Expiration
	var leaderboard string = "leaderboardTest"
	var leaderboardTTL string = "leaderboardTest:ttl"
	var amount int = 10
	var member string = "memberTest"

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = redis.NewMockRedis(ctrl)

		redisExpiration = &database.Redis{mock}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetExpirationLeaderboards", func() {
		It("Should return leaderboard expiration time if all is OK", func() {
			mock.EXPECT().SMembers(gomock.Any(), gomock.Eq(database.ExpirationSet)).Return([]string{leaderboardTTL}, nil)

			leaderboards, err := redisExpiration.GetExpirationLeaderboards(context.Background())
			Expect(err).NotTo(HaveOccurred())

			Expect(leaderboards).To(Equal([]string{leaderboard}))
		})

		It("Should return GeneralError if redis return any other error", func() {
			mock.EXPECT().SMembers(gomock.Any(), gomock.Eq(database.ExpirationSet)).Return(nil, fmt.Errorf("redis error"))

			_, err := redisExpiration.GetExpirationLeaderboards(context.Background())
			Expect(err).To(Equal(database.NewGeneralError("redis error")))

		})
	})

	Describe("GetMembersToExpire", func() {
		It("Should return members to expiration on leaderboard", func() {
			maxTTL := time.Now()
			maxTTLString := strconv.FormatInt(maxTTL.UTC().Unix(), 10)
			mock.EXPECT().Exists(gomock.Any(), gomock.Eq(leaderboardTTL)).Return(nil)
			mock.EXPECT().ZRangeByScore(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq("-inf"), gomock.Eq(maxTTLString), gomock.Eq(int64(0)), gomock.Eq(int64(amount))).Return([]string{member}, nil)

			members, err := redisExpiration.GetMembersToExpire(context.Background(), leaderboard, amount, maxTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(members).To(Equal([]string{member}))
		})

		It("Should return LeaderboardWithoutMemberToExpireError to expiration on leaderboard", func() {
			maxTTL := time.Now()
			mock.EXPECT().Exists(gomock.Any(), gomock.Eq(leaderboardTTL)).Return(redis.NewKeyNotFoundError(leaderboardTTL))

			_, err := redisExpiration.GetMembersToExpire(context.Background(), leaderboard, amount, maxTTL)
			Expect(err).To(MatchError(database.NewLeaderboardWithoutMemberToExpireError(leaderboard)))
		})

		It("Should return members to expiration on leaderboard", func() {
			maxTTL := time.Now()
			maxTTLString := strconv.FormatInt(maxTTL.UTC().Unix(), 10)
			mock.EXPECT().Exists(gomock.Any(), gomock.Eq(leaderboardTTL)).Return(nil)
			mock.EXPECT().ZRangeByScore(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq("-inf"), gomock.Eq(maxTTLString), gomock.Eq(int64(0)), gomock.Eq(int64(amount))).Return(nil, fmt.Errorf("New redis error"))

			_, err := redisExpiration.GetMembersToExpire(context.Background(), leaderboard, amount, maxTTL)
			Expect(err).To(MatchError(database.NewGeneralError("New redis error")))
		})
	})

	Describe("RemoveLeaderboardFromExpireList", func() {
		It("Should return nil if all is right", func() {
			mock.EXPECT().SRem(gomock.Any(), gomock.Eq(database.ExpirationSet), gomock.Eq(leaderboardTTL)).Return(nil)
			err := redisExpiration.RemoveLeaderboardFromExpireList(context.Background(), leaderboard)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should return GeneralError if redis return in error", func() {
			mock.EXPECT().SRem(gomock.Any(), gomock.Eq(database.ExpirationSet), gomock.Eq(leaderboardTTL)).Return(fmt.Errorf("New redis error"))
			err := redisExpiration.RemoveLeaderboardFromExpireList(context.Background(), leaderboard)
			Expect(err).To(MatchError(database.NewGeneralError("New redis error")))
		})
	})
	Describe("ExpireMembers", func() {
		It("Should return nil if all is ok", func() {
			member2 := "member2"

			mock.EXPECT().ZRem(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(member2)).Return(nil)
			mock.EXPECT().ZRem(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq(member), gomock.Eq(member2)).Return(nil)

			err := redisExpiration.ExpireMembers(context.Background(), leaderboard, []string{member, member2})
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should return GeneralError if redis return in error on remove member from leaderboard", func() {
			member2 := "member2"

			mock.EXPECT().ZRem(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(member2)).Return(fmt.Errorf("New redis error"))

			err := redisExpiration.ExpireMembers(context.Background(), leaderboard, []string{member, member2})
			Expect(err).To(MatchError(database.NewGeneralError("New redis error")))
		})

		It("Should return GeneralError if redis return in error on remove member from expiration set", func() {
			member2 := "member2"

			mock.EXPECT().ZRem(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(member2)).Return(nil)
			mock.EXPECT().ZRem(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq(member), gomock.Eq(member2)).Return(fmt.Errorf("New redis error"))

			err := redisExpiration.ExpireMembers(context.Background(), leaderboard, []string{member, member2})
			Expect(err).To(MatchError(database.NewGeneralError("New redis error")))
		})
	})
})
