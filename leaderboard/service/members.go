package service

import (
	"github.com/topfreegames/podium/leaderboard/database"
	"github.com/topfreegames/podium/leaderboard/model"
)

func convertDatabaseMembersIntoModelMembers(databaseMembers []*database.Member) []*model.Member {
	members := make([]*model.Member, 0, len(databaseMembers))
	for _, member := range databaseMembers {
		members = append(members, &model.Member{
			PublicID: member.Member,
			Score:    int64(member.Score),
			Rank:     int(member.Rank) + 1,
		})
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
