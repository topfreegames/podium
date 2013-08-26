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

func (s *S) TestRankMember(c *gocheck.C) {
	highScore := NewLeaderboard("highscore", 10)
	dayvson, err := highScore.rankMember("dayvson",  481516)
	c.Assert(err, gocheck.IsNil)
	arthur, err := highScore.rankMember("arthur",  1000)
	c.Assert(err, gocheck.IsNil)
	c.Assert(dayvson.rank, gocheck.DeepEquals, 0)
	c.Assert(arthur.rank, gocheck.DeepEquals, 1)
}

func (s *S) TestTotalMembers(c *gocheck.C) {
	bestTime := NewLeaderboard("bestTime", 10)
	for i := 0; i<10; i++ {
		bestTime.rankMember("member_" + strconv.Itoa(i),  1234 * i)	
	}
	c.Assert(bestTime.TotalMembers(), gocheck.DeepEquals, 10)
}
