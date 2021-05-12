package service

import (
	"context"
	"time"

	"github.com/topfreegames/podium/leaderboard/database"
	"github.com/topfreegames/podium/leaderboard/expiration"
)

func (s *Service) persistLeaderboardExpirationTime(ctx context.Context, leaderboard string) error {
	expireAt, err := expiration.GetExpireAt(leaderboard)
	if err != nil {
		return err
	}

	if expireAt == -1 {
		return nil
	}

	_, err = s.Database.GetLeaderboardExpiration(ctx, leaderboard)
	if err != nil {
		if _, ok := err.(*database.TTLNotFoundError); ok {
			err = s.Database.SetLeaderboardExpiration(ctx, leaderboard, time.Unix(expireAt, 0))
			if err != nil {
				return err
			}

		} else {
			return err
		}
	}

	return nil
}
