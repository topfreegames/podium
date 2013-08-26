package leaderboard

import (
	"fmt"
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

func (l *Leaderboard) rankMember(username string, score int) (User, error) {
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
	nUser := User{name: username, score: score, rank: rank}
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
