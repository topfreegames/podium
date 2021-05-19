package database

import (
	"context"
	"time"
)

// Expiration interface standardize expiration database calls
type Expiration interface {
	GetExpirationLeaderboards(ctx context.Context) ([]string, error)
	GetMembersToExpire(ctx context.Context, leaderboard string, amount int, maxTime time.Time) ([]string, error)
	RemoveLeaderboardFromExpireList(ctx context.Context, leaderboard string) error
	ExpireMembers(ctx context.Context, leaderboard string, members []string) error
}
