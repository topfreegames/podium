package enriching

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	podium_leaderboard_webhooks_v1 "github.com/topfreegames/podium/leaderboard/v2/enriching/proto/webhook/v1"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"net/http"
	"net/url"
	"strings"
)

const enrichURL = "/leaderboards/enrich"

type EnrichmentConfig struct {
	// WebhookUrls contains the necessary parameters to call a webhook for a given game.
	// The key should be the game tenantID.
	WebhookUrls map[string]string `mapstructure:"webhook_urls"`
}

type enricherImpl struct {
	config EnrichmentConfig
}

// NewEnricher returns a new Enricher implementation.
func NewEnricher(config EnrichmentConfig) Enricher {
	return &enricherImpl{
		config: config,
	}
}

func (e *enricherImpl) Enrich(tenantID, leaderboardID string, members []*model.Member) ([]*model.Member, error) {
	tenantUrl, exists := e.config.WebhookUrls[tenantID]
	if !exists {
		return members, nil
	}

	body := membersModelToProto(leaderboardID, members)
	jsonData, err := json.Marshal(podium_leaderboard_webhooks_v1.EnrichLeaderboardsRequest{Members: body})
	if err != nil {
		return nil, fmt.Errorf("could not marshal request: %w", errors.Join(err, ErrEnrichmentInternal))
	}

	webhookUrl, err := buildUrl(tenantUrl)
	if err != nil {
		return nil, fmt.Errorf("could not build webhook URL: %w", errors.Join(err, ErrEnrichmentInternal))
	}

	req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", errors.Join(err, ErrEnrichmentInternal))
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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

	return protoToMemberModels(result.Members), nil
}

func buildUrl(baseUrl string) (string, error) {
	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl += "/"
	}
	if !strings.HasPrefix(baseUrl, "http") {
		baseUrl = fmt.Sprintf("http://%s", baseUrl)
	}

	u, err := url.JoinPath(baseUrl, enrichURL)
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
			MemberId:      m.PublicID,
			Scores:        []*podium_leaderboard_webhooks_v1.Score{{Value: m.Score}},
			Rank:          int32(m.Rank),
		}
	}

	return protoMembers
}

func protoToMemberModels(protoMembers []*podium_leaderboard_webhooks_v1.Member) []*model.Member {
	members := make([]*model.Member, len(protoMembers))
	for i, m := range protoMembers {
		score := int64(0)
		if len(m.Scores) > 0 {
			score = m.Scores[0].Value
		}
		members[i] = &model.Member{
			PublicID: m.MemberId,
			Score:    score,
			Rank:     int(m.Rank),
			Metadata: m.Metadata,
		}
	}

	return members
}
