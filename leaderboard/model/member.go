package model

// Member maps an member identified by their publicID to their score and rank
type Member struct {
	PublicID     string `json:"publicID"`
	Score        int64  `json:"score"`
	Rank         int    `json:"rank"`
	PreviousRank int    `json:"previousRank"`
	ExpireAt     int    `json:"expireAt"`
}
