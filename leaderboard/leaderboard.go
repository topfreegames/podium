// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package leaderboard

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/topfreegames/extensions/redis/interfaces"
	"github.com/topfreegames/podium/leaderboard/expiration"
)

func getSetScoreScript(operation string) *redis.Script {
	return redis.NewScript(fmt.Sprintf(`
		-- Script params:
		-- KEYS[1] is the name of the leaderboard
		-- ARGV[1] are the Members JSON
		-- ARGV[2] is the leaderboard's expiration
		-- ARGV[3] defines if the previous rank should be returned
		-- ARGV[4] defines the ttl of the player score
		-- ARGV[5] defines the current unix timestamp

		-- creates leaderboard or just sets score of member
		local key_pairs = {}
		local members = cjson.decode(ARGV[1])
		local score_ttl = ARGV[4]
		if score_ttl == nil or score_ttl == "" then
			score_ttl = "inf"
		end
		for i,mem in ipairs(members) do
			table.insert(key_pairs, tonumber(mem["score"]))
			table.insert(key_pairs, mem["publicID"])
			if (ARGV[3] == "1") then
				mem["previousRank"] = tonumber(redis.call("ZREVRANK", KEYS[1], mem["publicID"])) or -2
			end
		end
		redis.call("%s", KEYS[1], unpack(key_pairs))

		-- If expiration is required set expiration
		if (ARGV[2] ~= "-1") then
			local expiration = redis.call("TTL", KEYS[1])
			if (expiration == -2) then
				return redis.error_reply("Leaderboard Set was not created in %s! Don't know how to proceed.")
			end
			if (expiration == -1) then
				redis.call("EXPIREAT", KEYS[1], ARGV[2])
			end
		end

		local expire_at = "nil"
		if (score_ttl ~= "inf") then
			local expiration_set_key = KEYS[1]..":ttl"
			expire_at = ARGV[5] + score_ttl
			key_pairs = {}
			for i,mem in ipairs(members) do
				table.insert(key_pairs, expire_at)
				table.insert(key_pairs, mem["publicID"])
			end
			redis.call("ZADD", expiration_set_key, unpack(key_pairs))
			redis.call("SADD", "expiration-sets", expiration_set_key)
		end

		-- return updated rank of member
		local result = {}
		for i,mem in ipairs(members) do
			table.insert(result, mem["publicID"])
			table.insert(result, tonumber(redis.call("ZREVRANK", KEYS[1], mem["publicID"])))
			table.insert(result, tonumber(redis.call("ZSCORE", KEYS[1], mem["publicID"])))
			if ARGV[3] == "1" then
				table.insert(result, mem["previousRank"])
			else
				table.insert(result, -1)
			end
			table.insert(result, expire_at)
		end
		return result
	`, operation, operation))
}

//GetMembersByRange for a given leaderboard
func (c *Client) GetMembersByRange(ctx context.Context, leaderboard string, startOffset int, endOffset int, order string) ([]*Member, error) {
	members, err := c.service.GetMembersByRange(ctx, leaderboard, startOffset, endOffset, order)
	if err != nil {
		return nil, err
	}

	return convertModelsToMembers(members), nil
}

// IncrementMemberScore sets the score to the member with the given ID
func (c *Client) IncrementMemberScore(ctx context.Context, leaderboardID string, memberID string, increment int, scoreTTL string) (*Member, error) {
	script := getSetScoreScript("ZINCRBY")

	expireAt, err := expiration.GetExpireAt(leaderboardID)
	if err != nil {
		if _, ok := err.(*expiration.LeaderboardExpiredError); ok {
			return nil, err
		} else {
			return nil, fmt.Errorf("Could not get expiration: %v", err)
		}
	}

	jsonMembers, _ := json.Marshal(Members{&Member{PublicID: memberID, Score: int64(increment)}})
	// TODO use prevRank instead of hard coded false
	result, err := script.Run(c.redisWithTracing(ctx), []string{leaderboardID}, jsonMembers, expireAt, false, scoreTTL, time.Now().Unix()).Result()
	if err != nil {
		return nil, fmt.Errorf("Could not increment score for member: %v", err)
	}
	rank := int(result.([]interface{})[1].(int64)) + 1
	score := result.([]interface{})[2].(int64)

	nMember := Member{PublicID: memberID, Score: score, Rank: rank}
	if scoreTTL != "" && scoreTTL != "inf" {
		nMember.ExpireAt = int(result.([]interface{})[4].(int64))
	}
	return &nMember, err
}

// SetMemberScore sets the score to the member with the given ID
func (c *Client) SetMemberScore(ctx context.Context, leaderboardID string, memberID string, score int64, prevRank bool, scoreTTL string) (*Member, error) {
	members := Members{&Member{PublicID: memberID, Score: score}}
	err := c.SetMembersScore(ctx, leaderboardID, members, prevRank, scoreTTL)
	return members[0], err
}

// SetMembersScore sets the scores of the members with the given IDs
func (c *Client) SetMembersScore(ctx context.Context, leaderboardID string, members Members, prevRank bool,
	scoreTTL string) error {

	expireAt, err := expiration.GetExpireAt(leaderboardID)
	if err != nil {
		if _, ok := err.(*expiration.LeaderboardExpiredError); ok {
			return err
		} else {
			return fmt.Errorf("Could not get expiration: %v", err)
		}
	}

	script := getSetScoreScript("ZADD")

	jsonMembers, _ := json.Marshal(members)
	newRanks, err := script.Run(c.redisWithTracing(ctx), []string{leaderboardID}, jsonMembers, expireAt, prevRank,
		scoreTTL, time.Now().Unix()).Result()
	if err != nil {
		return fmt.Errorf("Failed to update rank for members: %v", err)
	}

	res := newRanks.([]interface{})
	for i := 0; i < len(res); i += 5 {
		memberIndex := i / 5
		members[memberIndex].PublicID = res[i].(string)
		members[memberIndex].Score = res[i+2].(int64)
		members[memberIndex].Rank = int(res[i+1].(int64)) + 1
		members[memberIndex].PreviousRank = int(res[i+3].(int64)) + 1
		if scoreTTL != "" && scoreTTL != "inf" {
			members[memberIndex].ExpireAt = int(res[i+4].(int64))
		}
	}

	return err
}

// TotalMembers returns the total number of members in a given leaderboard
func (c *Client) TotalMembers(ctx context.Context, leaderboardID string) (int, error) {
	return c.service.GetTotalMembers(ctx, leaderboardID)
}

// RemoveMembers removes the members with the given publicIDs from the leaderboard
func (c *Client) RemoveMembers(ctx context.Context, leaderboardID string, memberIDs []interface{}) error {
	stringIds := make([]string, 0, len(memberIDs))
	for _, id := range memberIDs {
		stringIds = append(stringIds, fmt.Sprintf("%s", id))
	}
	return c.service.RemoveMembers(ctx, leaderboardID, stringIds)
}

// RemoveMember removes the member with the given publicID from the leaderboard
func (c *Client) RemoveMember(ctx context.Context, leaderboardID string, memberID string) error {
	return c.service.RemoveMember(ctx, leaderboardID, memberID)
}

// totalPages returns the number of pages of the leaderboard
func (c *Client) totalPages(redisClient interfaces.RedisClient, leaderboardID string, pageSize int) (int, error) {
	pages := 0
	total, err := redisClient.ZCard(leaderboardID).Result()
	if err != nil {
		return 0, fmt.Errorf("Number of pages could not be retrieved: %v", err)
	}
	pages = int(math.Ceil(float64(total) / float64(pageSize)))
	return pages, nil
}

func (c *Client) TotalPages(ctx context.Context, leaderboardID string, pageSize int) (int, error) {
	return c.service.TotalPages(ctx, leaderboardID, pageSize)
}

// GetMember returns the score and the rank of the member with the given ID
func (c *Client) GetMember(ctx context.Context, leaderboardID string, memberID string, order string, includeTTL bool) (*Member, error) {
	member, err := c.service.GetMember(ctx, leaderboardID, memberID, order, includeTTL)
	if err != nil {
		return nil, err
	}

	return convertModelToMember(member), nil
}

// GetMembers returns the score and the rank of the members with the given IDs
func (c *Client) GetMembers(ctx context.Context, leaderboardID string, memberIDs []string, order string, includeTTL bool) ([]*Member, error) {
	members, err := c.service.GetMembers(ctx, leaderboardID, memberIDs, order, includeTTL)
	if err != nil {
		return nil, err
	}

	return convertModelsToMembers(members), nil
}

// GetAroundMe returns a page of results centered in the member with the given ID
func (c *Client) GetAroundMe(ctx context.Context, leaderboardID string, pageSize int, memberID string, order string, getLastIfNotFound bool) ([]*Member, error) {
	members, err := c.service.GetAroundMe(ctx, leaderboardID, pageSize, memberID, order, getLastIfNotFound)
	if err != nil {
		return nil, err
	}

	return convertModelsToMembers(members), nil
}

// GetAroundScore returns a page of results centered in the score provided
func (c *Client) GetAroundScore(ctx context.Context, leaderboardID string, pageSize int, score int64, order string) ([]*Member, error) {
	members, err := c.service.GetAroundScore(ctx, leaderboardID, pageSize, score, order)
	if err != nil {
		return nil, err
	}

	return convertModelsToMembers(members), nil
}

// GetRank returns the rank of the member with the given ID
func (c *Client) GetRank(ctx context.Context, leaderboardID string, memberID string, order string) (int, error) {
	return c.service.GetRank(ctx, leaderboardID, memberID, order)
}

// GetLeaders returns a page of members with rank and score
func (c *Client) GetLeaders(ctx context.Context, leaderboardID string, pageSize, page int, order string) ([]*Member, error) {
	members, err := c.service.GetLeaders(ctx, leaderboardID, pageSize, page, order)
	if err != nil {
		return nil, err
	}

	return convertModelsToMembers(members), nil
}

//GetTopPercentage of members in the leaderboard.
func (c *Client) GetTopPercentage(ctx context.Context, leaderboardID string, pageSize, amount, maxMembers int, order string) ([]*Member, error) {
	if amount < 1 || amount > 100 {
		return nil, fmt.Errorf("Percentage must be a valid integer between 1 and 100.")
	}

	if order != "desc" && order != "asc" {
		order = "desc"
	}

	var operations = map[string]string{
		"range_desc": "ZREVRANGE",
		"rank_desc":  "ZREVRANK",
		"range_asc":  "ZRANGE",
		"rank_asc":   "ZRANK",
	}

	script := redis.NewScript(`
		-- Script params:
		-- KEYS[1] is the name of the leaderboard
		-- ARGV[1] is the desired percentage (0.0 to 1.0)
		-- ARGV[2] is the maximum number of members returned

		local totalNumber = redis.call("ZCARD", KEYS[1])
		local numberOfMembers = math.floor(ARGV[1] * totalNumber)
		if (numberOfMembers < 1) then
			numberOfMembers = 1
		end

		if (numberOfMembers > math.floor(ARGV[2])) then
			numberOfMembers = math.floor(ARGV[2])
		end

		local members = redis.call("` + operations["range_"+order] + `", KEYS[1], 0, numberOfMembers - 1, "WITHSCORES")
		local fullMembers = {}

		for index=1, #members, 2 do
			local publicID = members[index]
			local score = members[index + 1]
		 	local rank = redis.call("` + operations["rank_"+order] + `", KEYS[1], publicID)

			table.insert(fullMembers, publicID)
			table.insert(fullMembers, rank)
			table.insert(fullMembers, score)
		end

		return fullMembers
	`)

	result, err := script.Run(c.redisWithTracing(ctx), []string{leaderboardID}, float64(amount)/100.0, maxMembers).Result()

	if err != nil {
		return nil, fmt.Errorf("Getting top percentage of members failed; %v", err)
	}

	res := result.([]interface{})
	members := []*Member{}

	for i := 0; i < len(res); i += 3 {
		memberPublicID := res[i].(string)

		rank := int(res[i+1].(int64)) + 1
		score, _ := strconv.ParseInt(res[i+2].(string), 10, 64)

		members = append(members, &Member{
			PublicID: memberPublicID,
			Score:    score,
			Rank:     rank,
		})
	}

	return members, nil
}

// RemoveLeaderboard removes a leaderboard from redis
func (c *Client) RemoveLeaderboard(ctx context.Context, leaderboardID string) error {
	return c.service.RemoveLeaderboard(ctx, leaderboardID)
}

// Ping checks if the application is working properly
func (c *Client) Ping(ctx context.Context) (string, error) {
	err := c.service.Healthcheck(ctx)
	if err != nil {
		return "", err
	}

	return "Working", nil
}
