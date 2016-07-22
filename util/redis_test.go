// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package util_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/testing"
	"github.com/topfreegames/podium/util"
)

var _ = Describe("RedisClient", func() {

	logger := testing.NewMockLogger()
	var redisClient *util.RedisClient

	BeforeEach(func() {
		var err error
		redisClient, err = util.GetRedisClient("localhost", 1234, "", 0, 50, logger)
		Expect(err).NotTo(HaveOccurred())
		conn := redisClient.GetConnection()
		_, err = conn.Del("test").Result()
		Expect(err).NotTo(HaveOccurred())
	})

	It("It should set and get without error", func() {
		conn := redisClient.GetConnection()
		_, err := conn.Set("test", 1, time.Duration(-1)).Result()
		Expect(err).To(BeNil())
		res, err := conn.Get("test").Result()
		Expect(err).To(BeNil())
		Expect(res).To(BeEquivalentTo("1"))
	})
})
