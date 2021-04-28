package service

import (
	"context"
	"math"

	"github.com/topfreegames/podium/leaderboard/model"
)

const getLeadersServiceLabel = "get leaders"

// GetLeaders reurn leaders
func (s *Service) GetLeaders(ctx context.Context, leaderboard string, pageSize, page int, order string) ([]*model.Member, error) {
	err := s.ensureValidPage(ctx, leaderboard, pageSize, page)
	if err != nil {
		return nil, err
	}

	index := getIndexesByPage(pageSize, page)

	databaseMembers, err := s.Database.GetOrderedMembers(ctx, leaderboard, index.Start, index.Stop, order)
	if err != nil {
		return nil, NewGeneralError(getLeadersServiceLabel, err.Error())
	}

	members := convertDatabaseMembersIntoModelMembers(databaseMembers)
	return members, nil
}

func (s *Service) ensureValidPage(ctx context.Context, leaderboard string, pageSize int, page int) error {
	totalMembers, err := s.Database.GetTotalMembers(ctx, leaderboard)
	if err != nil {
		return NewGeneralError(getLeadersServiceLabel, err.Error())
	}

	totalPages := int(math.Ceil(float64(totalMembers) / float64(pageSize)))

	if page > totalPages || page < 1 {
		return NewPageOutOfRangeError(page, totalPages)
	}

	return nil
}
