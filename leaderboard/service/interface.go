package service

import (
	"context"

	"github.com/topfreegames/podium/leaderboard/v2/model"
)

// Leaderboard holds functional struct inside leaderboard and your dependecies to be executed
type Leaderboard interface {
	Healthcheck(ctx context.Context) error

	IncrementMemberScore(ctx context.Context, leaderboard string, member string, increment int, scoreTTL string) (*model.Member, error)
	SetMemberScore(ctx context.Context, leaderboard, member string, score int64, prevRank bool, scoreTTL string) (*model.Member, error)
	SetMembersScore(ctx context.Context, leaderboard string, members []*model.Member, prevRank bool, scoreTTL string) error

	RemoveLeaderboard(ctx context.Context, leaderboard string) error
	RemoveMember(ctx context.Context, leaderboard, member string) error
	RemoveMembers(ctx context.Context, leaderboard string, members []string) error

	GetMember(ctx context.Context, leaderboard, member string, order string, includeTTL bool) (*model.Member, error)
	GetMembers(ctx context.Context, leaderboard string, members []string, order string, includeTTL bool) ([]*model.Member, error)
	GetMembersByRange(ctx context.Context, leaderboard string, start int, stop int, order string) ([]*model.Member, error)
	GetRank(ctx context.Context, leaderboard, member, order string) (int, error)

	TotalMembers(ctx context.Context, leaderboard string) (int, error)
	TotalPages(ctx context.Context, leaderboard string, pageSize int) (int, error)

	GetLeaders(ctx context.Context, leaderboard string, pageSize, page int, order string) ([]*model.Member, error)
	GetTopPercentage(ctx context.Context, leaderboard string, pageSize, amount, maxMembers int, order string) ([]*model.Member, error)

	GetAroundMe(ctx context.Context, leaderboard string, pageSize int, member string, order string, getLastIfNotFound bool) ([]*model.Member, error)

	GetAroundScore(ctx context.Context, leaderboard string, pageSize int, score int64, order string) ([]*model.Member, error)
}
