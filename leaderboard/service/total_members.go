package service

import (
	"context"
)

const totalMembersServiceLabel = "total members"

// TotalMembers reurn how many members have in a leaderboard
func (s *Service) TotalMembers(ctx context.Context, leaderboard string) (int, error) {
	count, err := s.Database.GetTotalMembers(ctx, leaderboard)
	if err != nil {
		return -1, NewGeneralError(totalMembersServiceLabel, err.Error())
	}
	return int(count), nil
}
