package enriching

import (
	"context"
	"github.com/topfreegames/podium/leaderboard/v2/model"
)

type Enricher interface {
	// Enrich enriches the members list with some metadata, given there's an enrichment configuration for the tenantID.
	// Returns a map with the member's publicID as key and the metadata as value.
	// If there's no enrichment configuration for the tenantID, returns nil.
	Enrich(ctx context.Context, tenantID, leaderboardID string, members []*model.Member) ([]*model.Member, error)
}
