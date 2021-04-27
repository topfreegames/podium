package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/model"
)

func (s *Service) GetTopPercentage(ctx context.Context, leaderboard string, pageSize, amount, maxMembers int, order string) ([]*model.Member, error) {
	return nil, nil
}
