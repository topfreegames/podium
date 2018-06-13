package lib

// PodiumInterface defines the interface to be implemented
type PodiumInterface interface {
	GetTop(leaderboard string, page int, pageSize int) (*MemberList, error)
	GetTopPercent(leaderboard string, percentage int) (*MemberList, error)
	UpdateScore(leaderboard string, memberID string, score int) (*MemberList, error)
	IncrementScore(leaderboard string, memberID string, increment int) (*MemberList, error)
	UpdateScores(leaderboards []string, memberID string, score int) (*ScoreList, error)
	RemoveMemberFromLeaderboard(leaderboard string, member string) (*Response, error)
	GetMember(leaderboard string, memberID string) (*Member, error)
	GetMembers(leaderboard string, memberIDs []string) (*MemberList, error)
	Healthcheck() (string, error)
	DeleteLeaderboard(leaderboard string) (*Response, error)
}
