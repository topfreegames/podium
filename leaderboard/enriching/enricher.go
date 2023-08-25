package enriching

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	podium_leaderboard_webhooks_v1 "github.com/topfreegames/podium/leaderboard/v2/enriching/proto/webhook/v1"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"go.uber.org/zap"
)

const enrichWebhookEndpoint = "/leaderboards/enrich"
const cloudSaveEndpoint = "/get-public-documents/profile"

type (
	EnrichmentConfig struct {
		// CloudSaveURL is the URL to call the Cloud Save service.
		CloudSave CloudSaveConfig `mapstructure:"cloud_save"`

		// WebhookUrls contains the necessary parameters to call a webhook for a given game.
		// The key should be the game tenantID.
		WebhookUrls map[string]string `mapstructure:"webhook_urls"`

		// WebhookTimeout is the timeout for the webhook call.
		WebhookTimeout time.Duration `mapstructure:"webhook_timeout,default=2s"`
	}

	CloudSaveConfig struct {
		// Enabled indicates wheter the Cloud Save service should be used for enrichment.
		Disabled map[string]bool `mapstructure:"disabled"`

		// URL is the URL to call the Cloud Save service.
		Url string `mapstructure:"url"`
	}
)

type enricherImpl struct {
	config EnrichmentConfig
	logger *zap.Logger
	client *http.Client
}

// NewEnricher returns a new Enricher implementation.
func NewEnricher(config EnrichmentConfig, logger *zap.Logger) Enricher {
	return &enricherImpl{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: config.WebhookTimeout,
		},
	}
}

// Enrich enriches the members list with some metadata. By default, it will call the Cloud Save service.
// If there's a webhook configured for the tenantID, it will call it instead.
func (e *enricherImpl) Enrich(ctx context.Context, tenantID, leaderboardID string, members []*model.Member) ([]*model.Member, error) {
	if len(members) == 0 {
		return members, nil
	}

	tenantUrl, exists := e.config.WebhookUrls[tenantID]
	if !exists {
		e.logger.Debug(fmt.Sprintf("no webhook configured for tentantID '%s'. Will call Cloud Save.", tenantID))
		return e.enrichWithCloudSave(ctx, tenantID, members)
	}

	e.logger.Debug(fmt.Sprintf("calling webhook for tenantID '%s'.", tenantID))

	body := membersModelToProto(leaderboardID, members)
	jsonData, err := json.Marshal(podium_leaderboard_webhooks_v1.EnrichLeaderboardsRequest{Members: body})
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", errors.Join(err, ErrEnrichmentInternal))
	}

	webhookUrl, err := buildUrl(tenantUrl, enrichWebhookEndpoint)
	if err != nil {
		return nil, fmt.Errorf("could not build webhook URL: %w", errors.Join(err, ErrEnrichmentInternal))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookUrl, bytes.NewBuffer(jsonData))

	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", errors.Join(err, ErrEnrichmentInternal))
	}

	req.Header.Set("Content-Type", "application/json")

	e.logger.Debug(fmt.Sprintf("calling enrichment webhook '%s' for tenantID '%s'", webhookUrl, tenantID))
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not complete request to webhook: %w", errors.Join(err, ErrEnrichmentCall))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("webhook returned non OK response: %w", errors.Join(err, ErrEnrichmentCall))
	}

	var result podium_leaderboard_webhooks_v1.EnrichLeaderboardsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("could not unmarshal webhook response: %w", errors.Join(err, ErrEnrichmentCall))
	}

	metadataMap := make(map[string]map[string]string)
	for _, m := range result.Members {
		metadataMap[m.Id] = m.Metadata
	}

	for _, m := range members {
		if data, ok := metadataMap[m.PublicID]; ok {
			m.Metadata = data
		}
	}

	return members, nil
}

func (e *enricherImpl) enrichWithCloudSave(ctx context.Context, tenantID string, members []*model.Member) ([]*model.Member, error) {
	if e.config.CloudSave.Disabled[tenantID] {
		e.logger.Debug(fmt.Sprintf("cloud save enrich disabled for tenant %s. Skipping enrichment.", tenantID))
		return members, nil
	}

	if e.config.CloudSave.Url == "" {
		e.logger.Debug("cloud Save URL not configured. Skipping enrichment.")
		return members, nil
	}

	e.logger.Debug(fmt.Sprintf("calling cloud save for tenantID '%s'", tenantID))

	ids := make([]string, len(members))
	for i, m := range members {
		ids[i] = m.PublicID
	}

	request := CloudSaveGetProfilesRequest{
		TenantID: tenantID,
		IDs:      ids,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", errors.Join(err, ErrEnrichmentInternal))
	}

	url, err := buildUrl(e.config.CloudSave.Url, cloudSaveEndpoint)
	if err != nil {
		return nil, fmt.Errorf("could not build cloud save url: %w", errors.Join(ErrEnrichmentInternal, err))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", errors.Join(ErrEnrichmentInternal, err))
	}

	req.Header.Set("Content-Type", "application/json")
	raw, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not complete request to cloud save: %w", errors.Join(ErrEnrichmentCall, err))
	}
	defer raw.Body.Close()

	if raw.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cloud save returned non OK response: %w", errors.Join(err, ErrEnrichmentCall))
	}

	res := &CloudSaveGetProfilesResponse{}
	if err := json.NewDecoder(raw.Body).Decode(res); err != nil {
		return nil, fmt.Errorf("could not unmarshal cloud save response: %w", errors.Join(ErrEnrichmentCall, err))
	}

	cloudSaveMetadataMap := make(map[string]map[string]string)
	for _, d := range res.Documents {
		cloudSaveMap[d.AccountID] = d.Data
	}

	for _, m := range members {
		if data, ok := cloudSaveMap[m.PublicID]; ok {
			m.Metadata = data
		}
	}

	return members, nil
}

func buildUrl(baseUrl, endpoint string) (string, error) {
	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl += "/"
	}
	if !strings.HasPrefix(baseUrl, "http") {
		baseUrl = fmt.Sprintf("http://%s", baseUrl)
	}

	u, err := url.JoinPath(baseUrl, endpoint)
	if err != nil {
		return "", fmt.Errorf("could not join url paths: %w", errors.Join(err, ErrEnrichmentInternal))
	}

	parsedUrl, err := url.ParseRequestURI(u)
	if err != nil {
		return "", fmt.Errorf("could not parse url %s: %w", u, errors.Join(err, ErrEnrichmentInternal))
	}

	return parsedUrl.String(), nil
}

func membersModelToProto(leaderboardID string, members []*model.Member) []*podium_leaderboard_webhooks_v1.Member {
	protoMembers := make([]*podium_leaderboard_webhooks_v1.Member, len(members))
	for i, m := range members {
		protoMembers[i] = &podium_leaderboard_webhooks_v1.Member{
			LeaderboardId: leaderboardID,
			Id:            m.PublicID,
			Scores:        []*podium_leaderboard_webhooks_v1.Score{{Value: m.Score}},
			Rank:          int32(m.Rank),
		}
	}

	return protoMembers
}
