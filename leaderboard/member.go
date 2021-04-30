package leaderboard

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
