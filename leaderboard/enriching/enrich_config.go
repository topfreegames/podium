package enriching

import (
	"go.uber.org/zap"
	"time"
)

type (
	enrichmentConfig struct {
		// CloudSaveURL is the URL to call the Cloud Save service.
		cloudSave cloudSaveConfig `mapstructure:"cloud_save"`

		// WebhookUrls contains the necessary parameters to call a webhook for a given game.
		// The key should be the game tenantID.
		webhookUrls map[string]string `mapstructure:"webhook_urls"`

		// WebhookTimeout is the timeout for the webhook call.
		webhookTimeout time.Duration `mapstructure:"webhook_timeout"`

		cache *cache `mapstructure:"cache"`
	}

	cache struct {
		// TTL is the time to live for the cached data.
		ttl time.Duration `mapstructure:"ttl"`
	}

	cloudSaveConfig struct {
		// Enabled indicates whether the Cloud Save service should be used for enrichment.
		disabled map[string]bool `mapstructure:"disabled"`

		// URL is the URL to call the Cloud Save service.
		url string `mapstructure:"url"`
	}
)

func newDefaultEnrichConfig() enrichmentConfig {
	return enrichmentConfig{
		cloudSave: cloudSaveConfig{
			disabled: map[string]bool{},
		},
		webhookUrls:    map[string]string{},
		webhookTimeout: 500 * time.Millisecond,
		cache: &cache{
			ttl: 24 * time.Hour,
		},
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

// WithCloudSaveDisabled sets the map of disabled Cloud Save for each tenantID.
func WithCloudSaveDisabled(disabled map[string]bool) EnricherOptions {
	return func(impl *enricherImpl) {
		impl.config.cloudSave.disabled = disabled
	}
}

// WithLogger sets the logger.
func WithLogger(logger *zap.Logger) EnricherOptions {
	return func(impl *enricherImpl) {
		impl.logger = logger.With(zap.String("source", "enricher"))
	}
}
