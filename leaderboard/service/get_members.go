package service

import (
	"context"
	"sort"
	"time"

	"github.com/topfreegames/podium/leaderboard/v2/model"
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
			continue
		}

		var ttl int64
		if (member.TTL != time.Time{}) {
			ttl = member.TTL.Unix()
		}
		newMember := &model.Member{
			PublicID: member.Member,
			Score:    int64(member.Score),
			Rank:     int(member.Rank) + 1,
			ExpireAt: int(ttl),
		}
		membersToReturn = append(membersToReturn, newMember)

	}

	sort.SliceStable(membersToReturn, func(i, j int) bool { return membersToReturn[i].Rank < membersToReturn[j].Rank })

	return membersToReturn, nil
}
