package service

import (
	"context"
	"fmt"

	"github.com/topfreegames/podium/leaderboard/database"
	"github.com/topfreegames/podium/leaderboard/model"
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

	memberRank, err := s.fetchMemberRank(ctx, leaderboard, member, order, false)
	if err != nil {
		if _, ok := err.(*database.MemberNotFoundError); ok {
			return nil, NewMemberNotFoundError(leaderboard, member)
		}

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
	memberSlice, err := s.Database.GetMemberIDsWithScoreInsideRange(ctx, leaderboard, "-inf", fmt.Sprint(score), 0, 1)
	if err != nil {
		return "", NewGeneralError(getAroundScoreServiceLabel, err.Error())
	}
	if len(memberSlice) == 0 {
		return "", NewMemberNotFoundError(leaderboard, fmt.Sprint(score))
	}

	return memberSlice[0], nil
}
