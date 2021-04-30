package service

import (
	"context"
	"strconv"
	"time"

	"github.com/topfreegames/podium/leaderboard/database"
	"github.com/topfreegames/podium/leaderboard/expiration"
	"github.com/topfreegames/podium/leaderboard/model"
)

const setMembersScoreServiceLabel = "set members score"

// SetMembersScore return member informations that is
func (s *Service) SetMembersScore(ctx context.Context, leaderboard string, members []*model.Member, prevRank bool, scoreTTL string) error {
	err := s.setLeaderboardExpirationTime(ctx, leaderboard)
	if err != nil {
		return err
	}

	if prevRank {
		err := s.setMembersPreviousRank(ctx, leaderboard, members, "desc")
		if err != nil {
			return err
		}
	}

	err = s.setMembers(ctx, leaderboard, members)
	if err != nil {
		return err
	}

	if scoreTTL != "" {
		err = s.setMembersTTL(ctx, leaderboard, members, scoreTTL)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) setLeaderboardExpirationTime(ctx context.Context, leaderboard string) error {
	expireAt, err := expiration.GetExpireAt(leaderboard)
	if err != nil {
		return NewGeneralError(setMembersScoreServiceLabel, err.Error())
	}

	_, err = s.Database.GetLeaderboardExpiration(ctx, leaderboard)
	if err != nil {
		if _, ok := err.(*database.TTLNotFoundError); ok {
			err = s.Database.SetLeaderboardExpiration(ctx, leaderboard, time.Unix(expireAt, 0))
			if err != nil {
				return NewGeneralError(setMembersScoreServiceLabel, err.Error())
			}

		} else {
			return NewGeneralError(setMembersScoreServiceLabel, err.Error())
		}
	}

	return nil
}

func (s *Service) setMembersPreviousRank(ctx context.Context, leaderboard string, members []*model.Member, order string) error {
	memberIDs := make([]string, 0, len(members))
	for _, member := range members {
		memberIDs = append(memberIDs, member.PublicID)
	}

	databaseMembers, err := s.Database.GetMembers(ctx, leaderboard, order, true, memberIDs...)
	if err != nil {
		return NewGeneralError(getMembersServiceLabel, err.Error())
	}

	for i, member := range members {
		if databaseMembers[i] != nil {
			member.PreviousRank = int(databaseMembers[i].Rank)
		}
	}

	return nil
}

func (s *Service) setMembers(ctx context.Context, leaderboard string, members []*model.Member) error {
	databaseMembers := make([]*database.Member, 0, len(members))
	for _, member := range members {
		databaseMembers = append(databaseMembers, &database.Member{
			Member: member.PublicID,
			Score:  float64(member.Score),
		})
	}

	err := s.Database.SetMembers(ctx, leaderboard, databaseMembers)
	if err != nil {
		return NewGeneralError(setMembersScoreServiceLabel, err.Error())
	}

	return nil
}

func (s *Service) setMembersTTL(ctx context.Context, leaderboard string, members []*model.Member, scoreTTL string) error {
	ttl, err := strconv.ParseInt(scoreTTL, 10, 64)
	if err != nil {
		return NewGeneralError(getMembersServiceLabel, err.Error())
	}

	timeToExpire := time.Now().UTC().Add(time.Duration(ttl) * time.Second)

	databaseMembers := make([]*database.Member, 0, len(members))
	for _, member := range members {
		databaseMembers = append(databaseMembers, &database.Member{
			Member: member.PublicID,
			TTL:    timeToExpire,
		})
		member.ExpireAt = int(timeToExpire.Unix())
	}

	err = s.Database.SetMembersTTL(ctx, leaderboard, databaseMembers)
	if err != nil {
		return NewGeneralError(setMembersScoreServiceLabel, err.Error())
	}

	return nil

}
