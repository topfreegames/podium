package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/model"
)

const setMembersScoreServiceLabel = "set members score"

const setMembersOrder = "desc"

// SetMembersScore return member informations that is
func (s *Service) SetMembersScore(ctx context.Context, leaderboard string, members []*model.Member, prevRank bool, scoreTTL string) error {
	if prevRank {
		err := s.setMembersPreviousRank(ctx, leaderboard, members, setMembersOrder)
		if err != nil {
			return NewGeneralError(setMembersScoreServiceLabel, err.Error())
		}
	}

	err := s.persistMembers(ctx, leaderboard, members)
	if err != nil {
		return NewGeneralError(setMembersScoreServiceLabel, err.Error())
	}

	err = s.setMembersRank(ctx, leaderboard, members, setMembersOrder)
	if err != nil {
		return NewGeneralError(setMembersScoreServiceLabel, err.Error())
	}

	err = s.persistLeaderboardExpirationTime(ctx, leaderboard)
	if err != nil {
		return NewGeneralError(setMembersScoreServiceLabel, err.Error())
	}

	if scoreTTL != "" {
		err = s.persistMembersTTL(ctx, leaderboard, members, scoreTTL)
		if err != nil {
			return NewGeneralError(setMembersScoreServiceLabel, err.Error())
		}
	}

	return nil
}
