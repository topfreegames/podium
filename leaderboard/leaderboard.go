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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/topfreegames/extensions/redis/interfaces"
	"github.com/topfreegames/podium/leaderboard/expiration"

	tfgredis "github.com/topfreegames/extensions/redis"
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

//getMembersByRange for a given leaderboard
func getMembersByRange(redisClient interfaces.RedisClient, leaderboard string, startOffset int, endOffset int, order string) ([]*Member, error) {
	cli := redisClient

	var values []redis.Z
	var err error
	if strings.Compare(order, "desc") == 0 {
		values, err = cli.ZRevRangeWithScores(leaderboard, int64(startOffset), int64(endOffset)).Result()
	} else {
		values, err = cli.ZRangeWithScores(leaderboard, int64(startOffset), int64(endOffset)).Result()
	}
	if err != nil {
		return nil, fmt.Errorf("Retrieval of members for range (start %d end %d) failed: %v", startOffset, endOffset, err)
	}
	members := make([]*Member, len(values))
	for i := 0; i < len(members); i++ {
		publicID := values[i].Member.(string)
		score := values[i].Score
		nMember := Member{PublicID: publicID, Score: int64(score), Rank: int(startOffset + i + 1)}
		members[i] = &nMember
	}

	return members, nil
}

//GetMembersByRange for a given leaderboard
func (c *Client) GetMembersByRange(ctx context.Context, leaderboard string, startOffset int, endOffset int, order string) ([]*Member, error) {
	return getMembersByRange(c.redisWithTracing(ctx), leaderboard, startOffset, endOffset, order)
}

// getMemberIDWithClosestScore returns a member in a given leaderboard with score >= the score provided
func getMemberIDWithClosestScore(redisClient interfaces.RedisClient, leaderboard string, score int64) (string, error) {
	cli := redisClient

	values, err := cli.ZRevRangeByScore(leaderboard, redis.ZRangeBy{Min: "-inf", Max: strconv.FormatInt(score, 10), Offset: 0, Count: 1}).Result()

	if err != nil {
		return "", fmt.Errorf("Retrieval of member with closest score to %d failed: %v", score, err)
	}

	if len(values) < 1 {
		return "", nil
	}

	return values[0], nil
}

//NewClientWithRedis creates a leaderboard using an already connected tfg Redis
func NewClientWithRedis(cli *tfgredis.Client) *Client {
	return &Client{redisClient: cli}
}

// IncrementMemberScore sets the score to the member with the given ID
func (c *Client) IncrementMemberScore(ctx context.Context, leaderboardID string, memberID string, increment int,
	scoreTTL string) (*Member, error) {

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

func (c *Client) totalMembers(r interfaces.RedisClient, leaderboardID string) (int, error) {
	total, err := r.ZCard(leaderboardID).Result()
	if err != nil {
		return 0, fmt.Errorf("Retrieval of total members failed: %v", err)
	}
	return int(total), nil
}

// TotalMembers returns the total number of members in a given leaderboard
func (c *Client) TotalMembers(ctx context.Context, leaderboardID string) (int, error) {
	return c.totalMembers(c.redisWithTracing(ctx), leaderboardID)
}

// RemoveMembers removes the members with the given publicIDs from the leaderboard
func (c *Client) RemoveMembers(ctx context.Context, leaderboardID string, memberIDs []interface{}) error {
	_, err := c.redisWithTracing(ctx).ZRem(leaderboardID, memberIDs...).Result()
	if err != nil {
		return fmt.Errorf("Members removal failed: %v", err)
	}
	return nil
}

// RemoveMember removes the member with the given publicID from the leaderboard
func (c *Client) RemoveMember(ctx context.Context, leaderboardID string, memberID string) error {
	_, err := c.redisWithTracing(ctx).ZRem(leaderboardID, memberID).Result()
	if err != nil {
		return fmt.Errorf("Member removal failed: %v", err)
	}
	return nil
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
	return c.totalPages(c.redisWithTracing(ctx), leaderboardID, pageSize)
}

func (c *Client) getMember(r interfaces.RedisClient, leaderboardID string, memberID string, order string, includeTTL bool) (*Member, error) {
	if order != "desc" && order != "asc" {
		order = "desc"
	}

	var operations = map[string]string{
		"rank_desc": "ZREVRANK",
		"rank_asc":  "ZRANK",
	}

	script := redis.NewScript(`
		-- Script params:
		-- KEYS[1] is the name of the leaderboard
		-- KEYS[2] is member's public ID
		-- ARGV[1] is a bool indicating whether the score ttl should be retrieved

        local score_ttl = ARGV[1] == "true"
		-- gets rank of the member
		local rank = redis.call("` + operations["rank_"+order] + `", KEYS[1], KEYS[2])
		local score = redis.call("ZSCORE", KEYS[1], KEYS[2])
        if score_ttl then
			local expire_at = redis.call("ZSCORE", KEYS[1]..":ttl", KEYS[2])
			return {rank,score,expire_at}
        end
		return {rank,score}
	`)

	result, err := script.Run(r, []string{leaderboardID, memberID}, strconv.FormatBool(includeTTL)).Result()
	if err != nil {
		return nil, fmt.Errorf("Getting member information failed: %v", err)
	}

	res := result.([]interface{})

	if res[0] == nil || res[1] == nil {
		return nil, NewMemberNotFound(leaderboardID, memberID)
	}

	rank := int(res[0].(int64))
	score, _ := strconv.ParseInt(res[1].(string), 10, 64)

	nMember := Member{PublicID: memberID, Score: score, Rank: rank + 1}
	if includeTTL {
		if expireAtStr, ok := res[2].(string); ok {
			expireAtParsed, _ := strconv.ParseInt(expireAtStr, 10, 32)
			nMember.ExpireAt = int(expireAtParsed)
		}
	}
	return &nMember, nil
}

// GetMember returns the score and the rank of the member with the given ID
func (c *Client) GetMember(ctx context.Context, leaderboardID string, memberID string, order string, includeTTL bool) (*Member, error) {
	return c.getMember(c.redisWithTracing(ctx), leaderboardID, memberID, order, includeTTL)
}

// GetMembers returns the score and the rank of the members with the given IDs
func (c *Client) GetMembers(ctx context.Context, leaderboardID string, memberIDs []string, order string, includeTTL bool) ([]*Member, error) {

	var operations = map[string]string{
		"rank_desc": "ZREVRANK",
		"rank_asc":  "ZRANK",
	}

	script := redis.NewScript(`
		-- Script params:
		-- KEYS[1] is the name of the leaderboard
		-- ARGV[1] is member's public IDs
		-- ARGV[2] is a bool indicating whether the score ttl should be retrieved

        local score_ttl = ARGV[2] == "true"
		local members = {}

		for publicID in string.gmatch(ARGV[1], '([^,]+)') do
			-- gets rank of the member
			local rank = redis.call("` + operations["rank_"+order] + `", KEYS[1], publicID)
			local score = redis.call("ZSCORE", KEYS[1], publicID)

			table.insert(members, publicID)
			table.insert(members, rank)
			table.insert(members, score)

			if score_ttl then
				local expire_at = redis.call("ZSCORE", KEYS[1]..":ttl", publicID)
				table.insert(members, expire_at)
			else
				table.insert(members, "nil")
			end
		end

		return members
	`)

	result, err := script.Run(c.redisWithTracing(ctx), []string{leaderboardID}, strings.Join(memberIDs, ","), strconv.FormatBool(includeTTL)).Result()
	if err != nil {
		return nil, fmt.Errorf("Getting members information failed: %v", err)
	}

	res := result.([]interface{})
	members := Members{}
	for i := 0; i < len(res); i += 4 {
		memberPublicID := res[i].(string)
		if res[i+1] == nil || res[i+2] == nil {
			continue
		}

		rank := int(res[i+1].(int64)) + 1
		score, _ := strconv.ParseInt(res[i+2].(string), 10, 64)
		member := &Member{
			PublicID: memberPublicID,
			Score:    score,
			Rank:     rank,
		}
		if includeTTL {
			if expireAtStr, ok := res[i+3].(string); ok {
				expireAtParsed, _ := strconv.ParseInt(expireAtStr, 10, 32)
				member.ExpireAt = int(expireAtParsed)
			}
		}

		members = append(members, member)
	}

	sort.Sort(members)
	return members, nil
}

func (c *Client) getAroundMe(redisClient interfaces.RedisClient, leaderboardID string, pageSize int, memberID string,
	order string, getLastIfNotFound bool) ([]*Member, error) {

	if order != "desc" && order != "asc" {
		order = "desc"
	}

	currentMember, err := c.getMember(redisClient, leaderboardID, memberID, order, false)
	_, memberNotFound := err.(*MemberNotFoundError)
	if (err != nil && !memberNotFound) || (memberNotFound && !getLastIfNotFound) {
		return nil, err
	}

	totalMembers, err := c.totalMembers(redisClient, leaderboardID)
	if err != nil {
		return nil, err
	}

	if memberNotFound && getLastIfNotFound {
		currentMember = &Member{PublicID: memberID, Score: 0, Rank: totalMembers + 1}
	}

	startOffset := currentMember.Rank - (pageSize / 2)
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + pageSize) - 1
	if totalMembers < endOffset {
		endOffset = totalMembers
		startOffset = endOffset - pageSize
		if startOffset < 0 {
			startOffset = 0
		}
	}

	members, err := getMembersByRange(redisClient, leaderboardID, startOffset, endOffset, order)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve information around a specific member: %v", err)
	}

	return members, nil
}

// GetAroundMe returns a page of results centered in the member with the given ID
func (c *Client) GetAroundMe(ctx context.Context, leaderboardID string, pageSize int, memberID string, order string,
	getLastIfNotFound bool) ([]*Member, error) {
	return c.getAroundMe(c.redisWithTracing(ctx), leaderboardID, pageSize, memberID, order, getLastIfNotFound)
}

// GetAroundScore returns a page of results centered in the score provided
func (c *Client) GetAroundScore(ctx context.Context, leaderboardID string, pageSize int, score int64, order string) ([]*Member, error) {
	//getMembersByRange(c.RedisClient, c.PublicID, startOffset, endOffset, order, l)
	redisClient := c.redisWithTracing(ctx)
	memberID, err := getMemberIDWithClosestScore(redisClient, leaderboardID, score)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve information around a specific score (%d): %v", score, err)
	}

	return c.getAroundMe(redisClient, leaderboardID, pageSize, memberID, order, true)
}

// GetRank returns the rank of the member with the given ID
func (c *Client) GetRank(ctx context.Context, leaderboardID string, memberID string, order string) (int, error) {
	var rank int64
	var err error
	if order == "desc" {
		rank, err = c.redisWithTracing(ctx).ZRevRank(leaderboardID, memberID).Result()
	} else {
		rank, err = c.redisWithTracing(ctx).ZRank(leaderboardID, memberID).Result()
	}
	if err != nil {
		if strings.HasPrefix(err.Error(), "redis: nil") {
			return -1, NewMemberNotFound(leaderboardID, memberID)
		}

		return -1, fmt.Errorf("Failed to retrieve rank of specific member: %v", err)
	}
	return int(rank + 1), nil
}

// GetLeaders returns a page of members with rank and score
func (c *Client) GetLeaders(ctx context.Context, leaderboardID string, pageSize, page int, order string) ([]*Member, error) {
	redisClient := c.redisWithTracing(ctx)
	if page < 1 {
		page = 1
	}

	totalPages, err := c.totalPages(redisClient, leaderboardID, pageSize)
	if err != nil {
		return nil, err
	}

	if page > totalPages {
		return make([]*Member, 0), nil
	}

	redisIndex := page - 1
	startOffset := redisIndex * pageSize
	endOffset := (startOffset + pageSize) - 1
	return getMembersByRange(redisClient, leaderboardID, startOffset, endOffset, order)
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
	_, err := c.redisWithTracing(ctx).Del(leaderboardID).Result()
	if err != nil {
		return fmt.Errorf("Failed to remove leaderboard: %v", err)
	}

	return nil
}

// Ping checks if the application is working properly
func (c *Client) Ping(ctx context.Context) (string, error) {
	return c.redisWithTracing(ctx).Ping().Result()
}
