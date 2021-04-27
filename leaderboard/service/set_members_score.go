package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/model"
)

func (s *Service) SetMembersScore(ctx context.Context, leaderboardID string, members []*model.Member, prevRank bool, scoreTTL string) error {
	return nil
}
