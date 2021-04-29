package service

import "context"

const removeMembersServiceLabel = "remove members"

// RemoveMembers remove members from a certain leaderboard
func (s *Service) RemoveMembers(ctx context.Context, leaderboard string, members []string) error {
	err := s.Database.RemoveMembers(ctx, leaderboard, members...)
	if err != nil {
		return NewGeneralError(removeMembersServiceLabel, err.Error())
	}
	return nil
}
