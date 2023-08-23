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

type EnricherCache interface {
	// Get checks the cache for the enrichment metadata of the members. Returns true if all the members are found.
	Get(ctx context.Context, tenantID, leaderboardID string, members []*model.Member) (map[string]map[string]string, bool, error)

	// Set sets the enrichment metadata of the members in the cache.
	Set(ctx context.Context, tenantID, leaderboardID string, members []*model.Member) error
}
