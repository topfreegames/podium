package cache

import (
	"context"
	"github.com/opentracing/opentracing-go"
	extensions "github.com/topfreegames/extensions/middleware"
	"github.com/topfreegames/podium/leaderboard/v2/enriching"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"time"
)

const (
	enrichmentCacheGetTimingMilli = "enrichment_cache_get_duration_milliseconds"
	enrichmentCacheGets           = "enrichment_cache_gets"
	enrichmentCacheHits           = "enrichment_cache_gets"
	enrichmentCacheGetErrors      = "enrichment_cache_get_errors"
	enrichmentCacheSetTimingMilli = "enrichment_cache_set_duration_milliseconds"
	enrichmentCacheSets           = "enrichment_cache_sets"
	enrichmentCacheSetErrors      = "enrichment_cache_set_errors"
)

type instrumentedCache struct {
	impl            enriching.EnricherCache
	metricsReporter extensions.MetricsReporter
}

// NewInstrumentedCache returns an EnrichCache implementation wrapped
// with metrics reporting and tracing
func NewInstrumentedCache(
	impl enriching.EnricherCache,
	metricsReporter extensions.MetricsReporter,
) enriching.EnricherCache {
	return &instrumentedCache{
		impl:            impl,
		metricsReporter: metricsReporter,
	}
}

func (c *instrumentedCache) Get(
	ctx context.Context,
	tenantID,
	leaderboardID string,
	members []*model.Member,
) (map[string]map[string]string, bool, error) {
	start := time.Now()

	span, ctx := opentracing.StartSpanFromContext(ctx, "podium.enriching_cache", opentracing.Tags{
		"tenant_id":      tenantID,
		"leaderboard_id": leaderboardID,
	})
	defer span.Finish()

	metadata, hit, err := c.impl.Get(ctx, tenantID, leaderboardID, members)

	c.metricsReporter.Increment(enrichmentCacheGets)
	c.metricsReporter.Timing(enrichmentCacheGetTimingMilli, time.Since(start))

	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.message", err.Error())

		c.metricsReporter.Increment(enrichmentCacheGetErrors)
	}

	if hit {
		c.metricsReporter.Increment(enrichmentCacheHits)
	}

	return metadata, hit, err
}

func (c *instrumentedCache) Set(
	ctx context.Context,
	tenantID,
	leaderboardID string,
	members []*model.Member,
	ttl time.Duration,
) error {
	start := time.Now()

	span, ctx := opentracing.StartSpanFromContext(ctx, "podium.enriching_cache", opentracing.Tags{
		"tenant_id":      tenantID,
		"leaderboard_id": leaderboardID,
		"ttl":            ttl,
	})
	defer span.Finish()

	err := c.impl.Set(ctx, tenantID, leaderboardID, members, ttl)

	c.metricsReporter.Increment(enrichmentCacheSets)
	c.metricsReporter.Timing(enrichmentCacheSetTimingMilli, time.Since(start))

	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.message", err.Error())

		c.metricsReporter.Increment(enrichmentCacheSetErrors)
	}

	return err
}

var _ enriching.EnricherCache = &instrumentedCache{}
