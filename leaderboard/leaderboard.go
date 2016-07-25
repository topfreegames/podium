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
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/topfreegames/podium/util"
	"github.com/uber-go/zap"
	redis "gopkg.in/redis.v4"
)

//MemberNotFoundError indicates member was not found in Redis
type MemberNotFoundError struct {
	LeaderboardID string
	MemberID        string
}

func (e *MemberNotFoundError) Error() string {
	return fmt.Sprintf("Could not find data for member %s in leaderboard %s.", e.MemberID, e.LeaderboardID)
}

//NewMemberNotFound returns a new error for member not found
func NewMemberNotFound(leaderboardID, memberID string) *MemberNotFoundError {
	return &MemberNotFoundError{
		LeaderboardID: leaderboardID,
		MemberID:        memberID,
	}
}

// Member maps an member identified by their publicID to their score and rank
type Member struct {
	PublicID string
	Score    int
	Rank     int
}

// Team groups sets of members
type Team struct {
	PublicID string
	Members  map[string]Member
	Rank     int
}

// Leaderboard identifies a leaderboard with given redis client
type Leaderboard struct {
	Logger      zap.Logger
	RedisClient *util.RedisClient
	PublicID    string
	PageSize    int
}

func getMembersByRange(redisClient *util.RedisClient, leaderboard string, pageSize int, startOffset int, endOffset int, l zap.Logger) ([]*Member, error) {
	cli := redisClient.Client
	l.Debug(fmt.Sprintf("Retrieving members for range: %d - %d", startOffset, endOffset))
	values, err := cli.ZRevRangeWithScores(leaderboard, int64(startOffset), int64(endOffset)).Result()
	if err != nil {
		l.Error(fmt.Sprintf("Retrieval of members for range %d - %d failed.", startOffset, endOffset), zap.Error(err))
		return nil, err
	}
	l.Info(fmt.Sprintf("Retrieval of members for range %d - %d succeeded.", startOffset, endOffset))

	l.Debug("Building details of leaderboard members...")
	members := make([]*Member, len(values))
	for i := 0; i < len(members); i++ {
		publicID := values[i].Member.(string)
		score := int(values[i].Score)
		nMember := Member{PublicID: publicID, Score: score, Rank: int(startOffset + i + 1)}
		members[i] = &nMember
	}
	l.Info("Retrieval of leaderboard members' details succeeded.")
	return members, nil
}

// NewLeaderboard creates a new Leaderboard with given settings, ID and pageSize
func NewLeaderboard(redisClient *util.RedisClient, publicID string, pageSize int, logger zap.Logger) *Leaderboard {
	return &Leaderboard{RedisClient: redisClient, PublicID: publicID, PageSize: pageSize, Logger: logger}
}

//AddToLeaderboardSet adds a score to a leaderboard set respecting expiration
func (lb *Leaderboard) AddToLeaderboardSet(redisCli *util.RedisClient, memberID string, score int) (int, error) {
	cli := redisCli.Client

	l := lb.Logger.With(
		zap.String("operation", "AddToLeaderboardSet"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.String("memberID", memberID),
		zap.Int("score", score),
	)

	l.Debug("Calculating expiration for leaderboard...")
	expireAt, err := util.GetExpireAt(lb.PublicID)
	if err != nil {
		l.Error("Could not get expiration.", zap.Error(err))
		return -1, err
	}
	l.Debug("Expiration calculated successfully.", zap.Int64("expiration", expireAt))

	script := redis.NewScript(`
		-- Script params:
		-- KEYS[1] is the name of the leaderboard
		-- KEYS[2] is member's public ID
		-- ARGV[1] is the member's updated score
		-- ARGV[2] is the leaderboard's expiration

		-- creates leaderboard or just sets score of member
		local res = redis.call("ZADD", KEYS[1], ARGV[1], KEYS[2])

		-- If expiration is required set expiration
		if (ARGV[2] ~= "-1") then
			local expiration = redis.call("TTL", KEYS[1])
			if (expiration == -2) then
				return redis.error_reply("Leaderboard Set was not created in ZADD! Don't know how to proceed.")
			end
			if (expiration == -1) then
				redis.call("EXPIREAT", KEYS[1], ARGV[2])
			end
		end

		-- return updated rank of member
		local rank = redis.call("ZREVRANK", KEYS[1], KEYS[2])
		return rank
	`)

	l.Debug("Updating rank for member.")
	newRank, err := script.Run(cli, []string{lb.PublicID, memberID}, score, expireAt).Result()

	if err != nil {
		l.Error("Failed to update rank for member.", zap.Error(err))
		return -1, err
	}
	if newRank == nil {
		l.Error("Member was not found.", zap.Error(err))
		return -1, NewMemberNotFound(lb.PublicID, memberID)
	}

	r := int(newRank.(int64))
	l.Info("Rank for member retrieved successfully.", zap.Int("newRank", r))
	return r, err
}

// SetMemberScore sets the score to the member with the given ID
func (lb *Leaderboard) SetMemberScore(memberID string, score int) (*Member, error) {
	l := lb.Logger.With(
		zap.String("operation", "SetMemberScore"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.String("memberID", memberID),
		zap.Int("score", score),
	)
	l.Debug("Setting member score...")

	rank, err := lb.AddToLeaderboardSet(lb.RedisClient, memberID, score)
	if err != nil {
		return nil, err
	}

	l.Info("Member score set successfully.")
	nMember := Member{PublicID: memberID, Score: score, Rank: rank + 1}
	return &nMember, err
}

// TotalMembers returns the total number of members in a given leaderboard
func (lb *Leaderboard) TotalMembers() (int, error) {
	l := lb.Logger.With(
		zap.String("operation", "TotalMembers"),
		zap.String("leaguePublicID", lb.PublicID),
	)
	cli := lb.RedisClient.Client

	l.Debug("Retrieving total members...")
	total, err := cli.ZCard(lb.PublicID).Result()
	if err != nil {
		l.Error("Retrieval of total members failed.", zap.Error(err))
		return 0, err
	}
	l.Info("Total members of leaderboard retrieved successfully.")
	return int(total), nil
}

// RemoveMember removes the member with the given publicID from the leaderboard
func (lb *Leaderboard) RemoveMember(memberID string) (*Member, error) {
	l := lb.Logger.With(
		zap.String("operation", "RemoveMember"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.String("memberID", memberID),
	)

	cli := lb.RedisClient.Client

	l.Debug("Removing member from leaderboard...")
	nMember, err := lb.GetMember(memberID)
	_, err = cli.ZRem(lb.PublicID, memberID).Result()
	if err != nil {
		l.Error("Member removal failed...", zap.Error(err))
		return nil, err
	}
	l.Info("Member removed successfully.")
	return nMember, err
}

// TotalPages returns the number of pages of the leaderboard
func (lb *Leaderboard) TotalPages() (int, error) {
	l := lb.Logger.With(
		zap.String("operation", "TotalPages"),
		zap.String("leaguePublicID", lb.PublicID),
	)

	cli := lb.RedisClient.Client

	l.Debug("Retrieving number of pages for leaderboard.")
	pages := 0
	total, err := cli.ZCard(lb.PublicID).Result()
	if err != nil {
		l.Error("Number of pages could not be retrieved.", zap.Error(err))
		return 0, err
	}
	pages = int(math.Ceil(float64(total) / float64(lb.PageSize)))
	l.Info("Number of pages for leaderboard retrieved successfully.", zap.Int("numberOfPages", pages))
	return pages, nil
}

// GetMember returns the score and the rank of the member with the given ID
func (lb *Leaderboard) GetMember(memberID string) (*Member, error) {
	l := lb.Logger.With(
		zap.String("operation", "GetMember"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.String("memberID", memberID),
	)

	cli := lb.RedisClient.Client

	l.Debug("Getting member information...")
	script := redis.NewScript(`
		-- Script params:
		-- KEYS[1] is the name of the leaderboard
		-- KEYS[2] is member's public ID

		-- gets rank of the member
		local rank = redis.call("ZREVRANK", KEYS[1], KEYS[2])
		local score = redis.call("ZSCORE", KEYS[1], KEYS[2])

		return {rank,score}
	`)

	result, err := script.Run(cli, []string{lb.PublicID, memberID}).Result()
	if err != nil {
		l.Error("Getting member information failed.", zap.Error(err))
		return nil, err
	}

	res := result.([]interface{})

	if res[0] == nil || res[1] == nil {
		l.Error("Could not find member.", zap.Error(err))
		return nil, NewMemberNotFound(lb.PublicID, memberID)
	}

	rank := int(res[0].(int64))
	scoreParsed, _ := strconv.ParseInt(res[1].(string), 10, 32)
	score := int(scoreParsed)

	l.Info("Member information found.", zap.Int("rank", rank), zap.Int("score", score))
	nMember := Member{PublicID: memberID, Score: score, Rank: rank + 1}
	return &nMember, nil
}

// GetAroundMe returns a page of results centered in the member with the given ID
func (lb *Leaderboard) GetAroundMe(memberID string) ([]*Member, error) {
	l := lb.Logger.With(
		zap.String("operation", "GetAroundMe"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.String("memberID", memberID),
	)

	l.Debug("Getting information about members around a specific member...")
	currentMember, err := lb.GetMember(memberID)
	if err != nil {
		return nil, err
	}
	startOffset := currentMember.Rank - (lb.PageSize / 2)
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + lb.PageSize) - 1

	totalMembers, err := lb.TotalMembers()
	if err != nil {
		return nil, err
	}
	if totalMembers < endOffset {
		endOffset = totalMembers
		startOffset = endOffset - lb.PageSize
		if startOffset < 0 {
			startOffset = 0
		}
	}

	members, err := getMembersByRange(lb.RedisClient, lb.PublicID, lb.PageSize, startOffset, endOffset, l)
	if err != nil {
		l.Error("Failed to retrieve information around a specific member.", zap.Error(err))
		return nil, err
	}
	l.Info("Retrieved information around member successfully.")
	return members, nil
}

// GetRank returns the rank of the member with the given ID
func (lb *Leaderboard) GetRank(memberID string) (int, error) {
	l := lb.Logger.With(
		zap.String("operation", "GetRank"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.String("memberID", memberID),
	)

	cli := lb.RedisClient.Client

	l.Debug("Getting rank of specific member...")
	rank, err := cli.ZRevRank(lb.PublicID, memberID).Result()
	if err != nil {
		if strings.HasPrefix(err.Error(), "redis: nil") {
			l.Error("Member was not found in specified leaderboard.", zap.Error(err))
			return -1, NewMemberNotFound(lb.PublicID, memberID)
		}

		l.Error("Failed to retrieve rank of specific member.", zap.Error(err))
		return -1, err
	}
	l.Info("Rank retrieval succeeded.")
	return int(rank + 1), nil
}

// GetLeaders returns a page of members with rank and score
func (lb *Leaderboard) GetLeaders(page int) ([]*Member, error) {
	l := lb.Logger.With(
		zap.String("operation", "GetLeaders"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.Int("page", page),
	)

	if page < 1 {
		page = 1
	}
	totalPages, err := lb.TotalPages()
	if err != nil {
		return nil, err
	}
	if page > totalPages {
		return make([]*Member, 0), nil
	}
	redisIndex := page - 1
	startOffset := redisIndex * lb.PageSize
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + lb.PageSize) - 1
	return getMembersByRange(lb.RedisClient, lb.PublicID, lb.PageSize, startOffset, endOffset, l)
}

//GetTopPercentage of members in the leaderboard.
func (lb *Leaderboard) GetTopPercentage(amount, maxMembers int) ([]*Member, error) {
	l := lb.Logger.With(
		zap.String("operation", "GetTopPercentage"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.Int("amount", amount),
	)

	if amount < 1 || amount > 100 {
		err := fmt.Errorf("Percentage must be a valid integer between 1 and 100.")
		l.Error(err.Error(), zap.Error(err))
		return nil, err
	}

	cli := lb.RedisClient.Client

	l.Debug("Getting top percentage of members...")
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

		local members = redis.call("ZREVRANGE", KEYS[1], 0, numberOfMembers - 1, "WITHSCORES")
		local fullMembers = {}

		for index=1, #members, 2 do
			local publicID = members[index]
			local score = members[index + 1]
		 	local rank = redis.call("ZREVRANK", KEYS[1], publicID)

			table.insert(fullMembers, publicID)
			table.insert(fullMembers, rank)
			table.insert(fullMembers, score)
		end

		return fullMembers
	`)

	result, err := script.Run(cli, []string{lb.PublicID}, float64(amount)/100.0, maxMembers).Result()

	if err != nil {
		l.Error("Getting top percentage of members failed.", zap.Error(err))
		return nil, err
	}

	res := result.([]interface{})
	members := []*Member{}

	for i := 0; i < len(res); i += 3 {
		memberPublicID := res[i].(string)

		rank := int(res[i+1].(int64)) + 1
		s, _ := strconv.ParseInt(res[i+2].(string), 10, 32)
		score := int(s)

		members = append(members, &Member{
			PublicID: memberPublicID,
			Score:    score,
			Rank:     rank,
		})
	}

	l.Info("Top percentage of members retrieved successfully.")

	return members, nil
}
