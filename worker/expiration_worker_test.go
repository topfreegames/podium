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

	"github.com/topfreegames/podium/config"
	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/database/redis"
	lservice "github.com/topfreegames/podium/leaderboard/v2/service"
	"github.com/topfreegames/podium/worker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scores Expirer Worker", func() {

	var redisClient *database.Redis
	var expirationWorker *worker.ExpirationWorker
	var leaderboards lservice.Leaderboard

	const lbName string = "test-expire-leaderboard"

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

		expirationWorker, err = worker.GetExpirationWorker("../config/test.yaml")
		redisClient = database.NewRedisDatabase(database.RedisOptions{
			ClusterEnabled: expirationWorker.Config.GetBool("redis.cluster.enabled"),
			Addrs:          expirationWorker.Config.GetStringSlice("redis.addrs"),
			Host:           expirationWorker.Config.GetString("redis.host"),
			Port:           expirationWorker.Config.GetInt("redis.port"),
			Password:       expirationWorker.Config.GetString("redis.password"),
			DB:             expirationWorker.Config.GetInt("redis.db"),
		})
		leaderboards = lservice.NewService(redisClient)

		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		redisClient.Del(context.Background(), lbName)
		redisClient.Del(context.Background(), fmt.Sprintf("%s:ttl", lbName))
		redisClient.Del(context.Background(), database.ExpirationSet)
	})

	It("should expire scores and delete set", func() {
		ttl := "1"
		_, err := leaderboards.SetMemberScore(context.Background(), lbName, "denix", 481516, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		redisLBExpirationKey := fmt.Sprintf("%s:ttl", lbName)
		err = redisClient.Exists(context.Background(), redisLBExpirationKey)
		Expect(err).NotTo(HaveOccurred())
		err = redisClient.Exists(context.Background(), database.ExpirationSet)
		Expect(err).NotTo(HaveOccurred())
		result2, err := redisClient.SMembers(context.Background(), database.ExpirationSet)
		Expect(err).NotTo(HaveOccurred())
		Expect(result2).To(ContainElement(redisLBExpirationKey))
		result3, err := redisClient.ZRange(context.Background(), redisLBExpirationKey, 0, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(result3[0].Member).To(Equal("denix"))
		ttlInt, _ := strconv.ParseInt(ttl, 10, 64)
		Expect(result3[0].Score).To(BeNumerically("~", time.Now().Unix()+ttlInt, 1))
		result4, err := redisClient.ZRange(context.Background(), lbName, 0, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result4)).To(Equal(1))
		Expect(result4[0].Member).To(Equal("denix"))
		go func() {
			time.Sleep(time.Duration(6) * time.Second)
			expirationWorker.Stop()
		}()
		expirationWorker.Run(expirationSink, errorSink)

		res, err := redisClient.ZRange(context.Background(), lbName, 0, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(0))

		members, err := redisClient.SMembers(context.Background(), database.ExpirationSet)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(members)).To(Equal(0))

		err = redisClient.Exists(context.Background(), redisLBExpirationKey)
		Expect(err).To(MatchError(redis.NewKeyNotFoundError(redisLBExpirationKey)))
	})

	It("should not expire scores that are in the future", func() {
		ttl := "20"
		_, err := leaderboards.SetMemberScore(context.Background(), lbName, "denix", 481516, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		redisLBExpirationKey := fmt.Sprintf("%s:ttl", lbName)
		err = redisClient.Exists(context.Background(), redisLBExpirationKey)
		Expect(err).NotTo(HaveOccurred())
		err = redisClient.Exists(context.Background(), database.ExpirationSet)
		Expect(err).NotTo(HaveOccurred())
		result2, err := redisClient.SMembers(context.Background(), database.ExpirationSet)
		Expect(err).NotTo(HaveOccurred())
		Expect(result2).To(ContainElement(redisLBExpirationKey))
		result3, err := redisClient.ZRange(context.Background(), redisLBExpirationKey, 0, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(result3[0].Member).To(Equal("denix"))
		ttlInt, _ := strconv.ParseInt(ttl, 10, 64)
		Expect(result3[0].Score).To(BeNumerically("~", time.Now().Unix()+ttlInt, 1))
		result4, err := redisClient.ZRange(context.Background(), lbName, 0, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result4)).To(Equal(1))
		Expect(result4[0].Member).To(Equal("denix"))
		go func() {
			time.Sleep(time.Duration(6) * time.Second)
			expirationWorker.Stop()
		}()
		expirationWorker.Run(expirationSink, errorSink)

		res, err := redisClient.ZRange(context.Background(), lbName, 0, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(1))

		members, err := redisClient.SMembers(context.Background(), database.ExpirationSet)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(members)).To(Equal(1))

		err = redisClient.Exists(context.Background(), redisLBExpirationKey)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should not expire scores that are not inserted with scoreTTL set", func() {
		ttl := ""
		redisLBExpirationKey := fmt.Sprintf("%s:ttl", lbName)
		_, err := leaderboards.SetMemberScore(context.Background(), lbName, "denix", 481516, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		err = redisClient.Exists(context.Background(), redisLBExpirationKey)
		Expect(err).To(MatchError(redis.NewKeyNotFoundError(redisLBExpirationKey)))
		err = redisClient.Exists(context.Background(), database.ExpirationSet)
		Expect(err).To(MatchError(redis.NewKeyNotFoundError(database.ExpirationSet)))
		result4, err := redisClient.ZRange(context.Background(), lbName, 0, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result4)).To(Equal(1))
		Expect(result4[0].Member).To(Equal("denix"))
		go func() {
			time.Sleep(time.Duration(5) * time.Second)
			expirationWorker.Stop()
		}()
		expirationWorker.Run(expirationSink, errorSink)

		res, err := redisClient.ZRange(context.Background(), lbName, 0, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(1))

		members, err := redisClient.SMembers(context.Background(), database.ExpirationSet)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(members)).To(Equal(0))

		err = redisClient.Exists(context.Background(), redisLBExpirationKey)
		Expect(err).To(MatchError(redis.NewKeyNotFoundError(redisLBExpirationKey)))
	})

	It("a call to expireScores should only remove ExpirationLimitPerRun members from a set", func() {
		expirationWorker.ExpirationLimitPerRun = 1
		expirationWorker.ExpirationCheckInterval = time.Duration(4) * time.Second

		ttl := "2"
		_, err := leaderboards.SetMemberScore(context.Background(), lbName, "denix", 481516, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		_, err = leaderboards.SetMemberScore(context.Background(), lbName, "denix2", 481512, false, ttl)
		Expect(err).NotTo(HaveOccurred())
		redisLBExpirationKey := fmt.Sprintf("%s:ttl", lbName)
		err = redisClient.Exists(context.Background(), redisLBExpirationKey)
		Expect(err).NotTo(HaveOccurred())
		err = redisClient.Exists(context.Background(), database.ExpirationSet)
		Expect(err).NotTo(HaveOccurred())
		result2, err := redisClient.SMembers(context.Background(), database.ExpirationSet)
		Expect(err).NotTo(HaveOccurred())
		Expect(result2).To(ContainElement(redisLBExpirationKey))
		result3, err := redisClient.ZRange(context.Background(), redisLBExpirationKey, 0, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(result3[0].Member).To(Equal("denix"))
		Expect(result3[1].Member).To(Equal("denix2"))
		result4, err := redisClient.ZRange(context.Background(), lbName, 0, 2)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(result4)).To(Equal(2))
		Expect(result4[0].Member).To(Equal("denix2"))
		Expect(result4[1].Member).To(Equal("denix"))

		go func() {
			time.Sleep(time.Duration(6) * time.Second)
			expirationWorker.Stop()
		}()

		expirationWorker.Run(expirationSink, errorSink)

		res, err := redisClient.ZRange(context.Background(), lbName, 0, 2)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(res)).To(Equal(1))
		Expect(res[0].Member).To(Equal("denix2"))

		members, err := redisClient.SMembers(context.Background(), database.ExpirationSet)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(members)).To(Equal(1))

		err = redisClient.Exists(context.Background(), redisLBExpirationKey)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should create a valid expiration worker with external configuration", func() {
		config, err := config.GetDefaultConfig("../config/test.yaml")
		Expect(err).NotTo(HaveOccurred())

		expirationWorker, err = worker.NewExpirationWorker(config.GetString("redis.host"),
			config.GetInt("redis.port"), config.GetString("redis.password"), config.GetInt("redis.db"),
			config.GetDuration("worker.expirationCheckInterval"), config.GetInt("worker.expirationLimitPerRun"))
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
