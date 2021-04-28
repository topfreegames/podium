package service

import "context"

const removeLeaderboardServiceLabel = "remove leaderboard"

// RemoveLeaderboard reurn how many members have in a leaderboard
func (s *Service) RemoveLeaderboard(ctx context.Context, leaderboard string) error {
	err := s.Database.RemoveLeaderboard(ctx, leaderboard)
	if err != nil {
		return NewGeneralError(removeLeaderboardServiceLabel, err.Error())
	}
	return nil
}
