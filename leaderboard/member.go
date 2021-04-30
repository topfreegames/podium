package leaderboard

import "github.com/topfreegames/podium/leaderboard/model"

// Member maps an member identified by their publicID to their score and rank
type Member struct {
	PublicID     string `json:"publicID"`
	Score        int64  `json:"score"`
	Rank         int    `json:"rank"`
	PreviousRank int    `json:"previousRank"`
	ExpireAt     int    `json:"expireAt"`
}

//Members are a list of member
type Members []*Member

func (slice Members) Len() int {
	return len(slice)
}

func (slice Members) Less(i, j int) bool {
	return slice[i].Rank < slice[j].Rank
}

func (slice Members) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func convertModelsToMembers(models []*model.Member) []*Member {
	members := make([]*Member, 0, len(models))
	for _, model := range models {
		members = append(members, convertModelToMember(model))
	}

	return members
}

func convertModelToMember(model *model.Member) *Member {
	return &Member{
		PublicID:     model.PublicID,
		Score:        model.Score,
		Rank:         model.Rank,
		PreviousRank: model.PreviousRank,
		ExpireAt:     model.ExpireAt,
	}
}
