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

	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// RedisSettings identifies uniquely the settings of a redis client
type RedisSettings struct {
	Host     string
	Port     int
	Password string
}

// RedisClient identifies uniquely one redis client with a pool of connections
type RedisClient struct {
	Logger zap.Logger
	Pool   *redis.Pool
}

var client *RedisClient

func newPool(host string, port int, password string) *redis.Pool {
	redisAddress := fmt.Sprintf("%s:%d", host, port)
	return redis.NewPool(func() (redis.Conn, error) {
		if viper.GetString("redis.password") != "" {
			c, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", viper.GetString("redis.host"),
				viper.GetInt("redis.port")), redis.DialPassword(viper.GetString("redis.password")))
			if err != nil {
				client.Logger.Error(err.Error())
			}
			return c, err
		}
		c, err := redis.Dial("tcp", redisAddress)
		if err != nil {
			if err != nil {
				client.Logger.Error(err.Error())
			}
		}
		return c, err
	}, viper.GetInt("redis.maxPoolSize"))
}

// GetRedisClient creates and returns a new redis client based on the given settings
func GetRedisClient(settings RedisSettings) *RedisClient {
	client = &RedisClient{
		Logger: zap.NewJSON(zap.WarnLevel),
	}
	if client.Pool == nil {
		client.Pool = newPool(settings.Host, settings.Port, settings.Password)
	}
	return client
}

// GetConnection return a redis connection
func (c *RedisClient) GetConnection() redis.Conn {
	return c.Pool.Get()
}
