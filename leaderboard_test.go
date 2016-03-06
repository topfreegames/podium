package leaderboard

import (
	"strconv"
	"testing"

	"launchpad.net/gocheck"
)

func Test(t *testing.T) {
	gocheck.TestingT(t)
}

type S struct{}

var _ = gocheck.Suite(&S{})
var redisSettings = RedisSettings{
	Host:     "localhost:6379",
	Password: "",
}

func (s *S) TearDownSuite(c *gocheck.C) {

	conn := getConnection(redisSettings)
	conn.Do("DEL", "highscore")
	conn.Do("DEL", "bestTime")
	conn.Do("DEL", "bestWeek")
	conn.Do("DEL", "friendScore")
	conn.Do("DEL", "7days")
	conn.Do("DEL", "bestYear")
	conn.Do("DEL", "week")
}

func (s *S) TestRankMember(c *gocheck.C) {
	highScore := NewLeaderboard(redisSettings, "highscore", 10)
	dayvson, err := highScore.RankMember("dayvson", 481516)
	c.Assert(err, gocheck.IsNil)
	arthur, err := highScore.RankMember("arthur", 1000)
	c.Assert(err, gocheck.IsNil)
	c.Assert(dayvson.Rank, gocheck.Equals, 1)
	c.Assert(arthur.Rank, gocheck.Equals, 2)
}

func (s *S) TestTotalMembers(c *gocheck.C) {
	bestTime := NewLeaderboard(redisSettings, "bestTime", 10)
	for i := 0; i < 10; i++ {
		bestTime.RankMember("member_"+strconv.Itoa(i), 1234*i)
	}
	c.Assert(bestTime.TotalMembers(), gocheck.Equals, 10)
}

func (s *S) TestRemoveMember(c *gocheck.C) {
	bestTime := NewLeaderboard(redisSettings, "bestWeek", 10)
	for i := 0; i < 10; i++ {
		bestTime.RankMember("member_"+strconv.Itoa(i), 1234*i)
	}
	c.Assert(bestTime.TotalMembers(), gocheck.Equals, 10)
	bestTime.RemoveMember("member_5")
	c.Assert(bestTime.TotalMembers(), gocheck.Equals, 9)
}

func (s *S) TestTotalPages(c *gocheck.C) {
	bestTime := NewLeaderboard(redisSettings, "All", 25)
	for i := 0; i < 101; i++ {
		bestTime.RankMember("member_"+strconv.Itoa(i), 1234*i)
	}
	c.Assert(bestTime.TotalPages(), gocheck.Equals, 5)
}

func (s *S) TestGetUser(c *gocheck.C) {
	friendScore := NewLeaderboard(redisSettings, "friendScore", 10)
	dayvson, _ := friendScore.RankMember("dayvson", 12345)
	felipe, _ := friendScore.RankMember("felipe", 12344)
	c.Assert(dayvson.Rank, gocheck.Equals, 1)
	c.Assert(felipe.Rank, gocheck.Equals, 2)
	friendScore.RankMember("felipe", 12346)
	felipe, _ = friendScore.GetMember("felipe")
	dayvson, _ = friendScore.GetMember("dayvson")
	c.Assert(felipe.Rank, gocheck.Equals, 1)
	c.Assert(dayvson.Rank, gocheck.Equals, 2)
}

func (s *S) TestGetAroundMe(c *gocheck.C) {
	bestTime := NewLeaderboard(redisSettings, "BestAllTime", 25)
	for i := 0; i < 101; i++ {
		bestTime.RankMember("member_"+strconv.Itoa(i), 1234*i)
	}
	users := bestTime.GetAroundMe("member_20")
	firstAroundMe := users[0]
	lastAroundMe := users[bestTime.PageSize-1]
	c.Assert(len(users), gocheck.Equals, bestTime.PageSize)
	c.Assert(firstAroundMe.Name, gocheck.Equals, "member_31")
	c.Assert(lastAroundMe.Name, gocheck.Equals, "member_7")
}

func (s *S) TestGetRank(c *gocheck.C) {
	sevenDays := NewLeaderboard(redisSettings, "7days", 25)
	for i := 0; i < 101; i++ {
		sevenDays.RankMember("member_"+strconv.Itoa(i), 1234*i)
	}
	sevenDays.RankMember("member_6", 1000)
	c.Assert(sevenDays.GetRank("member_6"), gocheck.Equals, 100)
}

func (s *S) TestGetLeaders(c *gocheck.C) {
	bestYear := NewLeaderboard(redisSettings, "bestYear", 25)
	for i := 0; i < 1000; i++ {
		bestYear.RankMember("member_"+strconv.Itoa(i+1), 1234*i)
	}
	var users = bestYear.GetLeaders(1)

	firstOnPage := users[0]
	lastOnPage := users[len(users)-1]
	c.Assert(len(users), gocheck.Equals, bestYear.PageSize)
	c.Assert(firstOnPage.Name, gocheck.Equals, "member_1000")
	c.Assert(firstOnPage.Rank, gocheck.Equals, 1)
	c.Assert(lastOnPage.Name, gocheck.Equals, "member_976")
	c.Assert(lastOnPage.Rank, gocheck.Equals, 25)
}

func (s *S) TestGetUserByRank(c *gocheck.C) {
	sevenDays := NewLeaderboard(redisSettings, "week", 25)
	for i := 0; i < 101; i++ {
		sevenDays.RankMember("member_"+strconv.Itoa(i), 1234*i)
	}
	member := sevenDays.GetMemberByRank(10)
	c.Assert(member.Name, gocheck.Equals, "member_91")
	c.Assert(member.Rank, gocheck.Equals, 10)
}
