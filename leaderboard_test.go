package leaderboard

import (
	"launchpad.net/gocheck"
	"testing"
	"strconv"
)

func Test(t *testing.T) {
	gocheck.TestingT(t)
}

type S struct{}

var _ = gocheck.Suite(&S{})

func (s *S) TearDownSuite(c *gocheck.C) {
	conn := getConnection()
	conn.Do("DEL", "highscore")
	conn.Do("DEL", "bestTime")
	conn.Do("DEL", "bestWeek")
	conn.Do("DEL", "friendScore")

}

func (s *S) TestRankMember(c *gocheck.C) {
	highScore := NewLeaderboard("highscore", 10)
	dayvson, err := highScore.RankMember("dayvson",  481516)
	c.Assert(err, gocheck.IsNil)
	arthur, err := highScore.RankMember("arthur",  1000)
	c.Assert(err, gocheck.IsNil)
	c.Assert(dayvson.rank, gocheck.Equals, 1)
	c.Assert(arthur.rank, gocheck.Equals, 2)
}

func (s *S) TestTotalMembers(c *gocheck.C) {
	bestTime := NewLeaderboard("bestTime", 10)
	for i := 0; i<10; i++ {
		bestTime.RankMember("member_" + strconv.Itoa(i),  1234 * i)	
	}
	c.Assert(bestTime.TotalMembers(), gocheck.Equals, 10)
}

func (s *S) TestRemoveMember(c *gocheck.C) {
	bestTime := NewLeaderboard("bestWeek", 10)
	for i := 0; i<10; i++ {
		bestTime.RankMember("member_" + strconv.Itoa(i),  1234 * i)	
	}
	c.Assert(bestTime.TotalMembers(), gocheck.Equals, 10)
	bestTime.RemoveMember("member_5")
	c.Assert(bestTime.TotalMembers(), gocheck.Equals, 9)
}

func (s *S) TestTotalPages(c *gocheck.C) {
	bestTime := NewLeaderboard("All", 25)
	for i := 0; i<101; i++ {
		bestTime.RankMember("member_" + strconv.Itoa(i),  1234 * i)	
	}
	c.Assert(bestTime.TotalPages(), gocheck.Equals, 5)
}

func (s *S) TestGetUser( c *gocheck.C) {
	friendScore := NewLeaderboard("friendScore", 10)
	dayvson, _ := friendScore.RankMember("dayvson",  12345)
	felipe, _ := friendScore.RankMember("felipe",  12344)
	c.Assert(dayvson.rank, gocheck.Equals, 1)
	c.Assert(felipe.rank, gocheck.Equals, 2)
	friendScore.RankMember("felipe", 12346)
	felipe, _ = friendScore.GetMember("felipe")
	dayvson, _ = friendScore.GetMember("dayvson")
	c.Assert(felipe.rank, gocheck.Equals, 1)
	c.Assert(dayvson.rank, gocheck.Equals, 2)

}