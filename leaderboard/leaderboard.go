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

	"github.com/garyburd/redigo/redis"
	"github.com/topfreegames/podium/util"
	"github.com/uber-go/zap"
)

//UserNotFoundError indicates user was not found in Redis
type UserNotFoundError struct {
	LeaderboardID string
	UserID        string
}

func (e *UserNotFoundError) Error() string {
	return fmt.Sprintf("Could not find data for user %s in leaderboard %s.", e.UserID, e.LeaderboardID)
}

//NewUserNotFound returns a new error for user not found
func NewUserNotFound(leaderboardID, userID string) *UserNotFoundError {
	return &UserNotFoundError{
		LeaderboardID: leaderboardID,
		UserID:        userID,
	}
}

// User maps an user identified by their publicID to their score and rank
type User struct {
	PublicID string
	Score    int
	Rank     int
}

// Team groups sets of users
type Team struct {
	PublicID string
	Members  map[string]User
	Rank     int
}

// Leaderboard identifies a leaderboard with given redis client
type Leaderboard struct {
	Logger      zap.Logger
	RedisClient *util.RedisClient
	PublicID    string
	PageSize    int
}

func getMembersByRange(redisClient *util.RedisClient, leaderboard string, pageSize int, startOffset int, endOffset int, l zap.Logger) ([]User, error) {
	conn := redisClient.GetConnection()
	defer conn.Close()

	l.Debug("Getting top leaderboard members...")
	values, err := redis.Values(conn.Do("ZREVRANGE", leaderboard, startOffset, endOffset, "WITHSCORES"))
	if err != nil {
		l.Error("Retrieval of leaderboard top members failed.", zap.Error(err))
		return nil, err
	}
	l.Info("Retrieval of leaderboard top members succeeded.")

	l.Debug("Retrieving details of leaderboard top members...")
	users := make([]User, len(values)/2)
	var i = 0
	for len(values) > 0 {
		publicID := ""
		score := -1
		// Scan returns the slice of src following the copied values.
		values, err = redis.Scan(values, &publicID, &score)
		if err != nil {
			l.Error("Retrieval of leaderboard top members failed.", zap.Error(err))
			return nil, err
		}
		rank, err := redis.Int(conn.Do("ZREVRANK", leaderboard, publicID))
		if err != nil {
			l.Error("Retrieval of leaderboard top members rank failed.", zap.Error(err))
			return nil, err
		}
		nUser := User{PublicID: publicID, Score: score, Rank: rank + 1}
		users[i] = nUser
		i++
	}
	l.Info("Retrieval of leaderboard top members' details succeeded.")
	return users, nil
}

// NewLeaderboard creates a new Leaderboard with given settings, ID and pageSize
func NewLeaderboard(redisClient *util.RedisClient, publicID string, pageSize int, logger zap.Logger) *Leaderboard {
	return &Leaderboard{RedisClient: redisClient, PublicID: publicID, PageSize: pageSize, Logger: logger}
}

//AddToLeaderboardSet adds a score to a leaderboard set respecting expiration
func (lb *Leaderboard) AddToLeaderboardSet(redisCli redis.Conn, userID string, score int) (int, error) {
	l := lb.Logger.With(
		zap.String("operation", "AddToLeaderboardSet"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.String("userID", userID),
		zap.Int("score", score),
	)

	l.Debug("Calculating expiration for leaderboard...")
	expireAt, err := util.GetExpireAt(lb.PublicID)
	if err != nil {
		l.Error("Could not get expiration.", zap.Error(err))
		return -1, err
	}
	l.Debug("Expiration calculated successfully.", zap.Int64("expiration", expireAt))

	script := redis.NewScript(1, `
		-- Script params:
		-- KEYS[1] is the name of the leaderboard
		-- ARGV[1] is user's public ID
		-- ARGV[2] is the user's updated score
		-- ARGV[3] is the leaderboard's expiration

		-- creates leaderboard or just sets score of member
		local res = redis.call("ZADD", KEYS[1], ARGV[2], ARGV[1])

		-- If expiration is required set expiration
		if (ARGV[3] ~= "-1") then
			local expiration = redis.call("TTL", KEYS[1])
			if (expiration == -2) then
				return redis.error_reply("Leaderboard Set was not created in ZADD! Don't know how to proceed.")
			end
			if (expiration == -1) then
				redis.call("EXPIREAT", KEYS[1], ARGV[3])
			end
		end

		-- return updated rank of user
		local rank = redis.call("ZREVRANK", KEYS[1], ARGV[1])
		return rank
	`)

	l.Debug("Updating rank for user.")
	newRank, err := script.Do(redisCli, lb.PublicID, userID, score, expireAt)
	if err != nil {
		l.Error("Failed to update rank for user.", zap.Error(err))
		return -1, err
	}
	if newRank == nil {
		l.Error("User was not found.", zap.Error(err))
		return -1, NewUserNotFound(lb.PublicID, userID)
	}

	r := int(newRank.(int64))
	l.Info("Rank for user retrieved successfully.", zap.Int("newRank", r))
	return r, err
}

// SetUserScore sets the score to the user with the given ID
func (lb *Leaderboard) SetUserScore(userID string, score int) (*User, error) {
	l := lb.Logger.With(
		zap.String("operation", "SetUserScore"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.String("userID", userID),
		zap.Int("score", score),
	)
	conn := lb.RedisClient.GetConnection()
	defer conn.Close()

	l.Debug("Setting user score...")

	rank, err := lb.AddToLeaderboardSet(conn, userID, score)
	if err != nil {
		return nil, err
	}

	l.Info("User score set successfully.")
	nUser := User{PublicID: userID, Score: score, Rank: rank + 1}
	return &nUser, err
}

// TotalMembers returns the total number of members in a given leaderboard
func (lb *Leaderboard) TotalMembers() (int, error) {
	l := lb.Logger.With(
		zap.String("operation", "TotalMembers"),
		zap.String("leaguePublicID", lb.PublicID),
	)
	conn := lb.RedisClient.GetConnection()
	defer conn.Close()

	l.Debug("Retrieving total members...")
	total, err := redis.Int(conn.Do("ZCARD", lb.PublicID))
	if err != nil {
		l.Error("Retrieval of total members failed.", zap.Error(err))
		return 0, err
	}
	l.Info("Total members of leaderboard retrieved successfully.")
	return total, nil
}

// RemoveMember removes the member with the given publicID from the leaderboard
func (lb *Leaderboard) RemoveMember(userID string) (*User, error) {
	l := lb.Logger.With(
		zap.String("operation", "RemoveMember"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.String("userID", userID),
	)

	conn := lb.RedisClient.GetConnection()
	defer conn.Close()

	l.Debug("Removing member from leaderboard...")
	nUser, err := lb.GetMember(userID)
	_, err = conn.Do("ZREM", lb.PublicID, userID)
	if err != nil {
		l.Error("Member removal failed...", zap.Error(err))
		return nil, err
	}
	l.Info("Member removed successfully.")
	return nUser, err
}

// TotalPages returns the number of pages of the leaderboard
func (lb *Leaderboard) TotalPages() (int, error) {
	l := lb.Logger.With(
		zap.String("operation", "TotalPages"),
		zap.String("leaguePublicID", lb.PublicID),
	)

	conn := lb.RedisClient.GetConnection()
	defer conn.Close()

	l.Debug("Retrieving number of pages for leaderboard.")
	pages := 0
	total, err := redis.Int(conn.Do("ZCARD", lb.PublicID))
	if err != nil {
		l.Error("Number of pages could not be retrieved.", zap.Error(err))
		return 0, err
	}
	pages = int(math.Ceil(float64(total) / float64(lb.PageSize)))
	l.Info("Number of pages for leaderboard retrieved successfully.", zap.Int("numberOfPages", pages))
	return pages, nil
}

// GetMember returns the score and the rank of the user with the given ID
func (lb *Leaderboard) GetMember(userID string) (*User, error) {
	l := lb.Logger.With(
		zap.String("operation", "GetMember"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.String("userID", userID),
	)

	conn := lb.RedisClient.GetConnection()
	defer conn.Close()

	l.Debug("Getting member information...")
	script := redis.NewScript(1, `
		-- Script params:
		-- KEYS[1] is the name of the leaderboard
		-- ARGV[1] is user's public ID

		-- gets rank of the user
		local rank = redis.call("ZREVRANK", KEYS[1], ARGV[1])
		local score = redis.call("ZSCORE", KEYS[1], ARGV[1])

		return {rank,score}
	`)

	result, err := script.Do(conn, lb.PublicID, userID)
	if err != nil {
		l.Error("Getting member information failed.", zap.Error(err))
		return nil, err
	}

	res := result.([]interface{})

	if res[0] == nil || res[1] == nil {
		l.Error("Could not find user.", zap.Error(err))
		return nil, NewUserNotFound(lb.PublicID, userID)
	}

	rank := int(res[0].(int64))
	scoreParsed, _ := strconv.ParseInt(string(res[1].([]byte)), 10, 32)
	score := int(scoreParsed)

	l.Info("User information found.", zap.Int("rank", rank), zap.Int("score", score))
	nUser := User{PublicID: userID, Score: score, Rank: rank + 1}
	return &nUser, nil
}

// GetAroundMe returns a page of results centered in the user with the given ID
func (lb *Leaderboard) GetAroundMe(userID string) ([]User, error) {
	l := lb.Logger.With(
		zap.String("operation", "GetAroundMe"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.String("userID", userID),
	)

	l.Debug("Getting information about users around a specific user...")
	currentUser, err := lb.GetMember(userID)
	if err != nil {
		return nil, err
	}
	startOffset := currentUser.Rank - (lb.PageSize / 2)
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + lb.PageSize) - 1

	members, err := getMembersByRange(lb.RedisClient, lb.PublicID, lb.PageSize, startOffset, endOffset, l)
	if err != nil {
		l.Error("Failed to retrieve information around a specific user.", zap.Error(err))
		return nil, err
	}
	l.Info("Retrieved information around member successfully.")
	return members, nil
}

// GetRank returns the rank of the user with the given ID
func (lb *Leaderboard) GetRank(userID string) (int, error) {
	l := lb.Logger.With(
		zap.String("operation", "GetRank"),
		zap.String("leaguePublicID", lb.PublicID),
		zap.String("userID", userID),
	)

	conn := lb.RedisClient.GetConnection()
	defer conn.Close()

	l.Debug("Getting rank of specific user...")
	rank, err := redis.Int(conn.Do("ZREVRANK", lb.PublicID, userID))
	if err != nil {
		if strings.HasPrefix(err.Error(), "redigo: nil returned") {
			l.Error("User was not found in specified leaderboard.", zap.Error(err))
			return -1, NewUserNotFound(lb.PublicID, userID)
		}

		l.Error("Failed to retrieve rank of specific user.", zap.Error(err))
		return -1, err
	}
	l.Info("Rank retrieval succeeded.")
	return rank + 1, nil
}

// GetLeaders returns a page of users with rank and score
func (lb *Leaderboard) GetLeaders(page int) ([]User, error) {
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
		return make([]User, 0), nil
	}
	redisIndex := page - 1
	startOffset := redisIndex * lb.PageSize
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + lb.PageSize) - 1
	return getMembersByRange(lb.RedisClient, lb.PublicID, lb.PageSize, startOffset, endOffset, l)
}
