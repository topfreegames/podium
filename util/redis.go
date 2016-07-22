// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package util

import (
	"fmt"

	"github.com/uber-go/zap"
	"gopkg.in/redis.v4"
)

// RedisClient identifies uniquely one redis client with a pool of connections
type RedisClient struct {
	Logger zap.Logger
	Client *redis.Client
}

// GetRedisClient creates and returns a new redis client based on the given settings
func GetRedisClient(redisHost string, redisPort int, redisPassword string, redisDB int, maxPoolSize int, logger zap.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPassword,
		DB:       redisDB,
		PoolSize: maxPoolSize,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	cl := &RedisClient{
		Client: client,
		Logger: logger,
	}
	return cl, nil
}

// GetConnection return a redis connection
func (c *RedisClient) GetConnection() *redis.Client {
	return c.Client
}
