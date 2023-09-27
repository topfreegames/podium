package cache

import (
	"context"
	"github.com/topfreegames/podium/leaderboard/v2/enriching"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"go.uber.org/zap"
)

type cachedEnricher struct {
	cache  enriching.EnricherCache
	config cacheConfig
	impl   enriching.Enricher
	logger *zap.Logger
}

func NewCachedEnricher(
	cache enriching.EnricherCache,
	enricher enriching.Enricher,
	options ...CachedEnricherOptions,
) enriching.Enricher {
	e := &cachedEnricher{
		cache:  cache,
		impl:   enricher,
		config: newDefaultCacheConfig(),
		logger: zap.NewNop(),
	}

	for _, opt := range options {
		opt(e)
	}

	return e
}

var _ enriching.Enricher = &cachedEnricher{}

func (e *cachedEnricher) Enrich(
	ctx context.Context,
	tenantID,
	leaderboardID string,
	members []*model.Member,
) ([]*model.Member, error) {
	if len(members) == 0 {
		return members, nil
	}

	l := e.logger.With(
		zap.String("tenantID", tenantID),
		zap.String("leaderboardID", leaderboardID),
	)

	cached, hit, err := e.cache.Get(ctx, tenantID, members)
	if err != nil {
		l.Error("could not get cached enrichment data", zap.Error(err))
	}

	if hit {
		l.Debug("returning cached enrich data")
		for _, m := range members {
			if metadata, exists := cached[m.PublicID]; exists {
				m.Metadata = metadata
			}
		}

		return members, nil
	}

	members, err = e.impl.Enrich(ctx, tenantID, leaderboardID, members)

	if err != nil {
		return nil, err
	}

	err = e.cache.Set(ctx, tenantID, members, e.config.ttl)
	if err != nil {
		l.Error("could not set cached enrichment data", zap.Error(err))
	}

	return members, nil
}
