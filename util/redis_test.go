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
	"github.com/garyburd/redigo/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/util"
)

var _ = Describe("RedisClient", func() {

	testRedisSettings := util.RedisSettings{
		Host:     "localhost",
		Port:     1234,
		Password: "",
	}

	redisClient := util.GetRedisClient(testRedisSettings)

	BeforeSuite(func() {
		conn := redisClient.GetConnection()
		conn.Do("DEL", "test")
	})

	AfterSuite(func() {
		conn := redisClient.GetConnection()
		conn.Do("DEL", "test")
	})

	It("It should set and get without error", func() {
		conn := redisClient.GetConnection()
		_, err := conn.Do("set", "test", 1)
		Expect(err).To(BeNil())
		res, err := redis.Int(conn.Do("get", "test"))
		Expect(err).To(BeNil())
		Expect(res).To(Equal(1))
	})
})
