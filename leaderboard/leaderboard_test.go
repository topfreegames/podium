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

	BeforeEach(func() {
		conn := util.GetConnection(testRedisSettings)
		conn.Do("DEL", "testleaderbord")
	})

	AfterSuite(func() {
		conn := util.GetConnection(testRedisSettings)
		conn.Do("DEL", "testleaderbord")
	})

	It("TestSetUserScore", func() {
		testLeaderboard := NewLeaderboard(testRedisSettings, "testleaderbord", 10)
		dayvson, err := testLeaderboard.SetUserScore("dayvson", 481516)
		Expect(err).To(BeNil())
		arthur, err := testLeaderboard.SetUserScore("arthur", 1000)
		Expect(err).To(BeNil())
		Expect(dayvson.Rank).To(Equal(1))
		Expect(arthur.Rank).To(Equal(2))
	})

	It("TestTotalMembers", func() {
		testLeaderboard := NewLeaderboard(testRedisSettings, "testleaderbord", 10)
		for i := 0; i < 10; i++ {
			testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
		}
		Expect(testLeaderboard.TotalMembers()).To(Equal(10))
	})

	It("TestRemoveMember", func() {
		testLeaderboard := NewLeaderboard(testRedisSettings, "testleaderbord", 10)
		for i := 0; i < 10; i++ {
			testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
		}
		Expect(testLeaderboard.TotalMembers()).To(Equal(10))
		testLeaderboard.RemoveMember("member_5")
		Expect(testLeaderboard.TotalMembers()).To(Equal(9))
	})

	It("TestTotalPages", func() {
		testLeaderboard := NewLeaderboard(testRedisSettings, "testleaderbord", 25)
		for i := 0; i < 101; i++ {
			testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
		}
		Expect(testLeaderboard.TotalPages()).To(Equal(5))
	})

	It("TestGetUser", func() {
		friendScore := NewLeaderboard(testRedisSettings, "testleaderbord", 10)
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
		testLeaderboard := NewLeaderboard(testRedisSettings, "testleaderbord", 25)
		for i := 0; i < 101; i++ {
			testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
		}
		users := testLeaderboard.GetAroundMe("member_20")
		firstAroundMe := users[0]
		lastAroundMe := users[testLeaderboard.PageSize-1]
		Expect(len(users)).To(Equal(testLeaderboard.PageSize))
		Expect(firstAroundMe.PublicID).To(Equal("member_31"))
		Expect(lastAroundMe.PublicID).To(Equal("member_7"))
	})

	It("TestGetRank", func() {
		testLeaderboard := NewLeaderboard(testRedisSettings, "testleaderbord", 25)
		for i := 0; i < 101; i++ {
			testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
		}
		testLeaderboard.SetUserScore("member_6", 1000)
		Expect(testLeaderboard.GetRank("member_6")).To(Equal(100))
	})

	It("TestGetLeaders", func() {
		testLeaderboard := NewLeaderboard(testRedisSettings, "testleaderbord", 25)
		for i := 0; i < 1000; i++ {
			testLeaderboard.SetUserScore("member_"+strconv.Itoa(i+1), 1234*i)
		}
		var users = testLeaderboard.GetLeaders(1)

		firstOnPage := users[0]
		lastOnPage := users[len(users)-1]
		Expect(len(users)).To(Equal(testLeaderboard.PageSize))
		Expect(firstOnPage.PublicID).To(Equal("member_1000"))
		Expect(firstOnPage.Rank).To(Equal(1))
		Expect(lastOnPage.PublicID).To(Equal("member_976"))
		Expect(lastOnPage.Rank).To(Equal(25))
	})

	It("TestGetUserByRank", func() {
		testLeaderboard := NewLeaderboard(testRedisSettings, "testleaderbord", 25)
		for i := 0; i < 101; i++ {
			testLeaderboard.SetUserScore("member_"+strconv.Itoa(i), 1234*i)
		}
		member := testLeaderboard.GetMemberByRank(10)
		Expect(member.PublicID).To(Equal("member_91"))
		Expect(member.Rank).To(Equal(10))
	})
})
