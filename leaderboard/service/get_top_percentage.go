package service

import (
	"context"
	"math"

	"github.com/topfreegames/podium/leaderboard/v2/model"
)

const getTopPercentageServiceLabel = "get top percentage"

// GetTopPercentage retrieves top x% members from the leaderboard.
func (s *Service) GetTopPercentage(ctx context.Context, leaderboardID string, pageSize, amount, maxMembers int, order string) ([]*model.Member, error) {
	if amount < 1 || amount > 100 {
		return nil, NewPercentageError(amount)
	}

	if order != "desc" && order != "asc" {
		order = "desc"
	}

	amountInPercentage := float64(amount) / 100.0
	totalNumberMembers, err := s.Database.GetTotalMembers(ctx, leaderboardID)
	if err != nil {
		return nil, NewGeneralError(getTopPercentageServiceLabel, err.Error())
	}

	numberMembersToReturn := int(math.Floor(float64(totalNumberMembers) * amountInPercentage))

	if numberMembersToReturn < 1 {
		numberMembersToReturn = 1
	}

	if numberMembersToReturn > maxMembers {
		numberMembersToReturn = maxMembers
	}

	databaseMembers, err := s.Database.GetOrderedMembers(ctx, leaderboardID, 0, numberMembersToReturn-1, order)
	if err != nil {
		return nil, NewGeneralError(getTopPercentageServiceLabel, err.Error())
	}

	members := convertDatabaseMembersIntoModelMembers(databaseMembers)
	return members, nil
}
