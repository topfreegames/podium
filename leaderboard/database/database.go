package database

import (
	"context"
	"time"
)

// Database interface standardize database calls
type Database interface {
	GetLeaderboardExpiration(ctx context.Context, leaderboard string) (int64, error)
	GetMemberIDsWithScoreInsideRange(ctx context.Context, leaderboard string, min, max string, offset, count int) ([]string, error)
	GetMembers(ctx context.Context, leaderboard, order string, includeTTL bool, members ...string) ([]*Member, error)
	GetOrderedMembers(ctx context.Context, leaderboard string, start, stop int, order string) ([]*Member, error)
	GetRank(ctx context.Context, leaderboard, member, order string) (int, error)
	GetTotalMembers(ctx context.Context, leaderboard string) (int, error)
	Healthcheck(ctx context.Context) error
	IncrementMemberScore(ctx context.Context, leaderboard, member string, increment float64) error
	RemoveLeaderboard(ctx context.Context, leaderboard string) error
	RemoveMembers(ctx context.Context, leaderboard string, members ...string) error
	SetLeaderboardExpiration(ctx context.Context, leaderboard string, expireAt time.Time) error
	SetMembers(ctx context.Context, leaderboard string, databaseMembers []*Member) error
	SetMembersTTL(ctx context.Context, leaderboard string, databaseMembers []*Member) error
}

// Member is a struct to be used by users operations
type Member struct {
	Member string
	Score  float64
	Rank   int64
	TTL    time.Time
}
