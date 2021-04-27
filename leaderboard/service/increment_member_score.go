package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/model"
)

func (s *Service) IncrementMemberScore(ctx context.Context, leaderboard string, member string, increment int, scoreTTL string) (*model.Member, error) {
	return nil, nil
}
