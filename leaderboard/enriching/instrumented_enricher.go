package enriching

import (
	"context"
	"github.com/opentracing/opentracing-go"
	extensions "github.com/topfreegames/extensions/middleware"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"time"
)

const (
	enrichmentTimingMilli = "enrichment_duration_milliseconds"
	enrichmentCalls       = "enrichment_calls"
	enrichmentErrors      = "enrichment_errors"
)

type instrumentedEnricher struct {
	impl            Enricher
	metricsReporter extensions.MetricsReporter
}

func NewInstrumentedEnricher(
	impl Enricher,
	metricsReporter extensions.MetricsReporter,
) Enricher {
	return &instrumentedEnricher{
		impl:            impl,
		metricsReporter: metricsReporter,
	}
}

func (en *instrumentedEnricher) Enrich(ctx context.Context, tenantID, leaderboardID string, members []*model.Member) ([]*model.Member, error) {
	start := time.Now()

	span, ctx := opentracing.StartSpanFromContext(ctx, "podium.enriching", opentracing.Tags{
		"tenant_id":      tenantID,
		"leaderboard_id": leaderboardID,
	})
	defer span.Finish()

	members, err := en.impl.Enrich(ctx, tenantID, leaderboardID, members)

	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.message", err.Error())

		en.metricsReporter.Increment(enrichmentErrors)
	}

	en.metricsReporter.Increment(enrichmentCalls)
	en.metricsReporter.Timing(enrichmentTimingMilli, time.Since(start))

	return members, err
}

var _ Enricher = (*instrumentedEnricher)(nil)
