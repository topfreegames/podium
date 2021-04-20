package redis

import (
	"context"
	"time"
)

const (
	// TTLKeyNotFound is redis return status to TTL command that simbolize a key not found
	TTLKeyNotFound = -2
	// KeyWithoutTTL is redis return status to TTL command that simbolize a key without TTL set
	KeyWithoutTTL = -1
)

// Redis interface define wich redis methods will be used by leaderboard module
type Redis interface {
	ExpireAt(ctx context.Context, key string, time time.Time) error
	Ping(ctx context.Context) error
	SAdd(ctx context.Context, key, member string) error
	SRem(ctx context.Context, key, member string) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	ZAdd(ctx context.Context, key, member string, score float64) error
	ZCard(ctx context.Context, key string) (int64, error)
	ZIncrBy(ctx context.Context, key, member string, increment float64) error
	ZRange(ctx context.Context, key string, start, stop int64) ([]*Member, error)
	ZRank(ctx context.Context, key, member string) (int64, error)
	ZRem(ctx context.Context, key string, member string) error
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]*Member, error)
	ZRevRank(ctx context.Context, key, member string) (int64, error)
	ZScore(ctx context.Context, key, member string) (float64, error)
}

// Member is a struct to be used by sorted set range operations
type Member struct {
	Member string
	Score  float64
}
