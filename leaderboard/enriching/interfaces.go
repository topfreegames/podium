package enriching

import (
	"context"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"time"
)

type Enricher interface {
	// Enrich enriches the members list with some metadata.
	// Returns the members list with the metadata field filled.
	Enrich(ctx context.Context, tenantID, leaderboardID string, members []*model.Member) ([]*model.Member, error)
}

type EnricherCache interface {
	// Get checks the cache for the enrichment metadata of the members.
	// Returns a map with the member's publicID as key and the metadata as value.
	// Returns true if all the members are found, false if one or more are not.
	Get(ctx context.Context, tenantID string, members []*model.Member) (map[string]map[string]string, bool, error)

	// Set sets the enrichment metadata of the members in the cache.
	Set(ctx context.Context, tenantID string, members []*model.Member, ttl time.Duration) error
}
