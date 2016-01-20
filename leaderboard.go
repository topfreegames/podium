package leaderboard

import (
	"fmt"
	"github.com/garyburd/redigo"
	"math"
)

/* Structs model */
type User struct {
	name  string
	score int
	rank  int
}

type Team struct {
	name    string
	members map[string]User
	rank    int
}

type Leaderboard struct {
	host     string
	name     string
	pageSize int
}

/* End Structs model */

var pool *redis.Pool

/* Private functions */

func getConnection(host string) redis.Conn {
	if pool == nil {
		srv := host
		pool = redis.NewPool(func() (redis.Conn, error) {
			return redis.Dial("tcp", srv)
		}, 10)
	}
	return pool.Get()
}

func getMembersByRange(host string, leaderboard string, pageSize int, startOffset int, endOffset int) []User {
	conn := getConnection(host)
	defer conn.Close()
	users := make([]User, pageSize)
	values, _ := redis.Values(conn.Do("ZREVRANGE", leaderboard, startOffset, endOffset, "WITHSCORES"))
	var i = 0
	for len(values) > 0 {
		name := ""
		score := -1
		values, _ = redis.Scan(values, &name, &score)
		rank, _ := redis.Int(conn.Do("ZREVRANK", leaderboard, name))
		nUser := User{name: name, score: score, rank: rank + 1}
		users[i] = nUser
		i += 1
	}
	return users
}

/* End Private functions */

/* Public functions */

func NewLeaderboard(host string, name string, pageSize int) Leaderboard {
	l := Leaderboard{host: host, name: name, pageSize: pageSize}
	return l
}

func (l *Leaderboard) RankMember(username string, score int) (User, error) {
	conn := getConnection(l.host)
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
	nUser := User{name: username, score: score, rank: rank + 1}
	return nUser, err
}

func (l *Leaderboard) TotalMembers() int {
	conn := getConnection(l.host)
	defer conn.Close()
	total, err := redis.Int(conn.Do("ZCARD", l.name))
	if err != nil {
		fmt.Printf("error on get leaderboard total members")
		return 0
	}
	return total
}

func (l *Leaderboard) RemoveMember(username string) (User, error) {
	conn := getConnection(l.host)
	defer conn.Close()
	nUser, err := l.GetMember(username)
	_, err = conn.Do("ZREM", l.name, username)
	if err != nil {
		fmt.Printf("error on remove user from leaderboard")
	}
	return nUser, err
}

func (l *Leaderboard) TotalPages() int {
	conn := getConnection(l.host)
	defer conn.Close()
	pages := 0
	total, err := redis.Int(conn.Do("ZCOUNT", l.name, "-inf", "+inf"))
	if err == nil {
		pages = int(math.Ceil(float64(total) / float64(l.pageSize)))
	}
	return pages
}

func (l *Leaderboard) GetMember(username string) (User, error) {
	conn := getConnection(l.host)
	defer conn.Close()
	rank, err := redis.Int(conn.Do("ZREVRANK", l.name, username))
	if err != nil {
		rank = 0
	}
	score, err := redis.Int(conn.Do("ZSCORE", l.name, username))
	if err != nil {
		score = 0
	}
	nUser := User{name: username, score: score, rank: rank + 1}
	return nUser, err
}

func (l *Leaderboard) GetAroundMe(username string) []User {
	currentUser, _ := l.GetMember(username)
	startOffset := currentUser.rank - (l.pageSize / 2)
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + l.pageSize) - 1
	return getMembersByRange(l.name, l.pageSize, startOffset, endOffset)
}

func (l *Leaderboard) GetRank(username string) int {
	conn := getConnection(l.host)
	defer conn.Close()
	rank, _ := redis.Int(conn.Do("ZREVRANK", l.name, username))
	return rank + 1
}

func (l *Leaderboard) GetLeaders(page int) []User {
	if page < 1 {
		page = 1
	}
	if page > l.TotalPages() {
		page = l.TotalPages()
	}
	redisIndex := page - 1
	startOffset := redisIndex * l.pageSize
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + l.pageSize) - 1

	return getMembersByRange(l.name, l.pageSize, startOffset, endOffset)
}

func (l *Leaderboard) GetMemberByRank(position int) User {
	conn := getConnection(l.host)
	defer conn.Close()
	if position <= l.TotalMembers() {
		currentPage := int(math.Ceil(float64(position) / float64(l.pageSize)))
		offset := (position - 1) % l.pageSize
		leaders := l.GetLeaders(currentPage)
		if leaders[offset].rank == position {
			return leaders[offset]
		}
	}
	return User{}
}

/* End Public functions */
