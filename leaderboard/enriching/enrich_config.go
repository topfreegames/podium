package enriching

import (
	"go.uber.org/zap"
	"time"
)

type (
	enrichmentConfig struct {
		// CloudSaveURL is the URL to call the Cloud Save service.
		cloudSave cloudSaveConfig

		// WebhookUrls contains the necessary parameters to call a webhook for a given game.
		// The key should be the game tenantID.
		webhookUrls map[string]string

		// WebhookTimeout is the timeout for the webhook call.
		webhookTimeout time.Duration
	}

	cloudSaveConfig struct {
		// Enabled indicates whether the Cloud Save service should be used for enrichment.
		enabled map[string]bool

		// URL is the URL to call the Cloud Save service.
		url string
	}
)

func newDefaultEnrichConfig() enrichmentConfig {
	return enrichmentConfig{
		cloudSave: cloudSaveConfig{
			enabled: map[string]bool{},
		},
		webhookUrls:    map[string]string{},
		webhookTimeout: 500 * time.Millisecond,
	}
}

type EnricherOptions func(*enricherImpl)

// WithCloudSaveUrl sets the Cloud Save URL.
func WithCloudSaveUrl(url string) EnricherOptions {
	return func(impl *enricherImpl) {
		impl.config.cloudSave.url = url
	}
}

// WithWebhookUrls sets the map of webhook URL for each tenantID.
func WithWebhookUrls(urlsMap map[string]string) EnricherOptions {
	return func(impl *enricherImpl) {
		impl.config.webhookUrls = urlsMap
	}
}

// WithWebhookTimeout sets the webhook timeout.
func WithWebhookTimeout(timeout time.Duration) EnricherOptions {
	return func(impl *enricherImpl) {
		impl.config.webhookTimeout = timeout
	}
}

// WithCloudSaveEnabled sets the map of enabled Cloud Save for each tenantID.
func WithCloudSaveEnabled(enabled map[string]bool) EnricherOptions {
	return func(impl *enricherImpl) {
		impl.config.cloudSave.enabled = enabled
	}
}

// WithLogger sets the logger.
func WithLogger(logger *zap.Logger) EnricherOptions {
	return func(impl *enricherImpl) {
		impl.logger = logger.With(zap.String("source", "enricher"))
	}
}
