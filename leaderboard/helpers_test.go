package leaderboard_test

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/testing"
)

//Creates an empty context (shortcut for context.Background())
func NewEmptyCtx() context.Context {
	return context.Background()
}

func GetDefaultRedis() (*database.Redis, error) {
	config, err := testing.GetDefaultConfig("../config/test.yaml")
	if err != nil {
		return nil, err
	}

	return database.NewRedisDatabase(database.RedisOptions{
		ClusterEnabled: config.GetBool("redis.cluster.enabled"),
		Addrs:          config.GetStringSlice("redis.addrs"),
		Host:           config.GetString("redis.host"),
		Port:           config.GetInt("redis.port"),
		Password:       config.GetString("redis.password"),
		DB:             config.GetInt("redis.db"),
	}), nil
}
