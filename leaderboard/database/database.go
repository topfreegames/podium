package database

import (
	"context"
	"time"
)

// Database interface standardize database calls
type Database interface {
	GetMemberIDsWithScoreInsideRange(ctx context.Context, leaderboard string, min, max string, offset, count int) ([]string, error)
	GetMembers(ctx context.Context, leaderboard, order string, includeTTL bool, members ...string) ([]*Member, error)
	GetOrderedMembers(ctx context.Context, leaderboard string, start, stop int, order string) ([]*Member, error)
	GetRank(ctx context.Context, leaderboard, member, order string) (int, error)
	GetTotalMembers(ctx context.Context, leaderboard string) (int, error)
	Healthcheck(ctx context.Context) error
	RemoveLeaderboard(ctx context.Context, leaderboard string) error
	RemoveMembers(ctx context.Context, leaderboard string, members ...string) error
}

// Member is a struct to be used by users operations
type Member struct {
	Member string
	Score  float64
	Rank   int64
	TTL    time.Time
}
