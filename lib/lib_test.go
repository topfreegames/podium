package lib_test

import (
	"github.com/spf13/viper"
	"github.com/topfreegames/podium/lib"
	httpmock "gopkg.in/jarcoal/httpmock.v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lib", func() {

	var p lib.PodiumInterface
	var config *viper.Viper
	var globalLeaderboard string
	var localeLeaderboard string

	BeforeSuite(func() {
		config = viper.New()
		httpmock.Activate()
	})

	BeforeEach(func() {
		//default configs for each test
		config.Set("podium.url", "http://podium")
		config.Set("podium.user", "user")
		config.Set("podium.pass", "pass")
		globalLeaderboard = "game:leaderboard:global"
		localeLeaderboard = "game:leaderboard:us"
		p = lib.NewPodium(config)
		httpmock.Reset()
	})

	Describe("NewPodium", func() {
		It("Should start a new instance of Podium Lib", func() {
			p = lib.NewPodium(config)
			Expect(p).NotTo(BeNil())
		})
	})

	Describe("GetTop", func() {
		It("Should call podium API to get the top players", func() {
			leaderboard := globalLeaderboard

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/top/1?pageSize=1"
			httpmock.RegisterResponder("GET", url,
				httpmock.NewStringResponder(200, `{ "success": true, "members": [ { "publicID": "1", "score": 2, "rank": 1 } ] }`))

			members, err := p.GetTop(nil, leaderboard, 1, 1)

			Expect(members).NotTo(BeNil())
			Expect(members.Members[0].PublicID).To(Equal("1"))
			Expect(members.Members[0].Score).To(Equal(2))
			Expect(len(members.Members)).To(Equal(1))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetTopPercent", func() {
		It("Should call API to get the top x% players", func() {
			leaderboard := globalLeaderboard

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/top-percent/1"
			httpmock.RegisterResponder("GET", url,
				httpmock.NewStringResponder(200, `{ "success": true, "members": [ { "publicID": "1", "score": 2, "rank": 1 } ] }`))

			members, err := p.GetTopPercent(nil, leaderboard, 1)

			Expect(members).NotTo(BeNil())
			Expect(members.Members[0].PublicID).To(Equal("1"))
			Expect(members.Members[0].Score).To(Equal(2))
			Expect(len(members.Members)).To(Equal(1))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("UpdateScore", func() {
		It("Should call API to update score of a member", func() {
			leaderboard := globalLeaderboard

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/members/1/score"
			httpmock.RegisterResponder("PUT", url,
				httpmock.NewStringResponder(200, `{ "success": true, "member": { "publicID": "1", "score": 2, "rank": 1 } }`))

			members, err := p.UpdateScore(nil, leaderboard, "1", 10, 0)

			Expect(members).NotTo(BeNil())
			Expect(members.Member.PublicID).To(Equal("1"))
			Expect(members.Member.Score).To(Equal(2))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should call API to update score of a member with TTL", func() {
			leaderboard := globalLeaderboard

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/members/1/score?scoreTTL=10"
			httpmock.RegisterResponder("PUT", url,
				httpmock.NewStringResponder(200, `{ "success": true, "member": { "publicID": "1", "score": 2, "rank": 1 } }`))

			members, err := p.UpdateScore(nil, leaderboard, "1", 10, 10)

			Expect(members).NotTo(BeNil())
			Expect(members.Member.PublicID).To(Equal("1"))
			Expect(members.Member.Score).To(Equal(2))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("IncrementScore", func() {
		It("Should call API to increment score of a member", func() {
			leaderboard := globalLeaderboard

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/members/1/score"
			httpmock.RegisterResponder("PATCH", url,
				httpmock.NewStringResponder(200, `{ "success": true, "member": { "publicID": "123", "score": 12, "rank": 1 } }`))

			members, err := p.IncrementScore(nil, leaderboard, "1", 10, 0)

			Expect(members).NotTo(BeNil())
			Expect(members.Member.PublicID).To(Equal("123"))
			Expect(members.Member.Score).To(Equal(12))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should call API to increment score of a member with TTL", func() {
			leaderboard := globalLeaderboard

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/members/1/score?scoreTTL=10"
			httpmock.RegisterResponder("PATCH", url,
				httpmock.NewStringResponder(200, `{ "success": true, "member": { "publicID": "123", "score": 12, "rank": 1 } }`))

			members, err := p.IncrementScore(nil, leaderboard, "1", 10, 10)

			Expect(members).NotTo(BeNil())
			Expect(members.Member.PublicID).To(Equal("123"))
			Expect(members.Member.Score).To(Equal(12))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("UpdateScores", func() {
		It("Should call API to update scores of a member in different leaderboards", func() {
			leaderboard1 := globalLeaderboard
			leaderboard2 := localeLeaderboard

			//mock url that should be called
			url := "http://podium/m/1/scores"
			httpmock.RegisterResponder("PUT", url,
				httpmock.NewStringResponder(200, `{ "success": true, "scores": [ { "leaderboardID": "brazil", "publicID": "1", "score": 1, "rank": 3, "previousRank": 1 } ] }`))

			scores, err := p.UpdateScores(nil, []string{leaderboard1, leaderboard2}, "1", 10, 0)

			Expect(scores).NotTo(BeNil())
			Expect(scores.Scores[0]).NotTo(BeNil())
			Expect(scores.Scores[0].LeaderboardID).To(Equal("brazil"))
			Expect(scores.Scores[0].PublicID).To(Equal("1"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should call API to update scores of a member in different leaderboards with TTL", func() {
			leaderboard1 := globalLeaderboard
			leaderboard2 := localeLeaderboard

			//mock url that should be called
			url := "http://podium/m/1/scores?scoreTTL=10"
			httpmock.RegisterResponder("PUT", url,
				httpmock.NewStringResponder(200, `{ "success": true, "scores": [ { "leaderboardID": "brazil", "publicID": "1", "score": 1, "rank": 3, "previousRank": 1 } ] }`))

			scores, err := p.UpdateScores(nil, []string{leaderboard1, leaderboard2}, "1", 10, 10)

			Expect(scores).NotTo(BeNil())
			Expect(scores.Scores[0]).NotTo(BeNil())
			Expect(scores.Scores[0].LeaderboardID).To(Equal("brazil"))
			Expect(scores.Scores[0].PublicID).To(Equal("1"))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("RemoveMemberFromLeaderboard", func() {
		It("Should call API to remove player from Leaderboard", func() {
			leaderboard := globalLeaderboard

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/members/1"
			httpmock.RegisterResponder("DELETE", url,
				httpmock.NewStringResponder(200, `{ "success": true }`))

			res, err := p.RemoveMemberFromLeaderboard(nil, leaderboard, "1")

			Expect(res).NotTo(BeNil())
			Expect(res.Success).To(BeTrue())
			Expect(res.Reason).To(BeEmpty())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetMember", func() {
		It("Should call API to retrieve member information", func() {
			leaderboard := globalLeaderboard

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/members/1"
			httpmock.RegisterResponder("GET", url,
				httpmock.NewStringResponder(200, `{ "success": true,  "publicID": "1", "score": 2, "rank": 1 }`))

			member, err := p.GetMember(nil, leaderboard, "1")

			Expect(member).NotTo(BeNil())
			Expect(member.PublicID).To(Equal("1"))
			Expect(member.Score).To(Equal(2))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetMembers", func() {
		It("Should call API to retrieve members information", func() {
			leaderboard := globalLeaderboard

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/members?ids=1,2,3"
			httpmock.RegisterResponder("GET", url,
				httpmock.NewStringResponder(200, `{
					"success": true,
					"members": [
						{ "publicID": "1", "score": 5, "rank": 1 },
						{ "publicID": "3", "score": 4, "rank": 2 }
					],
					"notFound": ["2"]
				}`))

			members, err := p.GetMembers(nil, leaderboard, []string{"1", "2", "3"})

			Expect(err).To(BeNil())
			Expect(members).NotTo(BeNil())
			Expect(members.Members[0].PublicID).To(Equal("1"))
			Expect(members.Members[0].Score).To(Equal(5))
			Expect(members.Members[1].PublicID).To(Equal("3"))
			Expect(members.Members[1].Score).To(Equal(4))
			Expect(members.NotFound[0]).To(Equal("2"))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("DeleteLeaderboard", func() {
		It("Should call API to delete a leaderboard from podium", func() {
			leaderboard := globalLeaderboard

			//mock url that should be called
			url := "http://podium/l/" + leaderboard
			httpmock.RegisterResponder("DELETE", url,
				httpmock.NewStringResponder(200, `{ "success": true }`))

			res, err := p.DeleteLeaderboard(nil, leaderboard)

			Expect(res).NotTo(BeNil())
			Expect(res.Success).To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Healthcheck", func() {
		It("Should respond with default WORKING string", func() {
			//mock url that should be called
			url := "http://podium/healthcheck"
			httpmock.RegisterResponder("GET", url,
				httpmock.NewStringResponder(200, `WORKING`))

			body, err := p.Healthcheck(nil)

			Expect(body).To(Equal("WORKING"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should not respond if server is down", func() {
			//set podium url to be wrong
			config.Set("podium.url", "http://localhostel")
			p = lib.NewPodium(config)

			body, err := p.Healthcheck(nil)

			Expect(err).To(HaveOccurred())
			Expect(body).NotTo(Equal("WORKING"))
		})
	})

	AfterSuite(func() {
		defer httpmock.DeactivateAndReset()
	})
})
