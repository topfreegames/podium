package service

import (
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
	}
}
