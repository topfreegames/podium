package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/v2/database"
)

const getRankServiceLabel = "get rank"

// GetRank return the current member rank in a specific order
func (s *Service) GetRank(ctx context.Context, leaderboard, member, order string) (int, error) {
	rank, err := s.Database.GetRank(ctx, leaderboard, member, order)
	if err != nil {
		if _, ok := err.(*database.MemberNotFoundError); ok {
			return -1, NewMemberNotFoundError(leaderboard, member)
		}

		return -1, NewGeneralError(getRankServiceLabel, err.Error())
	}

	return int(rank + 1), nil
}
