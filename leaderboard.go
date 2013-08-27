package leaderboard

import (
	"fmt"
	"math"
	"github.com/garyburd/redigo/redis"
)

type User struct {
	name string
	score int
	rank int

}

type Team struct {
	name string
	members map[string]User
	rank int
}

type Leaderboard struct {
	name string
	pageSize int
}

var pool *redis.Pool

func getConnection() redis.Conn {
	if pool == nil {
		srv := "localhost:6379"
		pool = redis.NewPool(func() (redis.Conn, error) {
			return redis.Dial("tcp", srv)
		}, 10)
	}
	return pool.Get()
}

func NewLeaderboard(name string, pageSize int) Leaderboard {
	l := Leaderboard{name: name, pageSize: pageSize}
	return l
}

func (l *Leaderboard) RankMember(username string, score int) (User, error) {
	conn := getConnection()
	defer conn.Close()
	_, err := conn.Do("ZADD", l.name, score, username)
	if err != nil {
		fmt.Printf("error on store in redis in rankMember Leaderboard:%s - Username:%s - Score:%d", l.name, username, score)
	}
	rank, err := redis.Int(conn.Do("ZREVRANK", l.name, username))
	if err != nil {
		fmt.Printf("error on get user rank Leaderboard:%s - Username:%s", l.name, username)
		rank = -1
	}
	nUser := User{name: username, score: score, rank: rank + 1 }
	return nUser, err
}

func (l *Leaderboard) TotalMembers() int {
	conn := getConnection()
	defer conn.Close()
	total, err := redis.Int(conn.Do("ZCARD", l.name))
	if err != nil {
		fmt.Printf("error on get leaderboard total members")
		return 0
	}
	return total
}

func (l *Leaderboard) RemoveMember(username string) (User, error) {
	conn := getConnection()
	defer conn.Close()
	nUser, err := l.GetMember(username)
	_, err = conn.Do("ZREM", l.name, username)
	if err != nil {
		fmt.Printf("error on remove user from leaderboard")
	}
	return nUser, err
}

func (l *Leaderboard) TotalPages() int {
	conn := getConnection()
	defer conn.Close()
	pages := 0
	total, err := redis.Int(conn.Do("ZCOUNT", l.name, "-inf", "+inf"))
	if err == nil {
		pages = int(math.Ceil(float64(total) / float64(l.pageSize)))
	}
	return pages
}

func (l *Leaderboard) GetMember(username string) (User, error) {
	conn := getConnection()
	defer conn.Close()
	rank, err := redis.Int(conn.Do("ZREVRANK", l.name, username))
	if err != nil {
		rank = 0
	}
	score, err := redis.Int(conn.Do("ZSCORE", l.name, username))
	if err != nil {
		score = 0
	}
	nUser := User{name: username, score: score, rank: rank+1}
	return nUser, err
}

func (l *Leaderboard) GetAroundMe(username string) []User {
	conn := getConnection()
	defer conn.Close()
	currentUser, _ := l.GetMember(username)
	startOffset := currentUser.rank - (l.pageSize / 2)
    if startOffset < 0 {
    	startOffset = 0
    }

    endOffset := (startOffset + l.pageSize) - 1
	users := make([]User, l.pageSize)
    values, _ := redis.Values(conn.Do("ZREVRANGE", l.name, startOffset, endOffset, "WITHSCORES"))
    var i = 0
    for len(values) > 0 {
    	name := ""
    	score := -1
    	values, _ = redis.Scan(values, &name, &score)
    	rank, _ := redis.Int(conn.Do("ZREVRANK", l.name, name))
    	nUser := User{name: name, score: score, rank: rank + 1}
    	users[i] = nUser
    	i+= 1
    }
    return users
}

func (l *Leaderboard) GetRank(username string) int {
	conn := getConnection()
	defer conn.Close()
	rank, _ := redis.Int(conn.Do("ZREVRANK", l.name, username))
	return rank + 1
}