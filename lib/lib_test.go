package lib_test

import (
	"net/http"

	"github.com/spf13/viper"
	"github.com/topfreegames/podium/lib"
	httpmock "gopkg.in/jarcoal/httpmock.v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lib", func() {

	var p *lib.Podium
	var config *viper.Viper

	BeforeSuite(func() {
		config = viper.New()
		httpmock.Activate()
	})

	BeforeEach(func() {
		//default configs for each test
		config.Set("podium.url", "http://podium")
		config.Set("podium.user", "user")
		config.Set("podium.pass", "pass")
		config.Set("leaderboards.globalLeaderboard", "ecg:leaderboard:global")
		config.Set("leaderboards.localeLeaderboard", "ecg:leaderboard:%{locale}")
		p = lib.NewPodium(config)
		httpmock.Reset()
	})

	Describe("NewPodium", func() {
		It("Should start a new instance of Podium Lib", func() {
			p = lib.NewPodium(config)
			Expect(p).NotTo(BeNil())
		})
		It("Should use podium.url from config files as the podium URL", func() {
			config.Set("podium.url", "tempURL")
			p = lib.NewPodium(config)
			Expect(p.URL).To(Equal("tempURL"))
		})
	})

	Describe("GetBaseLeaderboards", func() {
		It("Should retrieve BaseLeaderboard from config files", func() {
			config.Set("leaderboards.globalLeaderboard", "MyTestLeaderboard")
			p = lib.NewPodium(config)
			Expect(p.GetBaseLeaderboards()).To(Equal("MyTestLeaderboard"))
		})
	})

	Describe("GetLocalizedLeaderboard", func() {
		It("Should retrieve LocalizedLeaderboard from config files", func() {
			config.Set("leaderboards.localeLeaderboard", "MyTestLeaderboard")
			p = lib.NewPodium(config)
			Expect(p.GetLocalizedLeaderboard("")).To(Equal("MyTestLeaderboard"))
		})
		It("Should change %{locale} in default localeLeaderboard value", func() {
			config.Set("leaderboards.localeLeaderboard", "MyTestLeaderboard:%{locale}")
			p = lib.NewPodium(config)
			Expect(p.GetLocalizedLeaderboard("Brazil")).To(Equal("MyTestLeaderboard:Brazil"))
		})
	})

	Describe("GetTop", func() {
		It("Should call podium API to get the top players", func() {
			leaderboard := p.GetBaseLeaderboards()

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/top/1?pageSize=1"
			httpmock.RegisterResponder("GET", url,
				httpmock.NewStringResponder(200, `{ "success": true, "members": [ { "publicID": "1", "score": 2, "rank": 1 } ] }`))

			status, members, err := p.GetTop(leaderboard, 1, 1)

			Expect(status).To(Equal(200))
			Expect(members).NotTo(BeNil())
			Expect(members.Members[0].PublicID).To(Equal("1"))
			Expect(members.Members[0].Score).To(Equal(2))
			Expect(len(members.Members)).To(Equal(1))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetTopPercent", func() {
		It("Should call API to get the top x% players", func() {
			leaderboard := p.GetBaseLeaderboards()

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/top-percent/1"
			httpmock.RegisterResponder("GET", url,
				httpmock.NewStringResponder(200, `{ "success": true, "members": [ { "publicID": "1", "score": 2, "rank": 1 } ] }`))

			status, members, err := p.GetTopPercent(leaderboard, 1)

			Expect(status).To(Equal(200))
			Expect(members).NotTo(BeNil())
			Expect(members.Members[0].PublicID).To(Equal("1"))
			Expect(members.Members[0].Score).To(Equal(2))
			Expect(len(members.Members)).To(Equal(1))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("UpdateScore", func() {
		It("Should call API to update score of a member", func() {
			leaderboard := p.GetBaseLeaderboards()

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/members/1/score"
			httpmock.RegisterResponder("PUT", url,
				httpmock.NewStringResponder(200, `{ "success": true, "member": { "publicID": "1", "score": 2, "rank": 1 } }`))

			status, members, err := p.UpdateScore(leaderboard, "1", 10)

			Expect(status).To(Equal(200))
			Expect(members).NotTo(BeNil())
			Expect(members.Member.PublicID).To(Equal("1"))
			Expect(members.Member.Score).To(Equal(2))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("IncrementScore", func() {
		It("Should call API to increment score of a member", func() {
			leaderboard := p.GetBaseLeaderboards()

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/members/1/score"
			httpmock.RegisterResponder("PATCH", url,
				httpmock.NewStringResponder(200, `{ "success": true, "member": { "publicID": "123", "score": 12, "rank": 1 } }`))

			status, members, err := p.IncrementScore(leaderboard, "1", 10)

			Expect(status).To(Equal(200))
			Expect(members).NotTo(BeNil())
			Expect(members.Member.PublicID).To(Equal("123"))
			Expect(members.Member.Score).To(Equal(12))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("UpdateScores", func() {
		It("Should call API to update scores of a member in different leaderboards", func() {
			leaderboard1 := p.GetBaseLeaderboards()
			leaderboard2 := p.GetLocalizedLeaderboard("Brazil")

			//mock url that should be called
			url := "http://podium/m/1/scores"
			httpmock.RegisterResponder("PUT", url,
				httpmock.NewStringResponder(200, `{ "success": true, "scores": [ { "leaderboardID": "brazil", "publicID": "1", "score": 1, "rank": 3, "previousRank": 1 } ] }`))

			status, scores, err := p.UpdateScores([]string{leaderboard1, leaderboard2}, "1", 10)

			Expect(status).To(Equal(200))
			Expect(scores).NotTo(BeNil())
			Expect(scores.Scores[0]).NotTo(BeNil())
			Expect(scores.Scores[0].LeaderboardID).To(Equal("brazil"))
			Expect(scores.Scores[0].PublicID).To(Equal("1"))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("RemoveMemberFromLeaderboard", func() {
		It("Should call API to remove player from Leaderboard", func() {
			leaderboard := p.GetBaseLeaderboards()

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/members/1"
			httpmock.RegisterResponder("DELETE", url,
				httpmock.NewStringResponder(200, `{ "success": true }`))

			status, res, err := p.RemoveMemberFromLeaderboard(leaderboard, "1")

			Expect(status).To(Equal(200))
			Expect(res).NotTo(BeNil())
			Expect(res.Success).To(BeTrue())
			Expect(res.Reason).To(BeEmpty())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetPlayer", func() {
		It("Should call API to retrieve player information", func() {
			leaderboard := p.GetBaseLeaderboards()

			//mock url that should be called
			url := "http://podium/l/" + leaderboard + "/members/1"
			httpmock.RegisterResponder("GET", url,
				httpmock.NewStringResponder(200, `{ "success": true,  "publicID": "1", "score": 2, "rank": 1 }`))

			status, member, err := p.GetPlayer(leaderboard, "1")

			Expect(status).To(Equal(200))
			Expect(member).NotTo(BeNil())
			Expect(member.PublicID).To(Equal("1"))
			Expect(member.Score).To(Equal(2))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("DeleteLeaderboard", func() {
		It("Should call API to delete a leaderboard from podium", func() {
			leaderboard := p.GetBaseLeaderboards()

			//mock url that should be called
			url := "http://podium/l/" + leaderboard
			httpmock.RegisterResponder("DELETE", url,
				httpmock.NewStringResponder(200, `{ "success": true }`))

			status, res, err := p.DeleteLeaderboard(leaderboard)

			Expect(status).To(Equal(200))
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

			status, body, err := p.Healthcheck()

			Expect(status).To(Equal(http.StatusOK))
			Expect(body).To(Equal("WORKING"))
			Expect(err).ToNot(HaveOccurred())
		})
		It("Should not respond if server is down", func() {
			//set podium url to be wrong
			config.Set("podium.url", "http://localhostel")
			p = lib.NewPodium(config)

			status, body, err := p.Healthcheck()

			Expect(err).To(HaveOccurred())
			Expect(status).NotTo(Equal(200))
			Expect(body).NotTo(Equal("WORKING"))
		})
	})

	AfterSuite(func() {
		defer httpmock.DeactivateAndReset()
	})
})
