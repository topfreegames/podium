package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/model"
)

const (
	getAroundScoreServiceLabel = "get around score"
	offset                     = 0
	limit                      = 1
)

// GetAroundScore find members around an score
func (s *Service) GetAroundScore(ctx context.Context, leaderboard string, pageSize int, score int64, order string) ([]*model.Member, error) {
	// memberID, err := s.getMemberIDWithClosestScore(ctx, leaderboard, score)
	// if err != nil {
	// }

	// return c.getAroundMe(redisClient, leaderboardID, pageSize, memberID, order, true)
	return nil, nil
}

// func (s *Service) getMemberIDWithClosestScore(ctx context.Context, leaderboard string, score int64) (string, error) {
// 	memberSlice, err := s.Database.GetMemberIDsWithScoreInsideRange(ctx, leaderboard, "-inf", fmt.Sprint(score), 0, 1)
// 	if err != nil {
// 		return "", NewGeneralError(getAroundScoreServiceLabel, err.Error())
// 	}
// 	if len(memberSlice) == 0 {
// 		return "", NewMemberNotFoundError(leaderboard, fmt.Sprint(score))
// 	}
//
// 	return memberSlice[0], nil
// }
