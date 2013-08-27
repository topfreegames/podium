package leaderboard

import (
	"fmt"
	"math"
	"github.com/garyburd/redigo/redis"
)

/* Structs model */
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

/* End Structs model */

var pool *redis.Pool

/* Private functions */

func getConnection() redis.Conn {
	if pool == nil {
		srv := "localhost:6379"
		pool = redis.NewPool(func() (redis.Conn, error) {
			return redis.Dial("tcp", srv)
		}, 10)
	}
	return pool.Get()
}

func getMembersByRange(leaderboard string, pageSize int, startOffset int, endOffset int) []User {
	conn := getConnection()
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
    	i+= 1
    }
    return users
}
/* End Private functions */


/* Public functions */

func NewLeaderboard(name string, pageSize int) Leaderboard {
	l := Leaderboard{name: name, pageSize: pageSize}
	return l
}

func (self *Leaderboard) RankMember(username string, score int) (User, error) {
	conn := getConnection()
	defer conn.Close()
	_, err := conn.Do("ZADD", self.name, score, username)
	if err != nil {
		fmt.Printf("error on store in redis in rankMember Leaderboard:%s - Username:%s - Score:%d", self.name, username, score)
	}
	rank, err := redis.Int(conn.Do("ZREVRANK", self.name, username))
	if err != nil {
		fmt.Printf("error on get user rank Leaderboard:%s - Username:%s", self.name, username)
		rank = -1
	}
	nUser := User{name: username, score: score, rank: rank + 1 }
	return nUser, err
}

func (self *Leaderboard) TotalMembers() int {
	conn := getConnection()
	defer conn.Close()
	total, err := redis.Int(conn.Do("ZCARD", self.name))
	if err != nil {
		fmt.Printf("error on get leaderboard total members")
		return 0
	}
	return total
}

func (self *Leaderboard) RemoveMember(username string) (User, error) {
	conn := getConnection()
	defer conn.Close()
	nUser, err := self.GetMember(username)
	_, err = conn.Do("ZREM", self.name, username)
	if err != nil {
		fmt.Printf("error on remove user from leaderboard")
	}
	return nUser, err
}

func (self *Leaderboard) TotalPages() int {
	conn := getConnection()
	defer conn.Close()
	pages := 0
	total, err := redis.Int(conn.Do("ZCOUNT", self.name, "-inf", "+inf"))
	if err == nil {
		pages = int(math.Ceil(float64(total) / float64(self.pageSize)))
	}
	return pages
}

func (self *Leaderboard) GetMember(username string) (User, error) {
	conn := getConnection()
	defer conn.Close()
	rank, err := redis.Int(conn.Do("ZREVRANK", self.name, username))
	if err != nil {
		rank = 0
	}
	score, err := redis.Int(conn.Do("ZSCORE", self.name, username))
	if err != nil {
		score = 0
	}
	nUser := User{name: username, score: score, rank: rank+1}
	return nUser, err
}

func (self *Leaderboard) GetAroundMe(username string) []User {
	currentUser, _ := self.GetMember(username)
	startOffset := currentUser.rank - (self.pageSize / 2)
    if startOffset < 0 {
    	startOffset = 0
    }
    endOffset := (startOffset + self.pageSize) - 1
    return getMembersByRange(self.name, self.pageSize, startOffset, endOffset)
}

func (self *Leaderboard) GetRank(username string) int {
	conn := getConnection()
	defer conn.Close()
	rank, _ := redis.Int(conn.Do("ZREVRANK", self.name, username))
	return rank + 1
}

func (self *Leaderboard) GetLeaders(page int) []User {
	if page < 1 {
		page = 1
	}
	if page > self.TotalPages() {
		page = self.TotalPages()
	}
    redisIndex := page - 1
    startOffset := redisIndex * self.pageSize
    if startOffset < 0 {
    	startOffset = 0
    }
	endOffset := (startOffset + self.pageSize) - 1
	
	return getMembersByRange(self.name, self.pageSize, startOffset, endOffset)
}

func (self *Leaderboard) GetMemberByRank(position int) User {
	conn := getConnection()
	defer conn.Close()
	if position <= self.TotalMembers() {
		currentPage := int(math.Ceil(float64(position) / float64(self.pageSize)))
		offset := (position - 1) % self.pageSize
		leaders := self.GetLeaders(currentPage)
		if leaders[offset].rank == position {
			return leaders[offset]
		}
	}
	return User{}
}
/* End Public functions */