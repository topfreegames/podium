package service

import (
	"context"
	"strconv"
	"time"

	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/model"
)

func convertDatabaseMembersIntoModelMembers(databaseMembers []*database.Member) []*model.Member {
	members := make([]*model.Member, 0, len(databaseMembers))
	for _, member := range databaseMembers {
		modelMember := convertDatabaseMemberIntoModelMember(member)
		members = append(members, modelMember)
	}

	return members
}

func convertDatabaseMemberIntoModelMember(member *database.Member) *model.Member {
	return &model.Member{
		PublicID: member.Member,
		Score:    int64(member.Score),
		Rank:     int(member.Rank + 1),
	}
}

func (s *Service) fetchMemberRank(ctx context.Context, leaderboard, member, order string, getLastIfNotFound bool) (int, error) {
	memberRank, err := s.Database.GetRank(ctx, leaderboard, member, order)
	if err != nil {
		if _, ok := err.(*database.MemberNotFoundError); ok {
			if !getLastIfNotFound {
				return -1, err
			}

			memberRank, err = s.Database.GetTotalMembers(ctx, leaderboard)
			if err != nil {
				return -1, err
			}
		} else {
			return -1, err
		}
	}

	return memberRank + 1, nil
}

func (s *Service) setMembersPreviousRank(ctx context.Context, leaderboard string, members []*model.Member, order string) error {
	databaseMembers, err := s.getDatabaseMembers(ctx, leaderboard, members, order)
	if err != nil {
		return err
	}

	for i, member := range members {
		if databaseMembers[i] != nil {
			member.PreviousRank = int(databaseMembers[i].Rank + 1)
			continue
		}

		member.PreviousRank = -1
	}

	return nil
}

func (s *Service) persistMembers(ctx context.Context, leaderboard string, members []*model.Member) error {
	databaseMembers := make([]*database.Member, 0, len(members))
	for _, member := range members {
		databaseMembers = append(databaseMembers, &database.Member{
			Member: member.PublicID,
			Score:  float64(member.Score),
		})
	}

	err := s.Database.SetMembers(ctx, leaderboard, databaseMembers)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) setMembersValues(ctx context.Context, leaderboard string, members []*model.Member, order string) error {
	databaseMembers, err := s.getDatabaseMembers(ctx, leaderboard, members, order)
	if err != nil {
		return err
	}

	for i, member := range databaseMembers {
		members[i].Rank = int(member.Rank + 1)
		members[i].Score = int64(member.Score)
	}

	return nil
}

func (s *Service) persistMembersTTL(ctx context.Context, leaderboard string, members []*model.Member, scoreTTL string) error {
	ttl, err := strconv.ParseInt(scoreTTL, 10, 64)
	if err != nil {
		return err
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
		return err
	}

	return nil

}

func (s *Service) getDatabaseMembers(ctx context.Context, leaderboard string, members []*model.Member, order string) ([]*database.Member, error) {
	memberIDs := make([]string, 0, len(members))
	for _, member := range members {
		memberIDs = append(memberIDs, member.PublicID)
	}

	databaseMembers, err := s.Database.GetMembers(ctx, leaderboard, order, true, memberIDs...)
	if err != nil {
		return nil, err
	}

	return databaseMembers, nil
}
