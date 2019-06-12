package leaderboard

import "context"

//Creates an empty context (shortcut for context.Background())
func NewEmptyCtx() context.Context {
	return context.Background()
}
