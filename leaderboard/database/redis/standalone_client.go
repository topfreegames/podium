package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type standaloneClient struct {
	*goredis.Client
}

var _ Client = &standaloneClient{}

// StandaloneOptions define configuration parameters to instantiate a new StandaloneClient
type StandaloneOptions struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewStandaloneClient returns a new redis standalone client instance
func NewStandaloneClient(options StandaloneOptions) Client {
	goRedisClient := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", options.Host, options.Port),
		Password: options.Password,
		DB:       options.DB,
	})

	return &standaloneClient{goRedisClient}
}

// Del call redis DEL function
func (c *standaloneClient) Del(ctx context.Context, key string) error {
	err := c.Client.Del(ctx, key).Err()
	if err != nil {
		return NewGeneralError(err.Error())
	}

	return nil
}

// Exists return if a key exists on redis
func (c *standaloneClient) Exists(ctx context.Context, key string) error {
	value, err := c.Client.Exists(ctx, key).Result()
	if err != nil {
		return NewGeneralError(err.Error())
	}
	if value != 1 {
		return NewKeyNotFoundError(key)
	}

	return nil
}

// ExpireAt call redis EXPIREAT function
func (c *standaloneClient) ExpireAt(ctx context.Context, key string, time time.Time) error {
	result, err := c.Client.ExpireAt(ctx, key, time).Result()
	if err != nil {
		return NewGeneralError(err.Error())
	}

	if result != true {
		return NewKeyNotFoundError(key)
	}
	return nil
}

// Ping call redis PING function
func (c *standaloneClient) Ping(ctx context.Context) (string, error) {
	result, err := c.Client.Ping(ctx).Result()
	if err != nil {
		return "", NewGeneralError(err.Error())
	}
	return result, nil
}

// SAdd call redis SADD function
func (c *standaloneClient) SAdd(ctx context.Context, key, member string) error {
	err := c.Client.SAdd(ctx, key, member).Err()
	if err != nil {
		return NewGeneralError(err.Error())
	}
	return nil
}

// SMembers return all members in a set
func (c *standaloneClient) SMembers(ctx context.Context, key string) ([]string, error) {
	result, err := c.Client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, NewGeneralError(err.Error())
	}
	return result, nil
}

// SRem call redis SREM function
func (c *standaloneClient) SRem(ctx context.Context, key string, members ...string) error {
	err := c.Client.SRem(ctx, key, members).Err()
	if err != nil {
		return NewGeneralError(err.Error())
	}
	return nil
}

// TTL call redis TTL function
func (c *standaloneClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	result, err := c.Client.TTL(ctx, key).Result()
	if err != nil {
		return -1, NewGeneralError(err.Error())
	}

	if result == TTLKeyNotFound {
		return -1, NewKeyNotFoundError(key)
	}

	if result == KeyWithoutTTL {
		return -1, NewTTLNotFoundError(key)
	}

	return result, nil
}

// ZAdd call redis ZADD function
func (c *standaloneClient) ZAdd(ctx context.Context, key string, members ...*Member) error {
	goRedisMembers := make([]*goredis.Z, 0, len(members))
	for _, member := range members {
		goRedisMembers = append(goRedisMembers, &goredis.Z{
			Member: member.Member,
			Score:  member.Score,
		})
	}
	err := c.Client.ZAdd(ctx, key, goRedisMembers...).Err()
	if err != nil {
		return NewGeneralError(err.Error())
	}
	return nil
}

// ZCard call redis ZCARD function
func (c *standaloneClient) ZCard(ctx context.Context, key string) (int64, error) {
	result, err := c.Client.ZCard(ctx, key).Result()
	if err != nil {
		return -1, NewGeneralError(err.Error())
	}

	if result == 0 {
		return -1, NewKeyNotFoundError(key)
	}

	return result, nil
}

// ZIncrBy call redis ZINCRBY function
func (c *standaloneClient) ZIncrBy(ctx context.Context, key, member string, increment float64) error {
	err := c.Client.ZIncrBy(ctx, key, increment, member).Err()
	if err != nil {
		return NewGeneralError(err.Error())
	}
	return nil
}

// ZRange call redis ZRANGE function it is inclusive it returns start and stop element
func (c *standaloneClient) ZRange(ctx context.Context, key string, start, stop int64) ([]*Member, error) {
	result, err := c.Client.ZRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return nil, NewGeneralError(err.Error())
	}

	var members []*Member = make([]*Member, 0, len(result))
	for _, member := range result {
		members = append(members, &Member{
			Member: fmt.Sprint(member.Member),
			Score:  member.Score,
		})
	}

	return members, nil
}

// ZRangeByScore call redis ZRANGEBYSCORE command
func (c *standaloneClient) ZRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error) {
	result, err := c.Client.ZRangeByScore(ctx, key, &goredis.ZRangeBy{Min: min, Max: max, Offset: offset, Count: count}).Result()
	if err != nil {
		return nil, NewGeneralError(err.Error())
	}
	return result, nil
}

// ZRank call redis ZRANK function
func (c *standaloneClient) ZRank(ctx context.Context, key, member string) (int64, error) {
	result, err := c.Client.ZRank(ctx, key, member).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return -1, NewMemberNotFoundError(key, member)
		}

		return -1, NewGeneralError(err.Error())
	}

	return result, nil
}

// ZRem call redis ZREM function
func (c *standaloneClient) ZRem(ctx context.Context, key string, members ...string) error {
	err := c.Client.ZRem(ctx, key, members).Err()
	if err != nil {
		return NewGeneralError(err.Error())
	}
	return nil
}

// ZRevRange call redis ZREVRANGE function it is inclusive it returns start and stop element
func (c *standaloneClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]*Member, error) {
	result, err := c.Client.ZRevRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return nil, NewGeneralError(err.Error())
	}

	var members []*Member = make([]*Member, 0, len(result))
	for _, member := range result {
		members = append(members, &Member{
			Member: fmt.Sprint(member.Member),
			Score:  member.Score,
		})
	}

	return members, nil
}

// ZRevRangeByScore call redis ZREVRANGEBYSCORE command
func (c *standaloneClient) ZRevRangeByScore(ctx context.Context, key string, min, max string, offset, count int64) ([]string, error) {
	result, err := c.Client.ZRevRangeByScore(ctx, key, &goredis.ZRangeBy{Min: min, Max: max, Offset: offset, Count: count}).Result()
	if err != nil {
		return nil, NewGeneralError(err.Error())
	}
	return result, nil
}

// ZRevRank call redis ZREVRANK function
func (c *standaloneClient) ZRevRank(ctx context.Context, key, member string) (int64, error) {
	result, err := c.Client.ZRevRank(ctx, key, member).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return -1, NewMemberNotFoundError(key, member)
		}

		return -1, NewGeneralError(err.Error())
	}

	return result, nil
}

// ZScore call redis ZScore function
func (c *standaloneClient) ZScore(ctx context.Context, key, member string) (float64, error) {
	result, err := c.Client.ZScore(ctx, key, member).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return -1, NewMemberNotFoundError(key, member)
		}

		return -1, NewGeneralError(err.Error())
	}

	return result, nil
}
