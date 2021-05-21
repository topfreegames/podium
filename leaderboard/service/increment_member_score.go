package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/v2/expiration"
	"github.com/topfreegames/podium/leaderboard/v2/model"
)

const incrementMemberScoreServiceLabel = "increment member score"

const incrementMemberOrder = "desc"

// IncrementMemberScore return member informations that had you score incremented
func (s *Service) IncrementMemberScore(ctx context.Context, leaderboard string, member string, increment int, scoreTTL string) (*model.Member, error) {
	modelMember := &model.Member{
		PublicID: member,
		Score:    int64(increment),
	}

	err := s.incrementMember(ctx, leaderboard, member, increment)
	if err != nil {
		return nil, NewGeneralError(incrementMemberScoreServiceLabel, err.Error())
	}

	members := []*model.Member{modelMember}

	err = s.setMembersValues(ctx, leaderboard, members, incrementMemberOrder)
	if err != nil {
		return nil, NewGeneralError(incrementMemberScoreServiceLabel, err.Error())
	}

	err = s.persistLeaderboardExpirationTime(ctx, leaderboard)
	if err != nil {
		if _, ok := err.(*expiration.LeaderboardExpiredError); ok {
			return nil, NewLeaderboardExpiredError(leaderboard)
		}
		return nil, NewGeneralError(incrementMemberScoreServiceLabel, err.Error())
	}

	if scoreTTL != "" {
		err = s.persistMembersTTL(ctx, leaderboard, members, scoreTTL)
		if err != nil {
			return nil, NewGeneralError(incrementMemberScoreServiceLabel, err.Error())
		}
	}

	return modelMember, nil
}

func (s *Service) incrementMember(ctx context.Context, leaderboard, member string, increment int) error {
	return s.Database.IncrementMemberScore(ctx, leaderboard, member, float64(increment))
}
