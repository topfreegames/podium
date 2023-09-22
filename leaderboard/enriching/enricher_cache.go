package enriching

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"go.uber.org/zap"
	"time"
)

// cacheKeyFormat is {tenantID}:{leaderboardID}:{memberID}
const cacheKeyFormat = "leaderboards-enrich-caching:%s:%s:%s"

type enricherCache struct {
	redis  *redis.Client
	logger *zap.Logger
}

var _ EnricherCache = &enricherCache{}

func NewEnricherCache(
	logger *zap.Logger,
	redis *redis.Client,
) EnricherCache {
	return &enricherCache{
		redis:  redis,
		logger: logger.With(zap.String("source", "enricherCache")),
	}
}

func (e *enricherCache) Get(
	ctx context.Context,
	tenantID,
	leaderboardID string,
	members []*model.Member,
) (map[string]map[string]string, bool, error) {
	l := e.logger.With(
		zap.String("method", "Get"),
		zap.String("tenantID", tenantID),
		zap.String("leaderboardID", leaderboardID),
	)

	keys := getKeysFromMemberArray(tenantID, leaderboardID, members)
	dataArray, err := e.redis.MGet(ctx, keys...).Result()
	if err != nil {
		l.With(zap.Error(err)).Error("failed to get members from cache")
		return nil, false, fmt.Errorf("failed to get data from cache: %w", err)
	}

	dataMap := make(map[string]map[string]string)
	for i, data := range dataArray {
		if data == nil {
			return nil, false, nil
		}

		unmarshaled := map[string]string{}
		err := json.Unmarshal([]byte(data.(string)), &unmarshaled)
		if err != nil {
			l.With(zap.Error(err)).Error("failed to unmarshal data")
			return nil, false, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		memberID := members[i].PublicID
		dataMap[memberID] = unmarshaled
	}

	return dataMap, true, nil
}

func (e *enricherCache) Set(
	ctx context.Context,
	tenantID,
	leaderboardID string,
	members []*model.Member,
	ttl time.Duration,
) error {
	l := e.logger.With(
		zap.String("method", "Set"),
		zap.String("tenantID", tenantID),
		zap.String("leaderboardID", leaderboardID),
	)

	keys := getKeysFromMemberArray(tenantID, leaderboardID, members)
	pipe := e.redis.TxPipeline()
	for i, member := range members {
		marshaled, err := json.Marshal(member.Metadata)
		if err != nil {
			l.With(zap.Error(err)).Error("failed to marshal metadata")
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		pipe.Set(ctx, keys[i], marshaled, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		l.With(zap.Error(err)).Error("failed to set members in cache")
		return fmt.Errorf("failed to set members in cache: %w", err)
	}

	return nil
}

func getKeysFromMemberArray(tenantID, leaderboardID string, members []*model.Member) []string {
	keys := make([]string, len(members))
	for i, member := range members {
		keys[i] = fmt.Sprintf(cacheKeyFormat, tenantID, leaderboardID, member.PublicID)
	}
	return keys
}
