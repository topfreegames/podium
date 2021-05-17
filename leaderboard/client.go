package leaderboard

import (
	"github.com/topfreegames/podium/leaderboard/database"
	"github.com/topfreegames/podium/leaderboard/database/redis"
	"github.com/topfreegames/podium/leaderboard/service"
)

// NewClient creates a leaderboard prepared to receive commands (host, port, password and db are used for connecting to Redis)
func NewClient(host string, port int, password string, db int) service.Leaderboard {
	database := database.NewRedisDatabase(database.RedisOptions{
		ClusterEnabled: false,
		Host:           host,
		Port:           port,
		Password:       password,
		DB:             db,
	})

	service := service.NewService(database)
	return service
}

// NewClientFromExistingRedis creates a client using an existing redis connection
func NewClientFromExistingRedis(redis redis.Redis) service.Leaderboard {
	database := &database.Redis{redis}
	return service.NewService(database)
}

// NewClusterClient creates a leaderboard prepared to receive commands and execute them in a redis cluster
func NewClusterClient(addrs []string, password string) service.Leaderboard {
	database := database.NewRedisDatabase(database.RedisOptions{
		ClusterEnabled: true,
		Addrs:          addrs,
		Password:       password,
	})

	service := service.NewService(database)
	return service
}
