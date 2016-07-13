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
	"fmt"
	"net/http"
	"strconv"

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
		conn.Do("DEL", "testkey")
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

	Describe("Get User Rank", func() {
		It("Should get user score from redis if score exists", func() {
			_, err := l.SetUserScore("userpublicid", 100)
			Expect(err).NotTo(HaveOccurred())

			res := api.Get(a, "/l/testkey/users/userpublicid/rank")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("userpublicid"))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			user, err := l.GetMember("userpublicid")
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Rank).To(Equal(1))
			Expect(user.Score).To(Equal(100))
			Expect(user.PublicID).To(Equal("userpublicid"))
		})

		It("Should fail with 404 if score does not exist", func() {
			res := api.Get(a, "/l/testkey/users/userpublicid/rank")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("User not found."))
		})
	})

	Describe("Get Around User Handler", func() {
		It("Should get user score and neighbours from redis if user score exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/users/user_50/around")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(20))
			start := 50 - 20/2
			for i, userObj := range users {
				user := userObj.(map[string]interface{})
				pos := start + i
				Expect(int(user["rank"].(float64))).To(Equal(pos + 1))
				Expect(user["publicID"]).To(Equal(fmt.Sprintf("user_%d", pos+1)))
				Expect(int(user["score"].(float64))).To(Equal(100 - pos))

				dbUser, err := l.GetMember(user["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbUser.Rank).To(Equal(int(user["rank"].(float64))))
				Expect(dbUser.Score).To(Equal(int(user["score"].(float64))))
				Expect(dbUser.PublicID).To(Equal(user["publicID"]))
			}
		})

		It("Should get user score and default limit neighbours from redis if user score and less than limit neighbours exist", func() {
			for i := 1; i <= 15; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 16-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/users/user_10/around")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(15))
			for i, userObj := range users {
				user := userObj.(map[string]interface{})
				pos := i
				Expect(int(user["rank"].(float64))).To(Equal(pos + 1))
				Expect(user["publicID"]).To(Equal(fmt.Sprintf("user_%d", pos+1)))
				Expect(int(user["score"].(float64))).To(Equal(15 - pos))

				dbUser, err := l.GetMember(user["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbUser.Rank).To(Equal(int(user["rank"].(float64))))
				Expect(dbUser.Score).To(Equal(int(user["score"].(float64))))
				Expect(dbUser.PublicID).To(Equal(user["publicID"]))
			}
		})

		It("Should get user score and limit neighbours from redis if user score exists and and custom limit", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/users/user_50/around", map[string]interface{}{
				"limit": 10,
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(10))
			start := 50 - 10/2
			for i, userObj := range users {
				user := userObj.(map[string]interface{})
				pos := start + i
				Expect(int(user["rank"].(float64))).To(Equal(pos + 1))
				Expect(user["publicID"]).To(Equal(fmt.Sprintf("user_%d", pos+1)))
				Expect(int(user["score"].(float64))).To(Equal(100 - pos))

				dbUser, err := l.GetMember(user["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbUser.Rank).To(Equal(int(user["rank"].(float64))))
				Expect(dbUser.Score).To(Equal(int(user["score"].(float64))))
				Expect(dbUser.PublicID).To(Equal(user["publicID"]))
			}
		})

		It("Should fail with 404 if score for player does not exist", func() {
			res := api.Get(a, "/l/testkey/users/userpublicid/rank")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("User not found."))
		})

		It("Should fail with 400 if bad limit provided", func() {
			res := api.Get(a, "/l/testkey/users/user_50/around", map[string]interface{}{
				"limit": "notint",
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Invalid limit provided: strconv.ParseInt: parsing \"notint\": invalid syntax"))
		})
	})
})
