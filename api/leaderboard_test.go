// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package api_test

import (
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/api"
	"github.com/topfreegames/podium/leaderboard"
)

var _ = Describe("Leaderboard Handler", func() {
	Describe("Upsert User Score", func() {
		It("Should set correct user score in redis and respond with the correct values", func() {
			a := api.GetDefaultTestApp()
			payload := map[string]interface{}{
				"score": 100,
			}
			res := api.PutJSON(a, "/abc_week23/users/userpublicid/score", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("userpublicid"))
			Expect(int(result["score"].(float64))).To(Equal(payload["score"]))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			l := leaderboard.NewLeaderboard(a.RedisClient, "abc_week23", 0)
			user, err := l.GetMember("userpublicid")
			Expect(err).To(BeNil())
			Expect(user.Rank).To(Equal(1))
			Expect(user.Score).To(Equal(100))
			Expect(user.PublicID).To(Equal("userpublicid"))
		})
	})

	It("Should fail if missing parameters", func() {
		a := api.GetDefaultTestApp()
		payload := map[string]interface{}{
			"notscore": 100,
		}
		res := api.PutJSON(a, "/abc_week23/users/userpublicid/score", payload)
		Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
		var result map[string]interface{}
		json.Unmarshal([]byte(res.Body().Raw()), &result)
		Expect(result["success"]).To(BeFalse())
		Expect(result["reason"]).To(Equal("score is required"))
	})

	It("Should fail if invalid payload", func() {
		a := api.GetDefaultTestApp()
		res := api.PutBody(a, "/abc_week23/users/userpublicid/score", "invalid")
		Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
		var result map[string]interface{}
		json.Unmarshal([]byte(res.Body().Raw()), &result)
		Expect(result["success"]).To(BeFalse())
		Expect(result["reason"]).To(ContainSubstring("While trying to read JSON"))
	})
})
