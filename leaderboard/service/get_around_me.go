package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/model"
)

func (s *Service) GetAroundMe(ctx context.Context, leaderboard string, pageSize int, member string, order string, getLastIfNotFound bool) ([]*model.Member, error) {
	return nil, nil
}
