package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/v2/model"
)

const getMembersByRangeServiceLabel = "get members by range"

// GetMembersByRange reurn how many pages members have in a leaderboard according to pageSize
func (s *Service) GetMembersByRange(ctx context.Context, leaderboard string, start int, stop int, order string) ([]*model.Member, error) {
	databaseMembers, err := s.Database.GetOrderedMembers(ctx, leaderboard, start, stop, order)
	if err != nil {
		return nil, NewGeneralError(getMembersByRangeServiceLabel, err.Error())
	}

	members := convertDatabaseMembersIntoModelMembers(databaseMembers)
	return members, nil
}
