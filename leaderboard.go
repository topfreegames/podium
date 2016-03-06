package leaderboard

import (
	"fmt"
	"math"
	"time"

	"github.com/garyburd/redigo/redis"
)

/* Structs model */
type User struct {
	Name  string
	Score int
	Rank  int
}

type Team struct {
	Name    string
	Members map[string]User
	Rank    int
}

type RedisSettings struct {
	Host     string
	Password string
}

type Leaderboard struct {
	Settings RedisSettings
	Name     string
	PageSize int
}

/* End Structs model */

var pool *redis.Pool

/* Private functions */

func newPool(server string, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func getConnection(settings RedisSettings) redis.Conn {
	if pool == nil {
		pool = newPool(settings.Host, settings.Password)
	}
	return pool.Get()
}

func getMembersByRange(settings RedisSettings, leaderboard string, pageSize int, startOffset int, endOffset int) []User {
	conn := getConnection(settings)
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
		i += 1
	}
	return users
}

/* End Private functions */

/* Public functions */

func NewLeaderboard(settings RedisSettings, name string, pageSize int) Leaderboard {
	l := Leaderboard{Settings: settings, Name: name, PageSize: pageSize}
	return l
}

func (l *Leaderboard) RankMember(username string, score int) (User, error) {
	conn := getConnection(l.Settings)
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

func (l *Leaderboard) TotalMembers() int {
	conn := getConnection(l.Settings)
	total, err := redis.Int(conn.Do("ZCARD", l.Name))
	if err != nil {
		fmt.Printf("error on get leaderboard total members")
		return 0
	}
	defer conn.Close()
	return total
}

func (l *Leaderboard) RemoveMember(username string) (User, error) {
	conn := getConnection(l.Settings)
	nUser, err := l.GetMember(username)
	_, err = conn.Do("ZREM", l.Name, username)
	if err != nil {
		fmt.Printf("error on remove user from leaderboard")
	}
	defer conn.Close()
	return nUser, err
}

func (l *Leaderboard) TotalPages() int {
	conn := getConnection(l.Settings)
	pages := 0
	total, err := redis.Int(conn.Do("ZCOUNT", l.Name, "-inf", "+inf"))
	if err == nil {
		pages = int(math.Ceil(float64(total) / float64(l.PageSize)))
	}
	defer conn.Close()
	return pages
}

func (l *Leaderboard) GetMember(username string) (User, error) {
	conn := getConnection(l.Settings)
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

func (l *Leaderboard) GetAroundMe(username string) []User {
	currentUser, _ := l.GetMember(username)
	startOffset := currentUser.Rank - (l.PageSize / 2)
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + l.PageSize) - 1
	return getMembersByRange(l.Settings, l.Name, l.PageSize, startOffset, endOffset)
}

func (l *Leaderboard) GetRank(username string) int {
	conn := getConnection(l.Settings)
	rank, _ := redis.Int(conn.Do("ZREVRANK", l.Name, username))
	defer conn.Close()
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
	startOffset := redisIndex * l.PageSize
	if startOffset < 0 {
		startOffset = 0
	}
	endOffset := (startOffset + l.PageSize) - 1
	return getMembersByRange(l.Settings, l.Name, l.PageSize, startOffset, endOffset)
}

func (l *Leaderboard) GetMemberByRank(position int) User {
	conn := getConnection(l.Settings)

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

/* End Public functions */
