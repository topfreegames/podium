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

	Describe("setting user scores", func() {
		It("should set scores and return ranks", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10, logger)
			dayvson, err := testLeaderboard.SetUserScore("dayvson", 481516)
			Expect(err).NotTo(HaveOccurred())
			arthur, err := testLeaderboard.SetUserScore("arthur", 1000)
			Expect(err).NotTo(HaveOccurred())
			Expect(dayvson.Rank).To(Equal(1))
			Expect(arthur.Rank).To(Equal(2))
		})
	})

	Describe("getting number of members", func() {
		It("should retrieve the number of members in a leaderboard", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10, logger)
			for i := 0; i < 10; i++ {
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
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
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
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
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
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

	Describe("getting user details for a given leaderboard", func() {
		It("should return user details", func() {
			friendScore := NewLeaderboard(redisClient, uuid.NewV4().String(), 10, logger)
			dayvson, err := friendScore.SetUserScore("dayvson", 12345)
			Expect(err).NotTo(HaveOccurred())
			felipe, err := friendScore.SetUserScore("felipe", 12344)
			Expect(err).NotTo(HaveOccurred())
			Expect(dayvson.Rank).To(Equal(1))
			Expect(felipe.Rank).To(Equal(2))
			friendScore.SetUserScore("felipe", 12346)
			felipe, err = friendScore.GetMember("felipe")
			Expect(err).NotTo(HaveOccurred())
			dayvson, err = friendScore.GetMember("dayvson")
			Expect(err).NotTo(HaveOccurred())
			Expect(felipe.Rank).To(Equal(1))
			Expect(dayvson.Rank).To(Equal(2))
		})

		It("should fail if user does not exist", func() {
			leaderboardID := uuid.NewV4().String()
			friendScore := NewLeaderboard(redisClient, leaderboardID, 10, logger)
			playerID := uuid.NewV4().String()
			player, err := friendScore.GetMember(playerID)
			Expect(player).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				fmt.Sprintf("Could not find data for user %s in leaderboard %s.", playerID, leaderboardID)),
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

	Describe("get users around someone in a leaderboard", func() {
		It("should get users around specific user", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
				Expect(err).NotTo(HaveOccurred())
			}
			users, err := testLeaderboard.GetAroundMe("member_20")
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := users[0]
			lastAroundMe := users[testLeaderboard.PageSize-1]
			Expect(len(users)).To(Equal(testLeaderboard.PageSize))
			Expect(firstAroundMe.PublicID).To(Equal("member_31"))
			Expect(lastAroundMe.PublicID).To(Equal("member_7"))
		})

		It("should get users around specific user if repeated scores", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 100)
				Expect(err).NotTo(HaveOccurred())
			}
			users, err := testLeaderboard.GetAroundMe("member_20")
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := users[0]
			lastAroundMe := users[testLeaderboard.PageSize-1]
			Expect(len(users)).To(Equal(testLeaderboard.PageSize))
			Expect(firstAroundMe.Score).To(Equal(100))
			Expect(lastAroundMe.Score).To(Equal(100))
		})
	})

	Describe("getting player ranking", func() {
		It("should return specific player ranking", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
				Expect(err).NotTo(HaveOccurred())
			}
			testLeaderboard.SetUserScore("member_6", 1000)
			Expect(testLeaderboard.GetRank("member_6")).To(Equal(100))
		})
	})

	Describe("getting leaderboard leaders", func() {
		It("should get specific number of leaders", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 0; i < 1000; i++ {
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i+1), 1234*i)
				Expect(err).NotTo(HaveOccurred())
			}
			users, err := testLeaderboard.GetLeaders(1)
			Expect(err).NotTo(HaveOccurred())

			firstOnPage := users[0]
			lastOnPage := users[len(users)-1]
			Expect(len(users)).To(Equal(testLeaderboard.PageSize))
			Expect(firstOnPage.PublicID).To(Equal("member_1000"))
			Expect(firstOnPage.Rank).To(Equal(1))
			Expect(lastOnPage.PublicID).To(Equal("member_976"))
			Expect(lastOnPage.Rank).To(Equal(25))
		})

		It("should get leaders if repeated scores", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25, logger)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 100)
				Expect(err).NotTo(HaveOccurred())
			}
			users, err := testLeaderboard.GetLeaders(1)
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := users[0]
			lastAroundMe := users[testLeaderboard.PageSize-1]
			Expect(len(users)).To(Equal(testLeaderboard.PageSize))
			Expect(firstAroundMe.Score).To(Equal(100))
			Expect(lastAroundMe.Score).To(Equal(100))
		})
	})

	Describe("expiration of leaderboards", func() {
		It("should fail if invalid leaderboard", func() {
			leaderboardID := "leaderboard_from20201039to20201011"
			friendScore := NewLeaderboard(redisClient, leaderboardID, 10, logger)
			_, err := friendScore.SetUserScore("dayvson", 12345)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("day out of range"))
		})

		It("should add yearly expiration if leaderboard supports it", func() {
			leaderboardID := "test-leaderboard-year2016"
			friendScore := NewLeaderboard(redisClient, leaderboardID, 10, logger)
			_, err := friendScore.SetUserScore("dayvson", 12345)
			Expect(err).NotTo(HaveOccurred())

			conn := redisClient.GetConnection()
			result, err := conn.TTL(leaderboardID).Result()
			Expect(err).NotTo(HaveOccurred())

			exp := result.Seconds()
			Expect(err).NotTo(HaveOccurred())
			Expect(exp).To(BeNumerically(">", int64(-1)))
		})
	})

	Describe("get top x percent of players in the leaderboard", func() {
		It("should get top 10 percent players in the leaderboard", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient, leaderboardID, 10, logger)

			members := []*User{}
			for i := 0; i < 100; i++ {
				member, err := leader.SetUserScore(fmt.Sprintf("friend-%d", i), (100-i)*100)
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

			members := []*User{}
			for i := 0; i < 10; i++ {
				member, err := leader.SetUserScore(fmt.Sprintf("friend-%d", i), (100-i)*100)
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

			members := []*User{}
			for i := 0; i < 2; i++ {
				member, err := leader.SetUserScore(fmt.Sprintf("friend-%d", i), (100-i)*100)
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

		It("should get top 10 percent players in the leaderboard if repeated scores", func() {
			leaderboardID := uuid.NewV4().String()
			leader := NewLeaderboard(redisClient, leaderboardID, 10, logger)

			members := []*User{}
			for i := 0; i < 100; i++ {
				member, err := leader.SetUserScore(fmt.Sprintf("friend-%d", i), 100)
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
