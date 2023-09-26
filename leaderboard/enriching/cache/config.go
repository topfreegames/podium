package cache

import (
	"github.com/topfreegames/podium/leaderboard/v2/enriching"
	"go.uber.org/zap"
	"time"
)

type cacheConfig struct {
	// TTL is the time to live for the cached data.
	ttl time.Duration
}

func newDefaultCacheConfig() cacheConfig {
	return cacheConfig{
		ttl: 24 * time.Hour,
	}
}

type CachedEnricherOptions func(enriching.Enricher)

// WithLogger sets the logger for the cached enricher.
func WithLogger(logger *zap.Logger) CachedEnricherOptions {
	return func(e enriching.Enricher) {
		impl := e.(*cachedEnricher)
		impl.logger = logger.With(zap.String("source", "cached_enricher"))
	}
}

func WithTTL(ttl time.Duration) CachedEnricherOptions {
	return func(e enriching.Enricher) {
		impl := e.(*cachedEnricher)
		impl.config.ttl = ttl
	}
}
