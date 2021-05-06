package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/model"
)

const setMemberScoreServiceLabel = "set member score"

const setMemberOrder = "desc"

// SetMemberScore return member informations that is
func (s *Service) SetMemberScore(ctx context.Context, leaderboard, member string, score int64, prevRank bool, scoreTTL string) (*model.Member, error) {
	members := []*model.Member{
		{
			PublicID: member,
			Score:    score,
		},
	}

	if prevRank {
		err := s.setMembersPreviousRank(ctx, leaderboard, members, setMemberOrder)
		if err != nil {
			return nil, NewGeneralError(setMemberScoreServiceLabel, err.Error())
		}
	}

	err := s.persistMembers(ctx, leaderboard, members)
	if err != nil {
		return nil, NewGeneralError(setMemberScoreServiceLabel, err.Error())
	}

	err = s.setMembersValues(ctx, leaderboard, members, setMemberOrder)
	if err != nil {
		return nil, NewGeneralError(setMemberScoreServiceLabel, err.Error())
	}

	err = s.persistLeaderboardExpirationTime(ctx, leaderboard)
	if err != nil {
		return nil, NewGeneralError(setMemberScoreServiceLabel, err.Error())
	}

	if scoreTTL != "" {
		err = s.persistMembersTTL(ctx, leaderboard, members, scoreTTL)
		if err != nil {
			return nil, NewGeneralError(setMemberScoreServiceLabel, err.Error())
		}
	}

	return members[0], nil
}
