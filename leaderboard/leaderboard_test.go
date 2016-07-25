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

	"gopkg.in/redis.v4"

	"github.com/satori/go.uuid"
	. "github.com/topfreegames/podium/leaderboard"
	"github.com/topfreegames/podium/testing"
	"github.com/topfreegames/podium/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Leaderboard Model", func() {

	var redisClient *util.RedisClient
	var faultyRedisClient *util.RedisClient
	var logger *testing.MockLogger

	BeforeEach(func() {
		var err error

		logger = testing.NewMockLogger()
		redisClient, err = util.GetRedisClient("localhost", 1234, "", 0, 50, logger)
		Expect(err).NotTo(HaveOccurred())

		//First we connect properly
		faultyRedisClient, err = util.GetRedisClient("localhost", 1234, "", 0, 50, logger)
		Expect(err).NotTo(HaveOccurred())
		//Then we change the connection to be faulty
		faultyRedisClient.Client = redis.NewClient(&redis.Options{
			Addr:     "localhost:1235",
			Password: "",
			DB:       0,
		})

		conn := redisClient.GetConnection()
		_, err = conn.Del("test-leaderboard").Result()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		conn := redisClient.GetConnection()
		_, err := conn.Del("test-leaderboard").Result()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("setting member scores", func() {
		It("should set scores and return ranks", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10, logger)
			dayvson, err := testLeaderboard.SetMemberScore("dayvson", 481516)
			Expect(err).NotTo(HaveOccurred())
			arthur, err := testLeaderboard.SetMemberScore("arthur", 1000)
			Expect(err).NotTo(HaveOccurred())
			Expect(dayvson.Rank).To(Equal(1))
			Expect(arthur.Rank).To(Equal(2))
		})
	})

	Describe("getting number of members", func() {
		It("should retrieve the number of members in a leaderboard", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10, logger)
			for i := 0; i < 10; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i)
				Expect(err).NotTo(HaveOccurred())
			}
			count, err := testLeaderboard.TotalMembers()
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(10))
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10, logger)
			testLeaderboard.RedisClient = faultyRedisClient
			_, err := testLeaderboard.TotalMembers()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("getsockopt: connection refused"))
		})
	})

	Describe("removing members", func() {
		It("should remove member", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10, logger)
			for i := 0; i < 10; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i)
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(testLeaderboard.TotalMembers()).To(Equal(10))
			testLeaderboard.RemoveMember("member_5")
			Expect(testLeaderboard.TotalMembers()).To(Equal(9))
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(faultyRedisClient, "test-leaderboard", 10, logger)
			_, err := testLeaderboard.RemoveMember("invalid member")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("getsockopt: connection refused"))
		})
	})

	Describe("getting number of pages in leaderboard", func() {
		It("should return total number of pages", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i)
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(testLeaderboard.TotalPages()).To(Equal(5))
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10, logger)
			testLeaderboard.RedisClient = faultyRedisClient
			_, err := testLeaderboard.TotalPages()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("getsockopt: connection refused"))
		})
	})

	Describe("getting member details for a given leaderboard", func() {
		It("should return member details", func() {
			friendScore := NewLeaderboard(redisClient, uuid.NewV4().String(), 10, logger)
			dayvson, err := friendScore.SetMemberScore("dayvson", 12345)
			Expect(err).NotTo(HaveOccurred())
			felipe, err := friendScore.SetMemberScore("felipe", 12344)
			Expect(err).NotTo(HaveOccurred())
			Expect(dayvson.Rank).To(Equal(1))
			Expect(felipe.Rank).To(Equal(2))
			friendScore.SetMemberScore("felipe", 12346)
			felipe, err = friendScore.GetMember("felipe")
			Expect(err).NotTo(HaveOccurred())
			dayvson, err = friendScore.GetMember("dayvson")
			Expect(err).NotTo(HaveOccurred())
			Expect(felipe.Rank).To(Equal(1))
			Expect(dayvson.Rank).To(Equal(2))
		})

		It("should fail if member does not exist", func() {
			leaderboardID := uuid.NewV4().String()
			friendScore := NewLeaderboard(redisClient, leaderboardID, 10, logger)
			memberID := uuid.NewV4().String()
			member, err := friendScore.GetMember(memberID)
			Expect(member).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				fmt.Sprintf("Could not find data for member %s in leaderboard %s.", memberID, leaderboardID)),
			)
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10, logger)
			testLeaderboard.RedisClient = faultyRedisClient
			_, err := testLeaderboard.TotalPages()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("getsockopt: connection refused"))
		})
	})

	Describe("get members around someone in a leaderboard", func() {
		It("should get members around specific member", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i)
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetAroundMe("member_20")
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[testLeaderboard.PageSize-1]
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			Expect(firstAroundMe.PublicID).To(Equal("member_31"))
			Expect(lastAroundMe.PublicID).To(Equal("member_7"))
		})

		It("should get members around specific member if repeated scores", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 100)
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetAroundMe("member_20")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			firstAroundMe := members[0]
			lastAroundMe := members[testLeaderboard.PageSize-1]
			Expect(firstAroundMe.Score).To(Equal(100))
			Expect(lastAroundMe.Score).To(Equal(100))
		})

		It("should get PageSize members around specific member even if member in ranking top", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 1; i <= 100; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 100-i)
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetAroundMe("member_2")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			firstAroundMe := members[0]
			lastAroundMe := members[testLeaderboard.PageSize-1]
			Expect(firstAroundMe.PublicID).To(Equal("member_1"))
			Expect(lastAroundMe.PublicID).To(Equal("member_25"))
		})

		It("should get PageSize members around specific member even if member in ranking bottom", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 1; i <= 100; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 100-i)
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetAroundMe("member_99")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			firstAroundMe := members[0]
			lastAroundMe := members[testLeaderboard.PageSize-1]
			Expect(firstAroundMe.PublicID).To(Equal("member_76"))
			Expect(lastAroundMe.PublicID).To(Equal("member_100"))
		})
	})

	Describe("getting member ranking", func() {
		It("should return specific member ranking", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 1234*i)
				Expect(err).NotTo(HaveOccurred())
			}
			testLeaderboard.SetMemberScore("member_6", 1000)
			Expect(testLeaderboard.GetRank("member_6")).To(Equal(100))
		})
	})

	Describe("getting leaderboard leaders", func() {
		It("should get specific number of leaders", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 0; i < 1000; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i+1), 1234*i)
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetLeaders(1)
			Expect(err).NotTo(HaveOccurred())

			firstOnPage := members[0]
			lastOnPage := members[len(members)-1]
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			Expect(firstOnPage.PublicID).To(Equal("member_1000"))
			Expect(firstOnPage.Rank).To(Equal(1))
			Expect(lastOnPage.PublicID).To(Equal("member_976"))
			Expect(lastOnPage.Rank).To(Equal(25))
		})

		It("should get leaders if repeated scores", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetMemberScore("member_"+strconv.Itoa(i), 100)
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := testLeaderboard.GetLeaders(1)
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[testLeaderboard.PageSize-1]
			Expect(len(members)).To(Equal(testLeaderboard.PageSize))
			Expect(firstAroundMe.Score).To(Equal(100))
			Expect(lastAroundMe.Score).To(Equal(100))
		})
	})

	Describe("expiration of leaderboards", func() {
		It("should fail if invalid leaderboard", func() {
			leaderboardID := "leaderboard_from20201039to20201011"
			friendScore := NewLeaderboard(redisClient, leaderboardID, 10, logger)
			_, err := friendScore.SetMemberScore("dayvson", 12345)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("day out of range"))
		})

		It("should add yearly expiration if leaderboard supports it", func() {
			leaderboardID := "test-leaderboard-year2016"
			friendScore := NewLeaderboard(redisClient, leaderboardID, 10, logger)
			_, err := friendScore.SetMemberScore("dayvson", 12345)
			Expect(err).NotTo(HaveOccurred())

			conn := redisClient.GetConnection()
			result, err := conn.TTL(leaderboardID).Result()
			Expect(err).NotTo(HaveOccurred())

			exp := result.Seconds()
			Expect(err).NotTo(HaveOccurred())
			Expect(exp).To(BeNumerically(">", int64(-1)))
		})
	})

	Describe("get top x percent of members in the leaderboard", func() {
		It("should get top 10 percent members in the leaderboard", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient, leaderboardID, 10, logger)

			members := []*Member{}
			for i := 0; i < 100; i++ {
				member, err := leader.SetMemberScore(fmt.Sprintf("friend-%d", i), (100-i)*100)
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leader.GetTopPercentage(10, 2000)
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(10))

			Expect(top10[0].PublicID).To(Equal("friend-0"))
			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(10000))

			Expect(top10[9].PublicID).To(Equal("friend-9"))
			Expect(top10[9].Rank).To(Equal(10))
			Expect(top10[9].Score).To(Equal(9100))
		})

		It("should get max members if query too broad", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient, leaderboardID, 10, logger)

			members := []*Member{}
			for i := 0; i < 10; i++ {
				member, err := leader.SetMemberScore(fmt.Sprintf("friend-%d", i), (100-i)*100)
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top3, err := leader.GetTopPercentage(100, 3)
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
			leader := NewLeaderboard(redisClient, leaderboardID, 10, logger)

			members := []*Member{}
			for i := 0; i < 2; i++ {
				member, err := leader.SetMemberScore(fmt.Sprintf("friend-%d", i), (100-i)*100)
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leader.GetTopPercentage(1, 2000)
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(1))

			Expect(top10[0].PublicID).To(Equal("friend-0"))
			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(10000))
		})

		It("should get top 10 percent members in the leaderboard if repeated scores", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient, leaderboardID, 10, logger)

			members := []*Member{}
			for i := 0; i < 100; i++ {
				member, err := leader.SetMemberScore(fmt.Sprintf("friend-%d", i), 100)
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leader.GetTopPercentage(10, 2000)
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(10))

			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(100))

			Expect(top10[9].Rank).To(Equal(10))
			Expect(top10[9].Score).To(Equal(100))
		})

		It("should fail if more than 100 percent", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient, leaderboardID, 10, logger)

			top10, err := leader.GetTopPercentage(101, 2000)
			Expect(top10).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Percentage must be a valid integer between 1 and 100."))
		})
	})
})
