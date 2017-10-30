// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package worker_test

import (
	"fmt"
	"time"

	"github.com/topfreegames/podium/leaderboard"
	"github.com/topfreegames/podium/testing"
	"github.com/topfreegames/podium/util"
	"github.com/topfreegames/podium/worker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scores Expirer Worker", func() {

	var redisClient *util.RedisClient
	var logger *testing.MockLogger
	var expirationWorker *worker.ExpirationWorker

	BeforeEach(func() {
		var err error

		logger = testing.NewMockLogger()
		redisClient, err = util.GetRedisClient("localhost", 1234, "", 0, 50, logger)
		Expect(err).NotTo(HaveOccurred())

		conn := redisClient.GetConnection()
		_, err = conn.Del("test-expire-leaderboard", "expiration-sets", "test-expire-leaderboard:ttl:2", "test-expire-leaderboard:ttl:20").Result()
		Expect(err).NotTo(HaveOccurred())

		expirationWorker, err = worker.GetExpirationWorker("../config/test.yaml", logger)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		conn := redisClient.GetConnection()
		_, err := conn.Del("test-expire-leaderboard", "expiration-sets", "test-expire-leaderboard:ttl:2", "test-expire-leaderboard:ttl:20").Result()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should expire scores and delete set", func() {
		ttl := "2"
		lbName := "test-expire-leaderboard"
		testLeaderboard := leaderboard.NewLeaderboard(redisClient, lbName, 10, logger)
		_, err := testLeaderboard.SetMemberScore("denix", 481516, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		redisLBExpirationKey := fmt.Sprintf("%s:ttl:%s", lbName, ttl)
		result, err := redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		redisExpirationSetKey := "expiration-sets"
		result, err = redisClient.Client.Exists(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(true))
		result2, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result2).To(ContainElement(redisLBExpirationKey))
		result3, err := redisClient.Client.ZRangeWithScores(redisLBExpirationKey, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result3[0].Member).To(Equal("denix"))
		Expect(result3[0].Score).To(BeNumerically("<=", time.Now().Unix()))
		result4, err := redisClient.Client.ZRangeWithScores(lbName, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result4)).To(Equal(1))
		Expect(result4[0].Member).To(Equal("denix"))
		go func() {
			time.Sleep(time.Duration(6) * time.Second)
			expirationWorker.Stop()
		}()
		expirationWorker.Run()
		res, err := redisClient.Client.ZRangeWithScores(lbName, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(0))

		members, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(members)).To(Equal(0))

		exists, err := redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeFalse())
	})

	It("should not expire scores that are in the future", func() {
		ttl := "20"
		lbName := "test-expire-leaderboard"
		testLeaderboard := leaderboard.NewLeaderboard(redisClient, lbName, 10, logger)
		_, err := testLeaderboard.SetMemberScore("denix", 481516, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		redisLBExpirationKey := fmt.Sprintf("%s:ttl:%s", lbName, ttl)
		result, err := redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		redisExpirationSetKey := "expiration-sets"
		result, err = redisClient.Client.Exists(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(true))
		result2, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result2).To(ContainElement(redisLBExpirationKey))
		result3, err := redisClient.Client.ZRangeWithScores(redisLBExpirationKey, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result3[0].Member).To(Equal("denix"))
		Expect(result3[0].Score).To(BeNumerically("<=", time.Now().Unix()))
		result4, err := redisClient.Client.ZRangeWithScores(lbName, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result4)).To(Equal(1))
		Expect(result4[0].Member).To(Equal("denix"))
		go func() {
			time.Sleep(time.Duration(6) * time.Second)
			expirationWorker.Stop()
		}()
		expirationWorker.Run()
		res, err := redisClient.Client.ZRangeWithScores(lbName, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(1))

		members, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(members)).To(Equal(1))

		exists, err := redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeTrue())
	})

	It("should not expire scores that are not inserted with scoreTTL set", func() {
		ttl := ""
		lbName := "test-expire-leaderboard"
		testLeaderboard := leaderboard.NewLeaderboard(redisClient, lbName, 10, logger)
		_, err := testLeaderboard.SetMemberScore("denix", 481516, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		redisLBExpirationKey := fmt.Sprintf("%s:ttl:%s", lbName, ttl)
		result, err := redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(false))
		redisExpirationSetKey := "expiration-sets"
		result, err = redisClient.Client.Exists(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(false))
		result4, err := redisClient.Client.ZRangeWithScores(lbName, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result4)).To(Equal(1))
		Expect(result4[0].Member).To(Equal("denix"))
		go func() {
			time.Sleep(time.Duration(5) * time.Second)
			expirationWorker.Stop()
		}()
		expirationWorker.Run()
		res, err := redisClient.Client.ZRangeWithScores(lbName, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(1))

		members, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(members)).To(Equal(0))

		exists, err := redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeFalse())
	})

})
