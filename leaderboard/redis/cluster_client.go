package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type clusterClient struct {
	*goredis.ClusterClient
}

type ClusterOptions struct {
	Hosts    []string
	Password string
}

// NewClusterClient returns a new redis instance
func NewClusterClient(clusterOptions ClusterOptions) *clusterClient {
	goRedisClient := goredis.NewClusterClient(&goredis.ClusterOptions{
		Addrs:    clusterOptions.Hosts,
		Password: clusterOptions.Password,
	})

	return &clusterClient{goRedisClient}
}

// ExpireAt call redis EXPIREAT function
func (cc *clusterClient) ExpireAt(ctx context.Context, key string, time time.Time) error {
	result, err := cc.ClusterClient.ExpireAt(ctx, key, time).Result()
	if err != nil {
		return err
	}

	if result != true {
		return NewKeyNotFoundError(key)
	}

	return nil
}

// Ping call redis PING function
func (cc *clusterClient) Ping(ctx context.Context) error {
	err := cc.ClusterClient.Ping(ctx).Err()
	return err
}

// SAdd call redis SADD function
func (cc *clusterClient) SAdd(ctx context.Context, key, member string) error {
	_, err := cc.ClusterClient.SAdd(ctx, key, member).Result()
	return err
}

// SRem call redis SREM function
func (cc *clusterClient) SRem(ctx context.Context, key, member string) error {
	err := cc.ClusterClient.SRem(ctx, key, member).Err()
	return err
}

// TTL call redis TTL function
func (cc *clusterClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	result, err := cc.ClusterClient.TTL(ctx, key).Result()
	if err != nil {
		return -1, err
	}

	if result == -2 {
		return -1, NewKeyNotFoundError(key)
	}

	if result == -1 {
		return -1, NewTTLNotFoundError(key)
	}

	return result, nil
}

// ZAdd call redis ZADD function
func (cc *clusterClient) ZAdd(ctx context.Context, key, member string, score float64) error {
	_, err := cc.ClusterClient.ZAdd(ctx, key, &goredis.Z{Score: score, Member: member}).Result()
	return err
}

// ZCard call redis ZCARD function
func (cc *clusterClient) ZCard(ctx context.Context, key string) (int64, error) {
	result, err := cc.ClusterClient.ZCard(ctx, key).Result()
	if err != nil {
		return -1, err
	}

	if result == 0 {
		return -1, NewKeyNotFoundError(key)
	}

	return result, nil
}

// ZIncrBy call redis ZINCRBY function
func (cc *clusterClient) ZIncrBy(ctx context.Context, key, member string, increment float64) error {
	_, err := cc.ClusterClient.ZIncrBy(ctx, key, increment, member).Result()
	return err
}

// ZRange call redis ZRANGE function it is inclusive it returns start and stop element
func (cc *clusterClient) ZRange(ctx context.Context, key string, start, stop int64) ([]*Member, error) {
	result, err := cc.ClusterClient.ZRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return []*Member{}, err
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

// ZRank call redis ZRANK function
func (cc *clusterClient) ZRank(ctx context.Context, key, member string) (int64, error) {
	result, err := cc.ClusterClient.ZRank(ctx, key, member).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, NewMemberNotFoundError(key)
		}

		return 0, err
	}

	return result, nil
}

// ZRem call redis ZREM function
func (cc *clusterClient) ZRem(ctx context.Context, key, member string) error {
	err := cc.ClusterClient.ZRem(ctx, key, member).Err()
	return err
}

// ZRevRange call redis ZREVRANGE function it is inclusive it returns start and stop element
func (cc *clusterClient) ZRevRange(ctx context.Context, key string, start, stop int64) ([]*Member, error) {
	result, err := cc.ClusterClient.ZRevRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return []*Member{}, err
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

// ZRevRank call redis ZRevRank function
func (cc *clusterClient) ZRevRank(ctx context.Context, key, member string) (int64, error) {
	result, err := cc.ClusterClient.ZRevRank(ctx, key, member).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, NewMemberNotFoundError(key)
		}

		return 0, err
	}

	return result, nil
}

// ZScore call redis ZScore function
func (cc *clusterClient) ZScore(ctx context.Context, key, member string) (float64, error) {
	result, err := cc.ClusterClient.ZScore(ctx, key, member).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, NewMemberNotFoundError(key)
		}

		return 0, err
	}

	return result, nil
}
