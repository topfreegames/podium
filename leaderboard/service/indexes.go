package service

import (
	"context"
)

type index struct {
	Start int
	Stop  int
}

func getIndexesByPage(pageSize, page int) *index {
	start := (page - 1) * pageSize
	stop := start + pageSize - 1

	return &index{Start: start, Stop: stop}
}

func (s *Service) calculateIndexesAroundMemberRank(ctx context.Context, leaderboard string, rank, pageSize int) (*index, error) {
	totalMembers, err := s.Database.GetTotalMembers(ctx, leaderboard)
	if err != nil {
		return nil, err
	}

	start := rank - (pageSize / 2)
	if start < 0 {
		start = 0
	}
	stop := (start + pageSize) - 1
	if stop >= totalMembers {
		stop = totalMembers - 1
		start = stop - pageSize + 1
		if start < 0 {
			start = 0
		}
	}

	return &index{Start: start, Stop: stop}, nil
}
