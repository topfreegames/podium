package leaderboard

import (
	"fmt"
	"math"

	"github.com/garyburd/redigo/redis"
	"github.com/topfreegames/go-leaderboard/util"
)

// User does something
type User struct {
	Name  string
	Score int
	Rank  int
}

// Team does something else
type Team struct {
	Name    string
	Members map[string]User
	Rank    int
}

// Leaderboard does something else else
type Leaderboard struct {
	Settings util.RedisSettings
	Name     string
	PageSize int
}

func getMembersByRange(settings util.RedisSettings, leaderboard string, pageSize int, startOffset int, endOffset int) []User {
	conn := util.GetConnection(settings)
	defer conn.Close()
	users := make([]User, pageSize)
	values, _ := redis.Values(conn.Do("ZREVRANGE", leaderboard, startOffset, endOffset, "WITHSCORES"))
	var i = 0
	for len(values) > 0 {
		name := ""
		score := -1
		values, _ = redis.Scan(values, &name, &score)
		rank, _ := redis.Int(conn.Do("ZREVRANK", leaderboard, name))
		nUser := User{Name: name, Score: score, Rank: rank + 1}
		users[i] = nUser
		i++
	}
	return users
}

// NewLeaderboard does something i dont know yet
func NewLeaderboard(settings util.RedisSettings, name string, pageSize int) Leaderboard {
	l := Leaderboard{Settings: settings, Name: name, PageSize: pageSize}
	return l
}

// RankMember does something i dont know yet
func (l *Leaderboard) RankMember(username string, score int) (User, error) {
	conn := util.GetConnection(l.Settings)
	defer conn.Close()
	_, err := conn.Do("ZADD", l.Name, score, username)
	if err != nil {
		fmt.Printf("error on store in redis in rankMember Leaderboard:%s - Username:%s - Score:%d", l.Name, username, score)
	}
	rank, err := redis.Int(conn.Do("ZREVRANK", l.Name, username))
	if err != nil {
		fmt.Printf("error on get user rank Leaderboard:%s - Username:%s", l.Name, username)
		rank = -1
	}
	nUser := User{Name: username, Score: score, Rank: rank + 1}
	return nUser, err
}

// TotalMembers does something i dont know yet
func (l *Leaderboard) TotalMembers() int {
	conn := util.GetConnection(l.Settings)
	total, err := redis.Int(conn.Do("ZCARD", l.Name))
	if err != nil {
		fmt.Printf("error on get leaderboard total members")
		return 0
	}
	defer conn.Close()
	return total
}

// RemoveMember does something i dont know yet
func (l *Leaderboard) RemoveMember(username string) (User, error) {
	conn := util.GetConnection(l.Settings)
	nUser, err := l.GetMember(username)
	_, err = conn.Do("ZREM", l.Name, username)
	if err != nil {
		fmt.Printf("error on remove user from leaderboard")
	}
	defer conn.Close()
	return nUser, err
}

// TotalPages does something i dont know yet
func (l *Leaderboard) TotalPages() int {
	conn := util.GetConnection(l.Settings)
	pages := 0
	total, err := redis.Int(conn.Do("ZCOUNT", l.Name, "-inf", "+inf"))
	if err == nil {
		pages = int(math.Ceil(float64(total) / float64(l.PageSize)))
	}
	defer conn.Close()
	return pages
}

// GetMember does something i dont know yet
func (l *Leaderboard) GetMember(username string) (User, error) {
	conn := util.GetConnection(l.Settings)
	rank, err := redis.Int(conn.Do("ZREVRANK", l.Name, username))
	if err != nil {
		rank = 0
	}
	score, err := redis.Int(conn.Do("ZSCORE", l.Name, username))
	if err != nil {
		score = 0
	}
	defer conn.Close()
	nUser := User{Name: username, Score: score, Rank: rank + 1}
	return nUser, err
}

// GetAroundMe does something i dont know yet
func (l *Leaderboard) GetAroundMe(username string) []User {
	currentUser, _ := l.GetMember(username)
	startOffset := currentUser.Rank - (l.PageSize / 2)
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + l.PageSize) - 1
	return getMembersByRange(l.Settings, l.Name, l.PageSize, startOffset, endOffset)
}

// GetRank does something i dont know yet
func (l *Leaderboard) GetRank(username string) int {
	conn := util.GetConnection(l.Settings)
	rank, _ := redis.Int(conn.Do("ZREVRANK", l.Name, username))
	defer conn.Close()
	return rank + 1
}

// GetLeaders does something i dont know yet
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
	return getMembersByRange(l.Settings, l.Name, l.PageSize, startOffset, endOffset)
}

// GetMemberByRank does something i dont know yet
func (l *Leaderboard) GetMemberByRank(position int) User {
	conn := util.GetConnection(l.Settings)

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
