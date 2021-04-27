package service

import "context"

const removeMemberServiceLabel = "remove member"

// RemoveMember dele specific member from leaderboard
func (s *Service) RemoveMember(ctx context.Context, leaderboard, member string) error {
	err := s.Database.RemoveMembers(ctx, leaderboard, member)
	if err != nil {
		return NewGeneralError(removeMemberServiceLabel, err.Error())
	}
	return nil
}
