package leaderboard

import (
	"context"
	"fmt"
	"net/url"

	"github.com/spf13/viper"
	tfgredis "github.com/topfreegames/extensions/redis"
	"github.com/topfreegames/extensions/redis/interfaces"
)

// Client represents the leaderboard manager object. Capable of managing multiple leaderboards.
type Client struct {
	redisClient *tfgredis.Client
}

func (c *Client) redisWithTracing(ctx context.Context) interfaces.RedisClient {
	return c.redisClient.Trace(ctx)
}

// NewClient creates a leaderboard prepared to receive commands (host, port, password, db and connectionTimeout are used for connecting to Redis)
func NewClient(host string, port int, password string, db int, connectionTimeout int) (*Client, error) {
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

	return &Client{redisClient: cli}, nil
}
