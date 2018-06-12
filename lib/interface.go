package lib

// PodiumInterface defines the interface to be implemented
type PodiumInterface interface {
	GetTop(leaderboard string, page int, pageSize int) (int, *MemberList, error)
	GetTopPercent(leaderboard string, percentage int) (int, *MemberList, error)
	UpdateScore(leaderboard string, memberID string, score int) (int, *MemberList, error)
	IncrementScore(leaderboard string, memberID string, increment int) (int, *MemberList, error)
	UpdateScores(leaderboards []string, memberID string, score int) (int, *ScoreList, error)
	RemoveMemberFromLeaderboard(leaderboard string, member string) (int, *Response, error)
	GetMember(leaderboard string, memberID string) (int, *Member, error)
	GetMembers(leaderboard string, memberIDs []string) (int, *MemberList, error)
	Healthcheck() (int, string, error)
	DeleteLeaderboard(leaderboard string) (int, *Response, error)
}
