package service

import (
	"context"
	"strconv"

	"github.com/topfreegames/podium/leaderboard/v2/model"
)

const (
	getAroundScoreServiceLabel = "get around score"
	offset                     = 0
	limit                      = 1
)

// GetAroundScore find members around an score
func (s *Service) GetAroundScore(ctx context.Context, leaderboard string, pageSize int, score int64, order string) ([]*model.Member, error) {
	member, err := s.getMemberIDWithClosestScore(ctx, leaderboard, score)
	if err != nil {
		return nil, err
	}

	memberRank, err := s.fetchMemberRank(ctx, leaderboard, member, order, true)
	if err != nil {
		return nil, NewGeneralError(getAroundScoreServiceLabel, err.Error())
	}

	indexes, err := s.calculateIndexesAroundMemberRank(ctx, leaderboard, memberRank, pageSize)
	if err != nil {
		return nil, NewGeneralError(getAroundScoreServiceLabel, err.Error())
	}

	databaseMembers, err := s.Database.GetOrderedMembers(ctx, leaderboard, indexes.Start, indexes.Stop, order)
	if err != nil {
		return nil, NewGeneralError(getAroundScoreServiceLabel, err.Error())
	}

	members := convertDatabaseMembersIntoModelMembers(databaseMembers)

	return members, nil
}

func (s *Service) getMemberIDWithClosestScore(ctx context.Context, leaderboard string, score int64) (string, error) {
	memberSlice, err := s.Database.GetMemberIDsWithScoreInsideRange(ctx, leaderboard, "-inf", strconv.FormatInt(score, 10), 0, 1)
	if err != nil {
		return "", NewGeneralError(getAroundScoreServiceLabel, err.Error())
	}
	if len(memberSlice) == 0 {
		return "", nil
	}

	return memberSlice[0], nil
}
