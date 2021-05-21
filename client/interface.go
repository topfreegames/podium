package client

import "context"

// PodiumInterface defines the interface to be implemented
type PodiumInterface interface {
	DeleteLeaderboard(ctx context.Context, leaderboard string) (*Response, error)
	GetCount(ctx context.Context, leaderboard string) (int, error)
	GetMember(ctx context.Context, leaderboard, memberID string) (*Member, error)
	GetMembers(ctx context.Context, leaderboard string, memberIDs []string) (*MemberList, error)
	GetMemberInLeaderboards(ctx context.Context, leaderboards []string, memberID string, order ...string) (*ScoreList, error)
	GetMembersAroundMember(ctx context.Context, leaderboard, memberID string, pageSize int, getLastIfNotFound bool, order ...string) (*MemberList, error)
	GetTop(ctx context.Context, leaderboard string, page, pageSize int) (*MemberList, error)
	GetTopPercent(ctx context.Context, leaderboard string, percentage int) (*MemberList, error)
	Healthcheck(ctx context.Context) (string, error)
	IncrementScore(ctx context.Context, leaderboard, memberID string, increment, scoreTTL int) (*MemberList, error)
	RemoveMemberFromLeaderboard(ctx context.Context, leaderboard, member string) (*Response, error)
	UpdateScore(ctx context.Context, leaderboard, memberID string, score, scoreTTL int) (*Member, error)
	UpdateScores(ctx context.Context, leaderboards []string, memberID string, score, scoreTTL int) (*ScoreList, error)
	UpdateMembersScore(ctx context.Context, leaderboard string, members []*Member, scoreTTL int) (*MemberList, error)
}
