package database

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/topfreegames/podium/leaderboard/v2/database/redis"
)

var _ Expiration = &Redis{}

// GetExpirationLeaderboards return leaderboards registerd with members to expire
func (r *Redis) GetExpirationLeaderboards(ctx context.Context) ([]string, error) {
	expirationKeys, err := r.Client.SMembers(ctx, ExpirationSet)
	if err != nil {
		return nil, NewGeneralError(err.Error())
	}

	expirationLeaderboards := make([]string, 0, len(expirationKeys))
	for _, expirationKey := range expirationKeys {
		expirationLeaderboards = append(expirationLeaderboards, strings.TrimSuffix(expirationKey, ":ttl"))
	}

	return expirationLeaderboards, nil
}

// GetMembersToExpire get members in the leaderboard to expire
func (r *Redis) GetMembersToExpire(ctx context.Context, leaderboard string, amount int, maxTime time.Time) ([]string, error) {
	expirationSet := fmt.Sprintf("%s:ttl", leaderboard)

	err := r.Client.Exists(ctx, expirationSet)
	if err != nil {
		if _, ok := err.(*redis.KeyNotFoundError); ok {
			return nil, NewLeaderboardWithoutMemberToExpireError(leaderboard)
		}
		return nil, NewGeneralError(err.Error())
	}

	unixTimestamp := maxTime.Unix()

	members, err := r.Client.ZRangeByScore(ctx, expirationSet, "-inf", strconv.FormatInt(unixTimestamp, 10), 0, int64(amount))
	if err != nil {
		return nil, NewGeneralError(err.Error())
	}

	return members, nil
}

// RemoveLeaderboardFromExpireList remove from leaderboard expiration list the leaderboard
func (r *Redis) RemoveLeaderboardFromExpireList(ctx context.Context, leaderboard string) error {
	leaderboardExpirationKey := fmt.Sprintf("%s:ttl", leaderboard)

	err := r.Client.SRem(ctx, ExpirationSet, leaderboardExpirationKey)
	if err != nil {
		return NewGeneralError(err.Error())
	}

	return nil
}

// ExpireMembers remove members from leaderboard
func (r *Redis) ExpireMembers(ctx context.Context, leaderboard string, members []string) error {
	leaderboardExpirationKey := fmt.Sprintf("%s:ttl", leaderboard)

	err := r.Client.ZRem(ctx, leaderboard, members...)
	if err != nil {
		return NewGeneralError(err.Error())
	}

	err = r.Client.ZRem(ctx, leaderboardExpirationKey, members...)
	if err != nil {
		return NewGeneralError(err.Error())
	}

	return nil
}
