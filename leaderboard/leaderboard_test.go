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

	"github.com/satori/go.uuid"
	. "github.com/topfreegames/podium/leaderboard"
	"github.com/topfreegames/podium/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Leaderboard Model", func() {

	var redisSettings util.RedisSettings
	var redisClient *util.RedisClient
	var faultyRedisClient *util.RedisClient

	BeforeEach(func() {
		redisSettings = util.RedisSettings{
			Host:     "localhost",
			Port:     1234,
			Password: "",
		}

		redisClient = util.GetRedisClient(redisSettings)

		redisSettings = util.RedisSettings{
			Host:     "localhost",
			Port:     1235,
			Password: "",
		}
		faultyRedisClient = util.GetRedisClient(redisSettings)

		conn := redisClient.GetConnection()
		conn.Do("DEL", "test-leaderboard")
	})

	AfterSuite(func() {
		conn := redisClient.GetConnection()
		conn.Do("DEL", "test-leaderboard")
	})

	Describe("setting user scores", func() {
		It("should set scores and return ranks", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10)
			dayvson, err := testLeaderboard.SetUserScore("dayvson", 481516)
			Expect(err).To(BeNil())
			arthur, err := testLeaderboard.SetUserScore("arthur", 1000)
			Expect(err).To(BeNil())
			Expect(dayvson.Rank).To(Equal(1))
			Expect(arthur.Rank).To(Equal(2))
		})
	})

	Describe("getting number of members", func() {
		It("should retrieve the number of members in a leaderboard", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10)
			for i := 0; i < 10; i++ {
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
				Expect(err).To(BeNil())
			}
			count, err := testLeaderboard.TotalMembers()
			Expect(err).To(BeNil())
			Expect(count).To(Equal(10))
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10)
			testLeaderboard.RedisClient = faultyRedisClient
			_, err := testLeaderboard.TotalMembers()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("getsockopt: connection refused"))
		})
	})

	Describe("removing members", func() {
		It("should remove member", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10)
			for i := 0; i < 10; i++ {
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
				Expect(err).To(BeNil())
			}
			Expect(testLeaderboard.TotalMembers()).To(Equal(10))
			testLeaderboard.RemoveMember("member_5")
			Expect(testLeaderboard.TotalMembers()).To(Equal(9))
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(faultyRedisClient, "test-leaderboard", 10)
			_, err := testLeaderboard.RemoveMember("invalid member")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("getsockopt: connection refused"))
		})
	})

	Describe("getting number of pages in leaderboard", func() {
		It("should return total number of pages", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
				Expect(err).To(BeNil())
			}
			Expect(testLeaderboard.TotalPages()).To(Equal(5))
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10)
			testLeaderboard.RedisClient = faultyRedisClient
			_, err := testLeaderboard.TotalPages()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("getsockopt: connection refused"))
		})
	})

	Describe("getting user details for a given leaderboard", func() {
		It("should return user details", func() {
			friendScore := NewLeaderboard(redisClient, uuid.NewV4().String(), 10)
			dayvson, err := friendScore.SetUserScore("dayvson", 12345)
			Expect(err).To(BeNil())
			felipe, err := friendScore.SetUserScore("felipe", 12344)
			Expect(err).To(BeNil())
			Expect(dayvson.Rank).To(Equal(1))
			Expect(felipe.Rank).To(Equal(2))
			friendScore.SetUserScore("felipe", 12346)
			felipe, err = friendScore.GetMember("felipe")
			Expect(err).To(BeNil())
			dayvson, err = friendScore.GetMember("dayvson")
			Expect(err).To(BeNil())
			Expect(felipe.Rank).To(Equal(1))
			Expect(dayvson.Rank).To(Equal(2))
		})

		It("should fail if user does not exist", func() {
			leaderboardID := uuid.NewV4().String()
			friendScore := NewLeaderboard(redisClient, leaderboardID, 10)
			playerID := uuid.NewV4().String()
			player, err := friendScore.GetMember(playerID)
			Expect(player).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				fmt.Sprintf("Could not find data for user %s in leaderboard %s.", playerID, leaderboardID)),
			)
		})

		It("should fail if faulty redis client", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 10)
			testLeaderboard.RedisClient = faultyRedisClient
			_, err := testLeaderboard.TotalPages()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("getsockopt: connection refused"))
		})
	})

	Describe("get users around someone in a leaderboard", func() {
		It("should get users around specific user", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
				Expect(err).To(BeNil())
			}
			users, err := testLeaderboard.GetAroundMe("member_20")
			Expect(err).To(BeNil())
			firstAroundMe := users[0]
			lastAroundMe := users[testLeaderboard.PageSize-1]
			Expect(len(users)).To(Equal(testLeaderboard.PageSize))
			Expect(firstAroundMe.PublicID).To(Equal("member_31"))
			Expect(lastAroundMe.PublicID).To(Equal("member_7"))
		})
	})

	Describe("getting player ranking", func() {
		It("should return specific player ranking", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25)
			for i := 0; i < 101; i++ {
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
				Expect(err).To(BeNil())
			}
			testLeaderboard.SetUserScore("member_6", 1000)
			Expect(testLeaderboard.GetRank("member_6")).To(Equal(100))
		})
	})

	Describe("getting leaderboard leaders", func() {
		It("should get specific number of leaders", func() {
			testLeaderboard := NewLeaderboard(redisClient, "test-leaderboard", 25)
			for i := 0; i < 1000; i++ {
				_, err := testLeaderboard.SetUserScore("member_"+strconv.Itoa(i+1), 1234*i)
				Expect(err).To(BeNil())
			}
			users, err := testLeaderboard.GetLeaders(1)
			Expect(err).To(BeNil())

			firstOnPage := users[0]
			lastOnPage := users[len(users)-1]
			Expect(len(users)).To(Equal(testLeaderboard.PageSize))
			Expect(firstOnPage.PublicID).To(Equal("member_1000"))
			Expect(firstOnPage.Rank).To(Equal(1))
			Expect(lastOnPage.PublicID).To(Equal("member_976"))
			Expect(lastOnPage.Rank).To(Equal(25))
		})
	})

	Describe("expiration of leaderboards", func() {
		It("should fail if invalid leaderboard", func() {
			leaderboardID := "leaderboard_from20201039to20201011"
			friendScore := NewLeaderboard(redisClient, leaderboardID, 10)
			_, err := friendScore.SetUserScore("dayvson", 12345)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("day out of range"))
		})

		It("should add yearly expiration if leaderboard supports it", func() {
			leaderboardID := "test-leaderboard-year2016"
			friendScore := NewLeaderboard(redisClient, leaderboardID, 10)
			_, err := friendScore.SetUserScore("dayvson", 12345)
			Expect(err).To(BeNil())

			conn := redisClient.GetConnection()
			result, err := conn.Do("TTL", leaderboardID)
			Expect(err).NotTo(HaveOccurred())

			exp := result.(int64)
			Expect(err).NotTo(HaveOccurred())
			Expect(exp).To(BeNumerically(">", int64(-1)))
		})
	})
})
