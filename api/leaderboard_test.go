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
	var a *api.App
	var l *leaderboard.Leaderboard

	BeforeEach(func() {
		a = api.GetDefaultTestApp()
		l = leaderboard.NewLeaderboard(a.RedisClient, "testkey", 0)
		conn := a.RedisClient.GetConnection()
		conn.Do("FLUSHALL")
	})

	Describe("Upsert User Score", func() {
		It("Should set correct user score in redis and respond with the correct values", func() {
			payload := map[string]interface{}{
				"score": 100,
			}
			res := api.PutJSON(a, "/l/testkey/users/userpublicid/score", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("userpublicid"))
			Expect(int(result["score"].(float64))).To(Equal(payload["score"]))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			user, err := l.GetMember("userpublicid")
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Rank).To(Equal(1))
			Expect(user.Score).To(Equal(100))
			Expect(user.PublicID).To(Equal("userpublicid"))
		})

		It("Should fail if missing parameters", func() {
			payload := map[string]interface{}{
				"notscore": 100,
			}
			res := api.PutJSON(a, "/l/testkey/users/userpublicid/score", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("score is required"))
		})

		It("Should fail if invalid payload", func() {
			res := api.PutBody(a, "/l/testkey/users/userpublicid/score", "invalid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring("While trying to read JSON"))
		})
	})

	Describe("Remove User Score", func() {
		It("Should delete user score from redis if score exists", func() {
			_, err := l.SetUserScore("userpublicid", 100)
			Expect(err).NotTo(HaveOccurred())

			res := api.Delete(a, "/l/testkey/users/userpublicid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			_, err = l.GetMember("userpublicid")
			Expect(err.Error()).To(Equal("redigo: nil returned"))
		})

		It("Should not fail in deleting user score from redis if score does not exist", func() {
			res := api.Delete(a, "/l/testkey/users/userpublicid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			_, err := l.GetMember("userpublicid")
			Expect(err.Error()).To(Equal("redigo: nil returned"))
		})
	})

	Describe("Get User", func() {
		It("Should get user score from redis if score exists", func() {
			_, err := l.SetUserScore("userpublicid", 100)
			Expect(err).NotTo(HaveOccurred())

			res := api.Get(a, "/l/testkey/users/userpublicid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("userpublicid"))
			Expect(int(result["score"].(float64))).To(Equal(100))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			user, err := l.GetMember("userpublicid")
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Rank).To(Equal(1))
			Expect(user.Score).To(Equal(100))
			Expect(user.PublicID).To(Equal("userpublicid"))
		})

		It("Should fail with 404 if score does not exist", func() {
			res := api.Get(a, "/l/testkey/users/userpublicid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("User not found."))
		})
	})
})
