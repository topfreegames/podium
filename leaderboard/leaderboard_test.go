// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package leaderboard_test

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	extredis "github.com/topfreegames/extensions/redis"
	"github.com/topfreegames/extensions/redis/interfaces"

	"github.com/satori/go.uuid"
	. "github.com/topfreegames/podium/leaderboard"
	"github.com/topfreegames/podium/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getFaultyRedis() interfaces.RedisClient {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:38465",
		Password: "",
		DB:       0,
		PoolSize: 20,
	})
}

var _ = Describe("Leaderboard Model", func() {

	var redisClient *extredis.Client
	var faultyRedisClient *extredis.Client
	var logger *testing.MockLogger

	BeforeEach(func() {
		var err error

		config := viper.New()
		config.Set("redis.url", "redis://localhost:1234/0")
		config.Set("redis.connectionTimeout", 200)

		logger = testing.NewMockLogger()
		redisClient, err = extredis.NewClient("redis", config)
		Expect(err).NotTo(HaveOccurred())

		//First we connect properly
		faultyRedisClient, err = extredis.NewClient("redis", config)
		Expect(err).NotTo(HaveOccurred())
		//Then we change the connection to be faulty
		faultyRedisClient.Client = redis.NewClient(&redis.Options{
			Addr:     "localhost:1235",
			Password: "",
			DB:       0,
		})

		_, err = redisClient.Client.Del("test-leaderboard").Result()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		_, err := redisClient.Client.Del("test-leaderboard").Result()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("setting member scores", func() {
		It("should set scores and return ranks", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 10, logger)
			dayvson, err := testLeaderboard.SetMemberScore("dayvson", 481516, false, "")
			Expect(err).NotTo(HaveOccurred())
			arthur, err := testLeaderboard.SetMemberScore("arthur", 1000, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(dayvson.Rank).To(Equal(1))
			Expect(arthur.Rank).To(Equal(2))
		})

		It("should set score expiration if expiry field is passed", func() {
			ttl := "100"
			lbName := "test-leaderboard"
			testLeaderboard := NewLeaderboard(redisClient.Client, lbName, 10, logger)
			_, err := testLeaderboard.SetMemberScore("denix", 481516, false, ttl)
			Expect(err).NotTo(HaveOccurred())
			redisLBExpirationKey := fmt.Sprintf("%s:ttl:%s", lbName, ttl)
			result, err := redisClient.Client.Exists(redisLBExpirationKey).Result()
			Expect(err).NotTo(HaveOccurred())
			redisExpirationSetKey := "expiration-sets"
			result, err = redisClient.Client.Exists(redisExpirationSetKey).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(1)))
			result2, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result2).To(ContainElement(redisLBExpirationKey))
			result3, err := redisClient.Client.ZRangeWithScores(redisLBExpirationKey, 0, 1).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result3[0].Member).To(Equal("denix"))
			Expect(result3[0].Score).To(BeNumerically("<=", time.Now().Unix()))
		})

		It("should set scores and return previous ranks", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 10, logger)
			member1, err := testLeaderboard.SetMemberScore("member1", 481516, true, "")
			Expect(err).NotTo(HaveOccurred())
			member2, err := testLeaderboard.SetMemberScore("member2", 1000, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(member1.Rank).To(Equal(1))
			Expect(member1.PreviousRank).To(Equal(-1))
			Expect(member2.Rank).To(Equal(2))
			Expect(member2.PreviousRank).To(Equal(0))
			nmember1, err := testLeaderboard.SetMemberScore("member1", 1, true, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(nmember1.Rank).To(Equal(2))
			Expect(nmember1.PreviousRank).To(Equal(1))
		})

		It("should fail if invalid connection to Redis", func() {
			testLeaderboard := NewLeaderboard(getFaultyRedis(), "test-leaderboard", 10, logger)
			_, err := testLeaderboard.SetMemberScore("dayvson", 481516, false, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("increment member scores", func() {
		It("should increment member score and return ranks", func() {
			lbID := uuid.NewV4().String()
			testLeaderboard := NewLeaderboard(redisClient.Client, lbID, 10, logger)

			_, err := testLeaderboard.SetMemberScore("dayvson", 1000, false, "")
			Expect(err).NotTo(HaveOccurred())

			member, err := testLeaderboard.IncrementMemberScore("dayvson", 10, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Score).To(Equal(1010))
			Expect(member.PublicID).To(Equal("dayvson"))

			score, err := redisClient.Client.ZScore(lbID, "dayvson").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(int(score)).To(Equal(1010))
		})

		It("should increment member score when leaderboard does not exist and return ranks", func() {
			lbID := uuid.NewV4().String()
			testLeaderboard := NewLeaderboard(redisClient.Client, lbID, 10, logger)

			member, err := testLeaderboard.IncrementMemberScore("dayvson", 10, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Score).To(Equal(10))
			Expect(member.PublicID).To(Equal("dayvson"))

			score, err := redisClient.Client.ZScore(lbID, "dayvson").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(int(score)).To(Equal(10))
		})

		It("should fail if invalid connection to Redis", func() {
			testLeaderboard := NewLeaderboard(getFaultyRedis(), "test-leaderboard", 10, logger)
			_, err := testLeaderboard.IncrementMemberScore("dayvson", 16, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("getting number of members", func() {
		It("should retrieve the number of members in a leaderboard", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 10, logger)
			for i := 0; i < 10; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			count, err := testLeaderboard.TotalMembers()
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(10))
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 10, logger)
			testLeaderboard.RedisClient = faultyRedisClient.Client
			_, err := testLeaderboard.TotalMembers()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("removing members", func() {
		It("should remove member", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 10, logger)
			for i := 0; i < 10; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(testLeaderboard.TotalMembers()).To(Equal(10))
			member := "member_5"
			testLeaderboard.RemoveMember(member)
			Expect(testLeaderboard.TotalMembers()).To(Equal(9))
		})

		It("should remove many members", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 10, logger)
			for i := 0; i < 10; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(testLeaderboard.TotalMembers()).To(Equal(10))
			members := make([]interface{}, 2)
			members[0] = "member_5"
			members[1] = "member_6"
			testLeaderboard.RemoveMembers(members)
			Expect(testLeaderboard.TotalMembers()).To(Equal(8))
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(faultyRedisClient.Client, "test-leaderboard", 10, logger)
			members := make([]interface{}, 1)
			members[0] = "invalid member"
			err := testLeaderboard.RemoveMembers(members)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("getting number of pages in leaderboard", func() {
		It("should return total number of pages", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(testLeaderboard.TotalPages()).To(Equal(5))
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 10, logger)
			testLeaderboard.RedisClient = faultyRedisClient.Client
			_, err := testLeaderboard.TotalPages()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("getting member details for a given leaderboard", func() {
		It("should return member details", func() {
			friendScore := NewLeaderboard(redisClient.Client, uuid.NewV4().String(), 10, logger)
			dayvson, err := friendScore.SetMemberScore("dayvson", 12345, false, "")
			Expect(err).NotTo(HaveOccurred())
			felipe, err := friendScore.SetMemberScore("felipe", 12344, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(dayvson.Rank).To(Equal(1))
			Expect(felipe.Rank).To(Equal(2))
			friendScore.SetMemberScore("felipe", 12346, false, "")
			felipe, err = friendScore.GetMember("felipe", "desc")
			Expect(err).NotTo(HaveOccurred())
			dayvson, err = friendScore.GetMember("dayvson", "desc")
			Expect(err).NotTo(HaveOccurred())
			Expect(felipe.Rank).To(Equal(1))
			Expect(dayvson.Rank).To(Equal(2))
		})

		It("should fail if member does not exist", func() {
			leaderboardID := uuid.NewV4().String()
			friendScore := NewLeaderboard(redisClient.Client, leaderboardID, 10, logger)
			memberID := uuid.NewV4().String()
			member, err := friendScore.GetMember(memberID, "desc")
			Expect(member).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				fmt.Sprintf("Could not find data for member %s in leaderboard %s.", memberID, leaderboardID)),
			)
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 10, logger)
			testLeaderboard.RedisClient = faultyRedisClient.Client
			_, err := testLeaderboard.GetMember("qwe", "desc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("get members around someone in a leaderboard", func() {
		It("should get members around specific member", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetAroundMe("member_20", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[testLeaderboard.PageSize-1]
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			Expect(firstAroundMe.PublicID).To(Equal("member_31"))
			Expect(lastAroundMe.PublicID).To(Equal("member_7"))
		})

		It("should get members around specific member in reverse order", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 20, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetAroundMe("member_20", "asc", false)
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[testLeaderboard.PageSize-1]
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			Expect(firstAroundMe.PublicID).To(Equal("member_11"))
			Expect(lastAroundMe.PublicID).To(Equal("member_30"))
		})

		It("should get members around specific member if repeated scores", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetAroundMe("member_20", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			firstAroundMe := members[0]
			lastAroundMe := members[testLeaderboard.PageSize-1]
			Expect(firstAroundMe.Score).To(Equal(100))
			Expect(lastAroundMe.Score).To(Equal(100))
		})

		It("should get PageSize members around specific member even if member in ranking top", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 1; i <= 100; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 100-i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetAroundMe("member_2", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			firstAroundMe := members[0]
			lastAroundMe := members[testLeaderboard.PageSize-1]
			Expect(firstAroundMe.PublicID).To(Equal("member_1"))
			Expect(lastAroundMe.PublicID).To(Equal("member_25"))
		})

		It("should get PageSize members around specific member even if member in ranking bottom", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 1; i <= 100; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 100-i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetAroundMe("member_99", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			firstAroundMe := members[0]
			lastAroundMe := members[testLeaderboard.PageSize-1]
			Expect(firstAroundMe.PublicID).To(Equal("member_76"))
			Expect(lastAroundMe.PublicID).To(Equal("member_100"))
		})

		It("should get PageSize members when interval larger than total members", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 1; i <= 10; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 100-i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetAroundMe("member_2", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(10))
			firstAroundMe := members[0]
			lastAroundMe := members[9]
			Expect(firstAroundMe.PublicID).To(Equal("member_1"))
			Expect(lastAroundMe.PublicID).To(Equal("member_10"))
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(getFaultyRedis(), "test-leaderboard", 10, logger)
			_, err := testLeaderboard.GetAroundMe("qwe", "desc", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("getting member ranking", func() {
		It("should return specific member ranking", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			testLeaderboard.SetMemberScore("member_6", 1000, false, "")
			Expect(testLeaderboard.GetRank("member_6", "desc")).To(Equal(100))
		})

		It("should return specific member ranking if asc order", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			testLeaderboard.SetMemberScore("member_6", 1000, false, "")
			Expect(testLeaderboard.GetRank("member_6", "asc")).To(Equal(2))
		})

		It("should fail if member does not exist", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, uuid.NewV4().String(), 25, logger)
			rank, err := testLeaderboard.GetRank("invalid-member", "desc")
			Expect(rank).To(Equal(-1))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member invalid-member in leaderboard"))
		})

		It("should fail if invalid redis connection", func() {
			testLeaderboard := NewLeaderboard(getFaultyRedis(), uuid.NewV4().String(), 25, logger)
			rank, err := testLeaderboard.GetRank("invalid-member", "desc")
			Expect(rank).To(Equal(-1))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("getting leaderboard leaders", func() {
		It("should get specific number of leaders", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 0; i < 1000; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i+1), 1234*i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetLeaders(1, "desc")
			Expect(err).NotTo(HaveOccurred())

			firstOnPage := members[0]
			lastOnPage := members[len(members)-1]
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			Expect(firstOnPage.PublicID).To(Equal("member_1000"))
			Expect(firstOnPage.Rank).To(Equal(1))
			Expect(lastOnPage.PublicID).To(Equal("member_976"))
			Expect(lastOnPage.Rank).To(Equal(25))
		})

		It("should get specific number of leaders in reverse order", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 0; i < 1000; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i+1), 1234*i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetLeaders(1, "asc")
			Expect(err).NotTo(HaveOccurred())

			firstOnPage := members[0]
			lastOnPage := members[len(members)-1]
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			Expect(firstOnPage.PublicID).To(Equal("member_1"))
			Expect(firstOnPage.Rank).To(Equal(1))
			Expect(lastOnPage.PublicID).To(Equal("member_25"))
			Expect(lastOnPage.Rank).To(Equal(25))
		})

		It("should get leaders if repeated scores", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetLeaders(1, "desc")
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[testLeaderboard.PageSize-1]
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			Expect(firstAroundMe.Score).To(Equal(100))
			Expect(lastAroundMe.Score).To(Equal(100))
		})

		It("should get leaders for negative pages get page 1", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetLeaders(-1, "desc")
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[testLeaderboard.PageSize-1]
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			Expect(firstAroundMe.Score).To(Equal(100))
			Expect(lastAroundMe.Score).To(Equal(100))
		})

		It("should get empty leaders for pages greater than total pages", func() {
			testLeaderboard := NewLeaderboard(redisClient.Client, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetLeaders(99999, "desc")
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(0))
		})

		It("should fail if invalid connection to Redis", func() {
			testLeaderboard := NewLeaderboard(getFaultyRedis(), "test-leaderboard", 25, logger)
			_, err := testLeaderboard.GetLeaders(1, "desc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("expiration of leaderboards", func() {
		It("should fail if invalid leaderboard", func() {
			leaderboardID := "leaderboard_from20201039to20201011"
			friendScore := NewLeaderboard(redisClient.Client, leaderboardID, 10, logger)
			_, err := friendScore.SetMemberScore("dayvson", 12345, false, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("day out of range"))
		})

		It("should add yearly expiration if leaderboard supports it", func() {
			leaderboardID := fmt.Sprintf("test-leaderboard-year%d", time.Now().UTC().Year())
			friendScore := NewLeaderboard(redisClient.Client, leaderboardID, 10, logger)
			_, err := friendScore.SetMemberScore("dayvson", 12345, false, "")
			Expect(err).NotTo(HaveOccurred())

			result, err := redisClient.Client.TTL(leaderboardID).Result()
			Expect(err).NotTo(HaveOccurred())

			exp := result.Seconds()
			Expect(err).NotTo(HaveOccurred())
			Expect(exp).To(BeNumerically(">", int64(-1)))
		})
	})

	Describe("get top x percent of members in the leaderboard", func() {
		It("should get top 10 percent members in the leaderboard", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient.Client, leaderboardID, 10, logger)

			members := []*Member{}
			for i := 0; i < 100; i++ {
				member, err := leader.SetMemberScore(fmt.Sprintf("friend-%d", i), (100-i)*100, false, "")
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leader.GetTopPercentage(10, 2000, "desc")
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(10))

			Expect(top10[0].PublicID).To(Equal("friend-0"))
			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(10000))

			Expect(top10[9].PublicID).To(Equal("friend-9"))
			Expect(top10[9].Rank).To(Equal(10))
			Expect(top10[9].Score).To(Equal(9100))
		})

		It("should not break if order is different from asc and desc, should only default to desc", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient.Client, leaderboardID, 10, logger)

			members := []*Member{}
			for i := 0; i < 100; i++ {
				member, err := leader.SetMemberScore(fmt.Sprintf("friend-%d", i), (100-i)*100, false, "")
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leader.GetTopPercentage(10, 2000, "lalala")
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(10))

			Expect(top10[0].PublicID).To(Equal("friend-0"))
			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(10000))

			Expect(top10[9].PublicID).To(Equal("friend-9"))
			Expect(top10[9].Rank).To(Equal(10))
			Expect(top10[9].Score).To(Equal(9100))
		})

		It("should get top 10 percent members in the leaderboard in reverse order", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient.Client, leaderboardID, 10, logger)

			members := []*Member{}
			for i := 0; i < 100; i++ {
				member, err := leader.SetMemberScore(fmt.Sprintf("friend-%d", i), (100-i)*100, false, "")
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leader.GetTopPercentage(10, 2000, "asc")
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(10))

			Expect(top10[0].PublicID).To(Equal("friend-99"))
			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(100))

			Expect(top10[9].PublicID).To(Equal("friend-90"))
			Expect(top10[9].Rank).To(Equal(10))
			Expect(top10[9].Score).To(Equal(1000))
		})

		It("should get max members if query too broad", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient.Client, leaderboardID, 10, logger)

			members := []*Member{}
			for i := 0; i < 10; i++ {
				member, err := leader.SetMemberScore(fmt.Sprintf("friend-%d", i), (100-i)*100, false, "")
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top3, err := leader.GetTopPercentage(100, 3, "desc")
			Expect(err).NotTo(HaveOccurred())

			Expect(top3).To(HaveLen(3))

			Expect(top3[0].PublicID).To(Equal("friend-0"))
			Expect(top3[0].Rank).To(Equal(1))
			Expect(top3[0].Score).To(Equal(10000))

			Expect(top3[2].PublicID).To(Equal("friend-2"))
			Expect(top3[2].Rank).To(Equal(3))
			Expect(top3[2].Score).To(Equal(9800))
		})

		It("should get top 1 percent return at least 1", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient.Client, leaderboardID, 10, logger)

			members := []*Member{}
			for i := 0; i < 2; i++ {
				member, err := leader.SetMemberScore(fmt.Sprintf("friend-%d", i), (100-i)*100, false, "")
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leader.GetTopPercentage(1, 2000, "desc")
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(1))

			Expect(top10[0].PublicID).To(Equal("friend-0"))
			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(10000))
		})

		It("should get top 10 percent members in the leaderboard if repeated scores", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient.Client, leaderboardID, 10, logger)

			members := []*Member{}
			for i := 0; i < 100; i++ {
				member, err := leader.SetMemberScore(fmt.Sprintf("friend-%d", i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leader.GetTopPercentage(10, 2000, "desc")
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(10))

			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(100))

			Expect(top10[9].Rank).To(Equal(10))
			Expect(top10[9].Score).To(Equal(100))
		})

		It("should fail if more than 100 percent", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient.Client, leaderboardID, 10, logger)

			top10, err := leader.GetTopPercentage(101, 2000, "desc")
			Expect(top10).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Percentage must be a valid integer between 1 and 100."))
		})

		It("should fail if invalid redis connection", func() {
			testLeaderboard := NewLeaderboard(getFaultyRedis(), uuid.NewV4().String(), 25, logger)
			members, err := testLeaderboard.GetTopPercentage(10, 2000, "desc")
			Expect(members).To(BeEmpty())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("get members by range in a given leaderboard", func() {
		It("should get members in a range in the leaderboard", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient.Client, leaderboardID, 10, logger)

			expMembers := []*Member{}
			for i := 0; i < 100; i++ {
				member, err := leader.SetMemberScore(fmt.Sprintf("friend-%d", i), 100-i, false, "")
				Expect(err).NotTo(HaveOccurred())
				expMembers = append(expMembers, member)
			}

			members, err := GetMembersByRange(redisClient.Client, leaderboardID, 20, 39, "desc", logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(20))

			for i := 20; i < 40; i++ {
				Expect(members[i-20].PublicID).To(Equal(expMembers[i].PublicID))
			}
		})

		It("should fail if invalid connection to Redis", func() {
			leaderboardID := uuid.NewV4().String()
			_, err := GetMembersByRange(getFaultyRedis(), leaderboardID, 20, 39, "desc", logger)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("remove leaderboard", func() {
		It("should remove a leaderboard from redis", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient.Client, leaderboardID, 10, logger)

			for i := 0; i < 10; i++ {
				_, err := leader.SetMemberScore(fmt.Sprintf("friend-%d", i), 100-i, false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			err := leader.RemoveLeaderboard()
			Expect(err).NotTo(HaveOccurred())

			exists, err := redisClient.Client.Exists(leaderboardID).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(Equal(int64(0)))
		})

		It("should fail if invalid connection to Redis", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(getFaultyRedis(), leaderboardID, 10, logger)
			err := leader.RemoveLeaderboard()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("getting many members at once", func() {
		It("should return all member details", func() {
			lb := NewLeaderboard(redisClient.Client, uuid.NewV4().String(), 10, logger)
			for i := 0; i < 100; i++ {
				lb.SetMemberScore(fmt.Sprintf("member-%d", i), 100-i, false, "")
			}

			members, err := lb.GetMembers([]string{"member-10", "member-30", "member-20"}, "desc")
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(3))

			Expect(members[0].PublicID).To(Equal("member-10"))
			Expect(members[0].Rank).To(Equal(11))
			Expect(members[0].Score).To(Equal(90))

			Expect(members[1].PublicID).To(Equal("member-20"))
			Expect(members[1].Rank).To(Equal(21))
			Expect(members[1].Score).To(Equal(80))

			Expect(members[2].PublicID).To(Equal("member-30"))
			Expect(members[2].Rank).To(Equal(31))
			Expect(members[2].Score).To(Equal(70))
		})

		It("should return all member details using reverse rank", func() {
			lb := NewLeaderboard(redisClient.Client, uuid.NewV4().String(), 10, logger)
			for i := 0; i < 100; i++ {
				lb.SetMemberScore(fmt.Sprintf("member-%d", i), 100-i, false, "")
			}

			members, err := lb.GetMembers([]string{"member-10", "member-30", "member-20"}, "asc")
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(3))

			Expect(members[0].PublicID).To(Equal("member-30"))
			Expect(members[0].Rank).To(Equal(70))
			Expect(members[0].Score).To(Equal(70))

			Expect(members[1].PublicID).To(Equal("member-20"))
			Expect(members[1].Rank).To(Equal(80))
			Expect(members[1].Score).To(Equal(80))

			Expect(members[2].PublicID).To(Equal("member-10"))
			Expect(members[2].Rank).To(Equal(90))
			Expect(members[2].Score).To(Equal(90))
		})

		It("should return empty list if invalid leaderboard id", func() {
			lb := NewLeaderboard(redisClient.Client, uuid.NewV4().String(), 10, logger)
			members, err := lb.GetMembers([]string{"test"}, "desc")
			Expect(err).NotTo(HaveOccurred())

			Expect(members).To(HaveLen(0))
		})

		It("should return empty list if invalid members", func() {
			lb := NewLeaderboard(redisClient.Client, uuid.NewV4().String(), 10, logger)

			for i := 0; i < 10; i++ {
				lb.SetMemberScore(fmt.Sprintf("member-%d", i), 100-i, false, "")
			}

			members, err := lb.GetMembers([]string{"member-0", "invalid-member"}, "desc")
			Expect(err).NotTo(HaveOccurred())

			Expect(members).To(HaveLen(1))
			Expect(members[0].PublicID).To(Equal("member-0"))
			Expect(members[0].Rank).To(Equal(1))
			Expect(members[0].Score).To(Equal(100))
		})

		It("should fail with faulty redis", func() {
			lb := NewLeaderboard(faultyRedisClient.Client, uuid.NewV4().String(), 10, logger)
			_, err := lb.GetMembers([]string{}, "desc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})
})
