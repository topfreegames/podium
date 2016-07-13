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

func getMembersByRange(redisClient *util.RedisClient, leaderboard string, pageSize int, startOffset int, endOffset int) []User {
	conn := redisClient.GetConnection()
	defer conn.Close()
	users := make([]User, pageSize)
	values, _ := redis.Values(conn.Do("ZREVRANGE", leaderboard, startOffset, endOffset, "WITHSCORES"))
	var i = 0
	for len(values) > 0 {
		publicID := ""
		score := -1
		values, _ = redis.Scan(values, &publicID, &score)
		rank, _ := redis.Int(conn.Do("ZREVRANK", leaderboard, publicID))
		nUser := User{PublicID: publicID, Score: score, Rank: rank + 1}
		users[i] = nUser
		i++
	}
	return users
}

// NewLeaderboard creates a new Leaderboard with given settings, ID and pageSize
func NewLeaderboard(redisClient *util.RedisClient, publicID string, pageSize int) *Leaderboard {
	return &Leaderboard{RedisClient: redisClient, PublicID: publicID, PageSize: pageSize}
}

// SetUserScore sets the score to the user with the given ID
func (l *Leaderboard) SetUserScore(userID string, score int) (User, error) {
	conn := l.RedisClient.GetConnection()
	defer conn.Close()
	_, err := conn.Do("ZADD", l.PublicID, score, userID)
	if err != nil {
		fmt.Printf("error on store in redis in SetUserScore Leaderboard:%s - UserID:%s - Score:%d", l.PublicID, userID, score)
	}
	rank, err := redis.Int(conn.Do("ZREVRANK", l.PublicID, userID))
	if err != nil {
		fmt.Printf("error on get user rank Leaderboard:%s - Username:%s", l.PublicID, userID)
		rank = -1
	}
	nUser := User{PublicID: userID, Score: score, Rank: rank + 1}
	return nUser, err
}

// TotalMembers returns the total number of members in a given leaderboard
func (l *Leaderboard) TotalMembers() int {
	conn := l.RedisClient.GetConnection()
	total, err := redis.Int(conn.Do("ZCARD", l.PublicID))
	if err != nil {
		fmt.Printf("error on get leaderboard total members")
		return 0
	}
	defer conn.Close()
	return total
}

// RemoveMember removes the member with the given publicID from the leaderboard
func (l *Leaderboard) RemoveMember(userID string) (User, error) {
	conn := l.RedisClient.GetConnection()
	nUser, err := l.GetMember(userID)
	_, err = conn.Do("ZREM", l.PublicID, userID)
	if err != nil {
		fmt.Printf("error on remove user from leaderboard")
	}
	defer conn.Close()
	return nUser, err
}

// TotalPages returns the number of pages of the leaderboard
func (l *Leaderboard) TotalPages() int {
	conn := l.RedisClient.GetConnection()
	pages := 0
	total, err := redis.Int(conn.Do("ZCOUNT", l.PublicID, "-inf", "+inf"))
	if err == nil {
		pages = int(math.Ceil(float64(total) / float64(l.PageSize)))
	}
	defer conn.Close()
	return pages
}

// GetMember returns the score and the rank of the user with the given ID
func (l *Leaderboard) GetMember(userID string) (User, error) {
	conn := l.RedisClient.GetConnection()
	rank, err := redis.Int(conn.Do("ZREVRANK", l.PublicID, userID))
	if err != nil {
		rank = 0
	}
	score, err := redis.Int(conn.Do("ZSCORE", l.PublicID, userID))
	if err != nil {
		score = 0
	}
	defer conn.Close()
	nUser := User{PublicID: userID, Score: score, Rank: rank + 1}
	return nUser, err
}

// GetAroundMe returns a page of results centered in the user with the given ID
func (l *Leaderboard) GetAroundMe(userID string) []User {
	currentUser, _ := l.GetMember(userID)
	startOffset := currentUser.Rank - (l.PageSize / 2)
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + l.PageSize) - 1
	return getMembersByRange(l.RedisClient, l.PublicID, l.PageSize, startOffset, endOffset)
}

// GetRank returns the rank of the user with the given ID
func (l *Leaderboard) GetRank(userID string) int {
	conn := l.RedisClient.GetConnection()
	rank, _ := redis.Int(conn.Do("ZREVRANK", l.PublicID, userID))
	defer conn.Close()
	return rank + 1
}

// GetLeaders returns a page of users with rank and score
func (l *Leaderboard) GetLeaders(page int) []User {
	if page < 1 {
		page = 1
	}
	if page > l.TotalPages() {
		page = l.TotalPages()
	}
	redisIndex := page - 1
	startOffset := redisIndex * l.PageSize
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + l.PageSize) - 1
	return getMembersByRange(l.RedisClient, l.PublicID, l.PageSize, startOffset, endOffset)
}

// GetMemberByRank returns a user that has the given rank
func (l *Leaderboard) GetMemberByRank(position int) User {
	conn := l.RedisClient.GetConnection()

	if position <= l.TotalMembers() {
		currentPage := int(math.Ceil(float64(position) / float64(l.PageSize)))
		offset := (position - 1) % l.PageSize
		leaders := l.GetLeaders(currentPage)
		defer conn.Close()
		if leaders[offset].Rank == position {
			return leaders[offset]
		}
	}
	defer conn.Close()
	return User{}
}
