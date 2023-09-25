package enriching

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/topfreegames/podium/leaderboard/v2/enriching/cloud-save"
	podium_leaderboard_webhooks_v1 "github.com/topfreegames/podium/leaderboard/v2/enriching/proto/webhook/v1"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
)

const enrichWebhookEndpoint = "/leaderboards/enrich"
const cloudSaveEndpoint = "/get-public-documents/wildlife-platform-player-profile"

type enricherImpl struct {
	config enrichmentConfig
	logger *zap.Logger
	client *http.Client
}

// NewEnricher returns a new Enricher implementation.
func NewEnricher(
	options ...EnricherOptions,
) Enricher {
	config := newDefaultEnrichConfig()
	e := &enricherImpl{
		config: config,
		logger: zap.NewNop(),
		client: &http.Client{
			Timeout: config.webhookTimeout,
		},
	}

	for _, opt := range options {
		opt(e)
	}

	return e
}

// Enrich enriches the members list with some metadata.
// By default, it will call the Cloud Save service, unless it's enabled for the tenantID or if there's a webhook for the tenantID.
// If there's a webhook configured for the tenantID, it will be called instead.
func (e *enricherImpl) Enrich(
	ctx context.Context,
	tenantID,
	leaderboardID string,
	members []*model.Member,
) ([]*model.Member, error) {
	if len(members) == 0 {
		return members, nil
	}

	l := e.logger.With(
		zap.String("method", "Enrich"),
		zap.String("tenantID", tenantID),
		zap.String("leaderboardID", leaderboardID),
	)

	tenantUrl, webHookExists := e.config.webhookUrls[tenantID]
	cloudSaveEnabled := e.config.cloudSave.enabled[tenantID]

	if !webHookExists && !cloudSaveEnabled {
		return members, nil
	}

	if webHookExists && tenantUrl != "" {
		members, err := e.enrichWithWebhook(ctx, tenantUrl, leaderboardID, members)

		if err != nil {
			l.Error("could not enrich with webhook", zap.Error(err))
			return nil, fmt.Errorf("could not enrich with webhook: %w", err)
		}

		return members, nil
	}

	if cloudSaveEnabled {
		e.logger.Debug(fmt.Sprintf("no webhook configured for tentantID '%s'. will call Cloud Save.", tenantID))
		members, err := e.enrichWithCloudSave(ctx, tenantID, members)

		if err != nil {
			l.Error("could not enrich with cloud save", zap.Error(err))
			return nil, fmt.Errorf("could not enrich with cloud save: %w", err)
		}

		return members, nil
	}

	l.Debug(fmt.Sprintf("no webhook configured for tentantID '%s' and cloud save enabled. Skipping enrichment.", tenantID))

	return members, nil
}

func (e *enricherImpl) enrichWithWebhook(
	ctx context.Context,
	url,
	leaderboardID string,
	members []*model.Member,
) ([]*model.Member, error) {
	l := e.logger.With(
		zap.String("url", url),
		zap.String("leaderboardID", leaderboardID),
		zap.String("method", "enrichWithWebhook"),
	)

	body := membersModelToProto(leaderboardID, members)
	jsonData, err := json.Marshal(podium_leaderboard_webhooks_v1.EnrichLeaderboardsRequest{Members: body})
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", errors.Join(err, ErrEnrichmentInternal))
	}

	webhookUrl, err := buildUrl(url, enrichWebhookEndpoint)
	if err != nil {
		return nil, fmt.Errorf("could not build webhook URL: %w", errors.Join(err, ErrEnrichmentInternal))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookUrl, bytes.NewBuffer(jsonData))

	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", errors.Join(err, ErrEnrichmentInternal))
	}

	req.Header.Set("Content-Type", "application/json")

	l.Debug(fmt.Sprintf("calling enrichment webhook '%s'", webhookUrl))
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not complete request to webhook: %w", errors.Join(err, ErrEnrichmentCall))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("webhook returned %s response: %w", resp.Status, ErrEnrichmentCall)
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
	l := e.logger.With(
		zap.String("method", "enrichWithCloudSave"),
		zap.String("tenantID", tenantID),
	)

	if e.config.cloudSave.url == "" {
		e.logger.Debug("cloud Save URL not configured. Skipping enrichment.")
		return members, nil
	}

	l.Debug(fmt.Sprintf("calling cloud save for tenantID '%s'", tenantID))

	ids := make([]string, len(members))
	for i, m := range members {
		ids[i] = m.PublicID
	}

	request := cloud_save.CloudSaveGetProfilesRequest{
		TenantID: tenantID,
		IDs:      ids,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", errors.Join(ErrEnrichmentInternal, err))
	}

	url, err := buildUrl(e.config.cloudSave.url, cloudSaveEndpoint)
	if err != nil {
		return nil, fmt.Errorf("could not build cloud save url: %w", errors.Join(ErrEnrichmentInternal, err))
	}

	l.Debug(fmt.Sprintf("calling cloud save endpoint '%s'", url))

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
		return nil, fmt.Errorf("cloud save returned %s response: %w", raw.Status, ErrEnrichmentCall)
	}

	res := &cloud_save.CloudSaveGetProfilesResponse{}
	if err := json.NewDecoder(raw.Body).Decode(res); err != nil {
		return nil, fmt.Errorf("could not unmarshal cloud save response: %w", errors.Join(ErrEnrichmentCall, err))
	}

	cloudSaveMetadataMap := make(map[string]map[string]string)
	for _, d := range res.Documents {
		cloudSaveMetadataMap[d.AccountID] = d.Data
	}

	for _, m := range members {
		if data, ok := cloudSaveMetadataMap[m.PublicID]; ok {
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
