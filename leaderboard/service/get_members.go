package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/model"
)

const getMembersServiceLabel = "get members"

// GetMembers return member informations that is
func (s *Service) GetMembers(ctx context.Context, leaderboard string, members []string, order string, includeTTL bool) ([]*model.Member, error) {
	databaseMembers, err := s.Database.GetMembers(ctx, leaderboard, order, includeTTL, members...)
	if err != nil {
		return nil, NewGeneralError(getMembersServiceLabel, err.Error())
	}

	membersToReturn := make([]*model.Member, 0, len(databaseMembers))
	for _, member := range databaseMembers {
		if member == nil {
			membersToReturn = append(membersToReturn, nil)
			continue
		}
		newMember := &model.Member{
			PublicID: member.Member,
			Score:    int64(member.Score),
			Rank:     int(member.Rank) + 1,
			ExpireAt: int(member.TTL.Unix()),
		}
		membersToReturn = append(membersToReturn, newMember)

	}

	return membersToReturn, nil
}
