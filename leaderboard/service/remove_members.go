package service

import "context"

const removeMembersServiceLabel = "remove members"

// RemoveMembers reurn how many members have in a leaderboard
func (s *Service) RemoveMembers(ctx context.Context, leaderboard string, members []string) error {
	err := s.Database.RemoveMembers(ctx, leaderboard, members...)
	if err != nil {
		return NewGeneralError(removeMembersServiceLabel, err.Error())
	}
	return nil
}
