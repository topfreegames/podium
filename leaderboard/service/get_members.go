package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/model"
)

func (s *Service) GetMembers(ctx context.Context, leaderboard string, members []string, order string, includeTTL bool) ([]*model.Member, error) {
	return nil, nil
}
