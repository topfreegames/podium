package service

import (
	"context"
	"math"
)

const totalPagesServiceLabel = "total pages"

// TotalPages return how many pages members have in a leaderboard according to pageSize
func (s *Service) TotalPages(ctx context.Context, leaderboard string, pageSize int) (int, error) {
	totalMembers, err := s.Database.GetTotalMembers(ctx, leaderboard)
	if err != nil {
		return -1, NewGeneralError(totalPagesServiceLabel, err.Error())
	}

	pages := int(math.Ceil(float64(totalMembers) / float64(pageSize)))
	return pages, nil
}
