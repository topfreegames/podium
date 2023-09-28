package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/topfreegames/podium/leaderboard/v2/enriching"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"time"
)

// cacheKeyFormat is {tenantID}:{memberID}
const cacheKeyFormat = "leaderboards-enrich-caching:%s:%s"

type enricherRedisCache struct {
	redis *redis.Client
}

var _ enriching.EnricherCache = &enricherRedisCache{}

func NewEnricherRedisCache(
	redis *redis.Client,
) enriching.EnricherCache {
	return &enricherRedisCache{
		redis: redis,
	}
}

func (e *enricherRedisCache) Get(
	ctx context.Context,
	tenantID string,
	members []*model.Member,
) (map[string]map[string]string, bool, error) {
	keys := getKeysFromMemberArray(tenantID, members)
	dataArray, err := e.redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, false, fmt.Errorf("failed to get data from cacheConfig: %w", err)
	}

	dataMap := make(map[string]map[string]string)
	for i, data := range dataArray {
		if data == nil {
			return nil, false, nil
		}

		unmarshaled := map[string]string{}
		err := json.Unmarshal([]byte(data.(string)), &unmarshaled)
		if err != nil {
			return nil, false, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		memberID := members[i].PublicID
		dataMap[memberID] = unmarshaled
	}

	return dataMap, true, nil
}

func (e *enricherRedisCache) Set(
	ctx context.Context,
	tenantID string,
	members []*model.Member,
	ttl time.Duration,
) error {
	keys := getKeysFromMemberArray(tenantID, members)
	cmds, err := e.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, member := range members {
			if member.Metadata != nil {
				marshaled, err := json.Marshal(member.Metadata)
				if err != nil {
					return fmt.Errorf("failed to marshal metadata: %w", err)
				}
				pipe.Set(ctx, keys[i], marshaled, ttl)
			}
		}
		return nil
	})

	if err != nil {
		cmdErrors := []error{}
		for _, cmd := range cmds {
			if cmd.Err() != nil {
				cmdErrors = append(cmdErrors, cmd.Err())
			}
		}
		cmdError := errors.Join(cmdErrors...)
		return fmt.Errorf("failed to set members in cache: %w", errors.Join(err, cmdError))
	}

	return nil
}

func getKeysFromMemberArray(tenantID string, members []*model.Member) []string {
	keys := make([]string, len(members))
	for i, member := range members {
		keys[i] = fmt.Sprintf(cacheKeyFormat, tenantID, member.PublicID)
	}
	return keys
}
