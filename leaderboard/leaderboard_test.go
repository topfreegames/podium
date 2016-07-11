package leaderboard_test

import (
	"strconv"

	. "github.com/topfreegames/go-leaderboard/leaderboard"
	"github.com/topfreegames/go-leaderboard/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Leaderboard", func() {

	testRedisSettings := util.RedisSettings{
		Host:     "localhost:1234",
		Password: "",
	}

	BeforeSuite(func() {
		conn := util.GetConnection(testRedisSettings)
		conn.Do("DEL", "highscore")
		conn.Do("DEL", "bestTime")
		conn.Do("DEL", "bestWeek")
		conn.Do("DEL", "friendScore")
		conn.Do("DEL", "7days")
		conn.Do("DEL", "bestYear")
		conn.Do("DEL", "week")
	})

	It("TestSetUserScore", func() {
		highScore := NewLeaderboard(testRedisSettings, "highscore", 10)
		dayvson, err := highScore.SetUserScore("dayvson", 481516)
		Expect(err).To(BeNil())
		arthur, err := highScore.SetUserScore("arthur", 1000)
		Expect(err).To(BeNil())
		Expect(dayvson.Rank).To(Equal(1))
		Expect(arthur.Rank).To(Equal(2))
	})

	It("TestTotalMembers", func() {
		bestTime := NewLeaderboard(testRedisSettings, "bestTime", 10)
		for i := 0; i < 10; i++ {
			bestTime.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
		}
		Expect(bestTime.TotalMembers()).To(Equal(10))
	})

	It("TestRemoveMember", func() {
		bestTime := NewLeaderboard(testRedisSettings, "bestWeek", 10)
		for i := 0; i < 10; i++ {
			bestTime.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
		}
		Expect(bestTime.TotalMembers()).To(Equal(10))
		bestTime.RemoveMember("member_5")
		Expect(bestTime.TotalMembers()).To(Equal(9))
	})

	It("TestTotalPages", func() {
		bestTime := NewLeaderboard(testRedisSettings, "All", 25)
		for i := 0; i < 101; i++ {
			bestTime.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
		}
		Expect(bestTime.TotalPages()).To(Equal(5))
	})

	It("TestGetUser", func() {
		friendScore := NewLeaderboard(testRedisSettings, "friendScore", 10)
		dayvson, _ := friendScore.SetUserScore("dayvson", 12345)
		felipe, _ := friendScore.SetUserScore("felipe", 12344)
		Expect(dayvson.Rank).To(Equal(1))
		Expect(felipe.Rank).To(Equal(2))
		friendScore.SetUserScore("felipe", 12346)
		felipe, _ = friendScore.GetMember("felipe")
		dayvson, _ = friendScore.GetMember("dayvson")
		Expect(felipe.Rank).To(Equal(1))
		Expect(dayvson.Rank).To(Equal(2))
	})

	It("TestGetAroundMe", func() {
		bestTime := NewLeaderboard(testRedisSettings, "BestAllTime", 25)
		for i := 0; i < 101; i++ {
			bestTime.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
		}
		users := bestTime.GetAroundMe("member_20")
		firstAroundMe := users[0]
		lastAroundMe := users[bestTime.PageSize-1]
		Expect(len(users)).To(Equal(bestTime.PageSize))
		Expect(firstAroundMe.PublicID).To(Equal("member_31"))
		Expect(lastAroundMe.PublicID).To(Equal("member_7"))
	})

	It("TestGetRank", func() {
		sevenDays := NewLeaderboard(testRedisSettings, "7days", 25)
		for i := 0; i < 101; i++ {
			sevenDays.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
		}
		sevenDays.SetUserScore("member_6", 1000)
		Expect(sevenDays.GetRank("member_6")).To(Equal(100))
	})

	It("TestGetLeaders", func() {
		bestYear := NewLeaderboard(testRedisSettings, "bestYear", 25)
		for i := 0; i < 1000; i++ {
			bestYear.SetUserScore("member_"+strconv.Itoa(i+1), 1234*i)
		}
		var users = bestYear.GetLeaders(1)

		firstOnPage := users[0]
		lastOnPage := users[len(users)-1]
		Expect(len(users)).To(Equal(bestYear.PageSize))
		Expect(firstOnPage.PublicID).To(Equal("member_1000"))
		Expect(firstOnPage.Rank).To(Equal(1))
		Expect(lastOnPage.PublicID).To(Equal("member_976"))
		Expect(lastOnPage.Rank).To(Equal(25))
	})

	It("TestGetUserByRank", func() {
		sevenDays := NewLeaderboard(testRedisSettings, "week", 25)
		for i := 0; i < 101; i++ {
			sevenDays.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
		}
		member := sevenDays.GetMemberByRank(10)
		Expect(member.PublicID).To(Equal("member_91"))
		Expect(member.Rank).To(Equal(10))
	})
})
