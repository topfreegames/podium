package leaderboard

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	tfgredis "github.com/topfreegames/extensions/redis"
	"github.com/topfreegames/extensions/redis/interfaces"
	"github.com/topfreegames/podium/leaderboard/database"
	"github.com/topfreegames/podium/leaderboard/service"
)

// Client represents the leaderboard manager object. Capable of managing multiple leaderboards.
type Client struct {
	service     *service.Service
	redisClient *tfgredis.Client
}

// NewClient creates a leaderboard prepared to receive commands (host, port, password, db and connectionTimeout are used for connecting to Redis)
func NewClient(host string, port int, password string, db int, connectionTimeout int) (*Client, error) {
	redisInstance := database.NewRedisDatabase(database.RedisOptions{
		ClusterEnabled: false,
		Host:           host,
		Port:           port,
		Password:       password,
		DB:             db,
	})

	service := service.NewService(redisInstance)

	redisURL := url.URL{
		Scheme: "redis",
		User:   url.UserPassword("", password),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   fmt.Sprint(db),
	}

	config := viper.New()
	config.Set("redis.url", redisURL.String())
	config.Set("redis.connectionTimeout", fmt.Sprint(connectionTimeout))

	cli, err := tfgredis.NewClient("redis", config)
	if err != nil {
		return nil, err
	}

	return &Client{service: service, redisClient: cli}, nil
}

// NewClusterClient creates a leaderboard prepared to receive commands and execute them in a redis cluster
func NewClusterClient(host string, port int, password string, db int, connectionTimeout int) (*Client, error) {
	redisInstance := database.NewRedisDatabase(database.RedisOptions{
		ClusterEnabled: true,
		Host:           host,
		Port:           port,
		Password:       password,
		DB:             db,
	})

	service := service.NewService(redisInstance)

	return &Client{service: service}, nil
}

//NewClientWithRedis creates a leaderboard using an already connected tfg Redis
func NewClientWithRedis(cli *tfgredis.Client) (*Client, error) {
	addr := strings.Split(cli.Options.Addr, ":")
	host := addr[0]
	port, err := strconv.Atoi(addr[1])
	if err != nil {
		return nil, fmt.Errorf("Could not parse address port. Addr: %s. %w", cli.Options.Addr, err)
	}

	password := cli.Options.Password
	db := cli.Options.DB
	timeoutInMS := int(cli.Options.DialTimeout / time.Millisecond)

	return NewClient(host, port, password, db, timeoutInMS)
}

func (c *Client) redisWithTracing(ctx context.Context) interfaces.RedisClient {
	return c.redisClient.Trace(ctx)
}
