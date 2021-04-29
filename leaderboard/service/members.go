package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/database"
	"github.com/topfreegames/podium/leaderboard/model"
)

func convertDatabaseMembersIntoModelMembers(databaseMembers []*database.Member, offset int) []*model.Member {
	members := make([]*model.Member, 0, len(databaseMembers))
	for i, member := range databaseMembers {
		modelMember := convertDatabaseMemberIntoModelMember(member)
		modelMember.Rank = int(offset + i + 1)
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

	return memberRank, nil
}
