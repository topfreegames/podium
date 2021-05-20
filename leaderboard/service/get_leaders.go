package service

import (
	"context"
	"math"

	"github.com/topfreegames/podium/leaderboard/v2/model"
)

const getLeadersServiceLabel = "get leaders"

// GetLeaders reurn leaders
func (s *Service) GetLeaders(ctx context.Context, leaderboard string, pageSize, page int, order string) ([]*model.Member, error) {
	page, err := s.ensureValidPage(ctx, leaderboard, pageSize, page)
	if err != nil {
		if _, ok := err.(*PageOutOfRangeError); ok {
			return []*model.Member{}, nil
		}

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

func (s *Service) ensureValidPage(ctx context.Context, leaderboard string, pageSize int, page int) (int, error) {
	totalMembers, err := s.Database.GetTotalMembers(ctx, leaderboard)
	if err != nil {
		return -1, NewGeneralError(getLeadersServiceLabel, err.Error())
	}

	totalPages := int(math.Ceil(float64(totalMembers) / float64(pageSize)))

	if page < 1 {
		return 1, nil
	}

	if page > totalPages {
		return -1, NewPageOutOfRangeError(page, totalPages)
	}

	return page, nil
}
