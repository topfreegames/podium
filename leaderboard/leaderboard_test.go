package leaderboard_test

import (
	"strconv"

	. "github.com/topfreegames/go-leaderboard/leaderboard"
	"github.com/topfreegames/go-leaderboard/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Leaderboard", func() {

	redisSettings := util.RedisSettings{
		Host:     "localhost:6379",
		Password: "",
	}

	BeforeSuite(func() {
		conn := util.GetConnection(redisSettings)
		conn.Do("DEL", "highscore")
		conn.Do("DEL", "bestTime")
		conn.Do("DEL", "bestWeek")
		conn.Do("DEL", "friendScore")
		conn.Do("DEL", "7days")
		conn.Do("DEL", "bestYear")
		conn.Do("DEL", "week")
	})

	It("TestRankMember", func() {
		highScore := NewLeaderboard(redisSettings, "highscore", 10)
		dayvson, err := highScore.RankMember("dayvson", 481516)
		Expect(err).To(BeNil())
		arthur, err := highScore.RankMember("arthur", 1000)
		Expect(err).To(BeNil())
		Expect(dayvson.Rank).To(Equal(1))
		Expect(arthur.Rank).To(Equal(2))
	})

	It("TestTotalMembers", func() {
		bestTime := NewLeaderboard(redisSettings, "bestTime", 10)
		for i := 0; i < 10; i++ {
			bestTime.RankMember("member_"+strconv.Itoa(i), 1234*i)
		}
		Expect(bestTime.TotalMembers()).To(Equal(10))
	})

	It("TestRemoveMember", func() {
		bestTime := NewLeaderboard(redisSettings, "bestWeek", 10)
		for i := 0; i < 10; i++ {
			bestTime.RankMember("member_"+strconv.Itoa(i), 1234*i)
		}
		Expect(bestTime.TotalMembers()).To(Equal(10))
		bestTime.RemoveMember("member_5")
		Expect(bestTime.TotalMembers()).To(Equal(9))
	})

	It("TestTotalPages", func() {
		bestTime := NewLeaderboard(redisSettings, "All", 25)
		for i := 0; i < 101; i++ {
			bestTime.RankMember("member_"+strconv.Itoa(i), 1234*i)
		}
		Expect(bestTime.TotalPages()).To(Equal(5))
	})

	It("TestGetUser", func() {
		friendScore := NewLeaderboard(redisSettings, "friendScore", 10)
		dayvson, _ := friendScore.RankMember("dayvson", 12345)
		felipe, _ := friendScore.RankMember("felipe", 12344)
		Expect(dayvson.Rank).To(Equal(1))
		Expect(felipe.Rank).To(Equal(2))
		friendScore.RankMember("felipe", 12346)
		felipe, _ = friendScore.GetMember("felipe")
		dayvson, _ = friendScore.GetMember("dayvson")
		Expect(felipe.Rank).To(Equal(1))
		Expect(dayvson.Rank).To(Equal(2))
	})

	It("TestGetAroundMe", func() {
		bestTime := NewLeaderboard(redisSettings, "BestAllTime", 25)
		for i := 0; i < 101; i++ {
			bestTime.RankMember("member_"+strconv.Itoa(i), 1234*i)
		}
		users := bestTime.GetAroundMe("member_20")
		firstAroundMe := users[0]
		lastAroundMe := users[bestTime.PageSize-1]
		Expect(len(users)).To(Equal(bestTime.PageSize))
		Expect(firstAroundMe.Name).To(Equal("member_31"))
		Expect(lastAroundMe.Name).To(Equal("member_7"))
	})

	It("TestGetRank", func() {
		sevenDays := NewLeaderboard(redisSettings, "7days", 25)
		for i := 0; i < 101; i++ {
			sevenDays.RankMember("member_"+strconv.Itoa(i), 1234*i)
		}
		sevenDays.RankMember("member_6", 1000)
		Expect(sevenDays.GetRank("member_6")).To(Equal(100))
	})

	It("TestGetLeaders", func() {
		bestYear := NewLeaderboard(redisSettings, "bestYear", 25)
		for i := 0; i < 1000; i++ {
			bestYear.RankMember("member_"+strconv.Itoa(i+1), 1234*i)
		}
		var users = bestYear.GetLeaders(1)

		firstOnPage := users[0]
		lastOnPage := users[len(users)-1]
		Expect(len(users)).To(Equal(bestYear.PageSize))
		Expect(firstOnPage.Name).To(Equal("member_1000"))
		Expect(firstOnPage.Rank).To(Equal(1))
		Expect(lastOnPage.Name).To(Equal("member_976"))
		Expect(lastOnPage.Rank).To(Equal(25))
	})

	It("TestGetUserByRank", func() {
		sevenDays := NewLeaderboard(redisSettings, "week", 25)
		for i := 0; i < 101; i++ {
			sevenDays.RankMember("member_"+strconv.Itoa(i), 1234*i)
		}
		member := sevenDays.GetMemberByRank(10)
		Expect(member.Name).To(Equal("member_91"))
		Expect(member.Rank).To(Equal(10))
	})
})
