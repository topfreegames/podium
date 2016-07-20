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

	"github.com/garyburd/redigo/redis"
	"github.com/topfreegames/podium/util"
)

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
	RedisClient *util.RedisClient
	PublicID    string
	PageSize    int
}

func getMembersByRange(redisClient *util.RedisClient, leaderboard string, pageSize int, startOffset int, endOffset int) ([]User, error) {
	conn := redisClient.GetConnection()
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZREVRANGE", leaderboard, startOffset, endOffset, "WITHSCORES"))
	if err != nil {
		return nil, err
	}
	users := make([]User, len(values)/2)
	var i = 0
	for len(values) > 0 {
		publicID := ""
		score := -1
		// Scan returns the slice of src following the copied values.
		values, err = redis.Scan(values, &publicID, &score)
		if err != nil {
			return nil, err
		}
		rank, err := redis.Int(conn.Do("ZREVRANK", leaderboard, publicID))
		if err != nil {
			return nil, err
		}
		nUser := User{PublicID: publicID, Score: score, Rank: rank + 1}
		users[i] = nUser
		i++
	}
	return users, nil
}

// NewLeaderboard creates a new Leaderboard with given settings, ID and pageSize
func NewLeaderboard(redisClient *util.RedisClient, publicID string, pageSize int) *Leaderboard {
	return &Leaderboard{RedisClient: redisClient, PublicID: publicID, PageSize: pageSize}
}

//AddToLeagueSet adds a score to a league set respecting expiration
func (l *Leaderboard) AddToLeagueSet(redisCli redis.Conn, userID string, score int) (int, error) {
	expireAt, err := util.GetExpireAt(l.PublicID)
	if err != nil {
		return -1, err
	}
	script := redis.NewScript(1, `
		-- Script params:
		-- KEYS[1] is the name of the league
		-- ARGV[1] is user's public ID
		-- ARGV[2] is the user's updated score
		-- ARGV[3] is the league's expiration

		-- creates league or just sets score of member
		local res = redis.call("ZADD", KEYS[1], ARGV[2], ARGV[1])

		-- If expiration is required set expiration
		if (ARGV[3] ~= "-1") then
			local expiration = redis.call("TTL", KEYS[1])
			if (expiration == -2) then
				return redis.error_reply("League Set was not created in ZADD! Don't know how to proceed.")
			end
			if (expiration == -1) then
				redis.call("EXPIREAT", KEYS[1], ARGV[3])
			end
		end

		-- return updated rank of player
		local rank = redis.call("ZREVRANK", KEYS[1], ARGV[1])
		return rank
	`)

	newRank, err := script.Do(redisCli, l.PublicID, userID, score, expireAt)
	return int(newRank.(int64)), err
}

// SetUserScore sets the score to the user with the given ID
func (l *Leaderboard) SetUserScore(userID string, score int) (*User, error) {
	conn := l.RedisClient.GetConnection()
	defer conn.Close()

	rank, err := l.AddToLeagueSet(conn, userID, score)
	if err != nil {
		return nil, err
	}

	nUser := User{PublicID: userID, Score: score, Rank: rank + 1}
	return &nUser, err
}

// TotalMembers returns the total number of members in a given leaderboard
func (l *Leaderboard) TotalMembers() (int, error) {
	conn := l.RedisClient.GetConnection()
	defer conn.Close()
	total, err := redis.Int(conn.Do("ZCARD", l.PublicID))
	if err != nil {
		fmt.Printf("error on get leaderboard total members")
		return 0, err
	}
	return total, nil
}

// RemoveMember removes the member with the given publicID from the leaderboard
func (l *Leaderboard) RemoveMember(userID string) (*User, error) {
	conn := l.RedisClient.GetConnection()
	defer conn.Close()
	nUser, err := l.GetMember(userID)
	_, err = conn.Do("ZREM", l.PublicID, userID)
	if err != nil {
		fmt.Printf("error on remove user from leaderboard")
		return nil, err
	}
	return nUser, err
}

// TotalPages returns the number of pages of the leaderboard
func (l *Leaderboard) TotalPages() (int, error) {
	conn := l.RedisClient.GetConnection()
	defer conn.Close()
	pages := 0
	total, err := redis.Int(conn.Do("ZCARD", l.PublicID))
	if err != nil {
		fmt.Printf("error on get leaderboard total pages")
		return 0, err
	}
	pages = int(math.Ceil(float64(total) / float64(l.PageSize)))
	return pages, nil
}

// GetMember returns the score and the rank of the user with the given ID
func (l *Leaderboard) GetMember(userID string) (*User, error) {
	conn := l.RedisClient.GetConnection()
	defer conn.Close()
	rank, err := redis.Int(conn.Do("ZREVRANK", l.PublicID, userID))
	if err != nil {
		return nil, err
	}
	score, err := redis.Int(conn.Do("ZSCORE", l.PublicID, userID))
	if err != nil {
		return nil, err
	}
	nUser := User{PublicID: userID, Score: score, Rank: rank + 1}
	return &nUser, err
}

// GetAroundMe returns a page of results centered in the user with the given ID
func (l *Leaderboard) GetAroundMe(userID string) ([]User, error) {
	currentUser, err := l.GetMember(userID)
	if err != nil {
		return nil, err
	}
	startOffset := currentUser.Rank - (l.PageSize / 2)
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + l.PageSize) - 1
	return getMembersByRange(l.RedisClient, l.PublicID, l.PageSize, startOffset, endOffset)
}

// GetRank returns the rank of the user with the given ID
func (l *Leaderboard) GetRank(userID string) (int, error) {
	conn := l.RedisClient.GetConnection()
	defer conn.Close()
	rank, err := redis.Int(conn.Do("ZREVRANK", l.PublicID, userID))
	if err != nil {
		rank = -1
	}
	return rank + 1, err
}

// GetLeaders returns a page of users with rank and score
func (l *Leaderboard) GetLeaders(page int) ([]User, error) {
	if page < 1 {
		page = 1
	}
	totalPages, err := l.TotalPages()
	if err != nil {
		return nil, err
	}
	if page > totalPages {
		return make([]User, 0), nil
	}
	redisIndex := page - 1
	startOffset := redisIndex * l.PageSize
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + l.PageSize) - 1
	return getMembersByRange(l.RedisClient, l.PublicID, l.PageSize, startOffset, endOffset)
}
