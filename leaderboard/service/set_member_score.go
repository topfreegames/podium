package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/model"
)

func (s *Service) SetMemberScore(ctx context.Context, leadeboard, member string, score int64, prevRank bool, scoreTTL string) (*model.Member, error) {
	return nil, nil
}
