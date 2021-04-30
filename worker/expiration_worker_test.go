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
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/topfreegames/extensions/redis"
	"github.com/topfreegames/podium/leaderboard"
	"github.com/topfreegames/podium/testing"
	"github.com/topfreegames/podium/worker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scores Expirer Worker", func() {

	var redisClient *redis.Client
	var expirationWorker *worker.ExpirationWorker
	var leaderboards *leaderboard.Client
	expirationSink := make(chan []*worker.ExpirationResult)
	errorSink := make(chan error)

	go func() {
		for {
			select {
			case <-expirationSink:
			case <-errorSink:
			}
		}
	}()

	BeforeEach(func() {
		var err error

		config, err := testing.GetDefaultConfig("../config/test.yaml")
		Expect(err).NotTo(HaveOccurred())

		redisHost := config.GetString("redis.host")
		redisPort := config.GetInt("redis.port")
		redisDB := config.GetInt("redis.db")

		redisURL := fmt.Sprintf("redis://%s:%d/%d", redisHost, redisPort, redisDB)

		config.SetDefault("redis.url", redisURL)
		config.SetDefault("redis.connectionTimeout", 200)

		redisClient, err = redis.NewClient("redis", config)
		Expect(err).NotTo(HaveOccurred())

		leaderboards, err = leaderboard.NewClientWithRedis(redisClient)
		Expect(err).NotTo(HaveOccurred())

		p := redisClient.Client.TxPipeline()
		p.FlushAll()
		_, err = p.Exec()
		Expect(err).NotTo(HaveOccurred())

		expirationWorker, err = worker.GetExpirationWorker("../config/test.yaml")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		p := redisClient.Client.TxPipeline()
		p.FlushAll()
		_, err := p.Exec()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should expire scores and delete set", func() {
		ttl := "2"
		lbName := "test-expire-leaderboard"
		_, err := leaderboards.SetMemberScore(context.Background(), lbName, "denix", 481516, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		redisLBExpirationKey := fmt.Sprintf("%s:ttl", lbName)
		_, err = redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		redisExpirationSetKey := "expiration-sets"
		result, err := redisClient.Client.Exists(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(int64(1)))
		result2, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result2).To(ContainElement(redisLBExpirationKey))
		result3, err := redisClient.Client.ZRangeWithScores(redisLBExpirationKey, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result3[0].Member).To(Equal("denix"))
		ttlInt, _ := strconv.ParseInt(ttl, 10, 64)
		Expect(result3[0].Score).To(BeNumerically("~", time.Now().Unix()+ttlInt, 1))
		result4, err := redisClient.Client.ZRangeWithScores(lbName, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result4)).To(Equal(1))
		Expect(result4[0].Member).To(Equal("denix"))
		go func() {
			time.Sleep(time.Duration(6) * time.Second)
			expirationWorker.Stop()
		}()
		expirationWorker.Run(expirationSink, errorSink)

		res, err := redisClient.Client.ZRangeWithScores(lbName, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(0))

		members, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(members)).To(Equal(0))

		exists, err := redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(Equal(int64(0)))
	})

	It("should not expire scores that are in the future", func() {
		ttl := "20"
		lbName := "test-expire-leaderboard"
		_, err := leaderboards.SetMemberScore(context.Background(), lbName, "denix", 481516, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		redisLBExpirationKey := fmt.Sprintf("%s:ttl", lbName)
		_, err = redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		redisExpirationSetKey := "expiration-sets"
		result, err := redisClient.Client.Exists(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(int64(1)))
		result2, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result2).To(ContainElement(redisLBExpirationKey))
		result3, err := redisClient.Client.ZRangeWithScores(redisLBExpirationKey, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result3[0].Member).To(Equal("denix"))
		ttlInt, _ := strconv.ParseInt(ttl, 10, 64)
		Expect(result3[0].Score).To(BeNumerically("~", time.Now().Unix()+ttlInt, 1))
		result4, err := redisClient.Client.ZRangeWithScores(lbName, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result4)).To(Equal(1))
		Expect(result4[0].Member).To(Equal("denix"))
		go func() {
			time.Sleep(time.Duration(6) * time.Second)
			expirationWorker.Stop()
		}()
		expirationWorker.Run(expirationSink, errorSink)

		res, err := redisClient.Client.ZRangeWithScores(lbName, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(1))

		members, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(members)).To(Equal(1))

		exists, err := redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(Equal(int64(1)))
	})

	It("should not expire scores that are not inserted with scoreTTL set", func() {
		lbName := "test-expire-leaderboard"
		ttl := ""
		redisLBExpirationKey := fmt.Sprintf("%s:ttl", lbName)
		_, err := leaderboards.SetMemberScore(context.Background(), lbName, "denix", 481516, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		result, err := redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(int64(0)))
		redisExpirationSetKey := "expiration-sets"
		result, err = redisClient.Client.Exists(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(int64(0)))
		result4, err := redisClient.Client.ZRangeWithScores(lbName, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result4)).To(Equal(1))
		Expect(result4[0].Member).To(Equal("denix"))
		go func() {
			time.Sleep(time.Duration(5) * time.Second)
			expirationWorker.Stop()
		}()
		expirationWorker.Run(expirationSink, errorSink)

		res, err := redisClient.Client.ZRangeWithScores(lbName, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(1))

		members, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(members)).To(Equal(0))

		exists, err := redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(Equal(int64(0)))
	})

	It("a call to expireScores should only remove ExpirationLimitPerRun members from a set", func() {
		expirationWorker.ExpirationLimitPerRun = 1
		expirationWorker.ExpirationCheckInterval = time.Duration(4) * time.Second

		ttl := "2"
		lbName := "test-expire-leaderboard"
		_, err := leaderboards.SetMemberScore(context.Background(), lbName, "denix", 481516, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		_, err = leaderboards.SetMemberScore(context.Background(), lbName, "denix2", 481512, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		redisLBExpirationKey := fmt.Sprintf("%s:ttl", lbName)
		_, err = redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		redisExpirationSetKey := "expiration-sets"
		result, err := redisClient.Client.Exists(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(int64(1)))
		result2, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result2).To(ContainElement(redisLBExpirationKey))
		result3, err := redisClient.Client.ZRangeWithScores(redisLBExpirationKey, 0, 1).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(result3[0].Member).To(Equal("denix"))
		Expect(result3[1].Member).To(Equal("denix2"))
		result4, err := redisClient.Client.ZRangeWithScores(lbName, 0, 2).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result4)).To(Equal(2))
		Expect(result4[0].Member).To(Equal("denix2"))
		Expect(result4[1].Member).To(Equal("denix"))

		go func() {
			time.Sleep(time.Duration(6) * time.Second)
			expirationWorker.Stop()
		}()

		expirationWorker.Run(expirationSink, errorSink)

		res, err := redisClient.Client.ZRangeWithScores(lbName, 0, 2).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(1))
		Expect(res[0].Member).To(Equal("denix2"))

		members, err := redisClient.Client.SMembers(redisExpirationSetKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(members)).To(Equal(1))

		exists, err := redisClient.Client.Exists(redisLBExpirationKey).Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(Equal(int64(1)))
	})

	It("should create a valid expiration worker with external configuration", func() {
		config, err := testing.GetDefaultConfig("../config/test.yaml")
		Expect(err).NotTo(HaveOccurred())

		expirationWorker, err = worker.NewExpirationWorker(config.GetString("redis.host"),
			config.GetInt("redis.port"), config.GetString("redis.password"), config.GetInt("redis.db"),
			config.GetInt("redis.connectionTimeout"), config.GetDuration("worker.expirationCheckInterval"),
			config.GetInt("worker.expirationLimitPerRun"))
		Expect(err).NotTo(HaveOccurred())
	})

	It("should print correctly expiration results", func() {
		results := []*worker.ExpirationResult{
			{DeletedMembers: 1, DeletedSet: true, Set: "s1"},
			{DeletedMembers: 0, DeletedSet: false, Set: "s2"},
		}

		got := fmt.Sprintf("results: %v", results)
		want :=
			"results: [(DeletedMembers: 1, DeletedSet: true, Set: s1) (DeletedMembers: 0, DeletedSet: false, Set: s2)]"

		Expect(got).To(Equal(want))
	})
})
