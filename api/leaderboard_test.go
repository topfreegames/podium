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
	"github.com/satori/go.uuid"
	"github.com/topfreegames/podium/api"
	"github.com/topfreegames/podium/leaderboard"
	"github.com/topfreegames/podium/testing"
)

var _ = Describe("Leaderboard Handler", func() {
	var a *api.App
	var l *leaderboard.Leaderboard
	var lg *testing.MockLogger

	BeforeSuite(func() {
		a = api.GetDefaultTestApp()
	})

	BeforeEach(func() {
		lg = testing.NewMockLogger()
		l = leaderboard.NewLeaderboard(a.RedisClient, "testkey", 0, lg)

		conn := a.RedisClient.Client
		conn.Del("testkey")
		conn.Del("testkey1")
		conn.Del("testkey2")
		conn.Del("testkey3")
		conn.Del("testkey4")
		conn.Del("testkey5")
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

		Measure("it should set correct user score", func(b Benchmarker) {
			runtime := b.Time("runtime", func() {
				payload := map[string]interface{}{
					"score": 100,
				}
				res := api.PutJSON(a, "/l/testkey/users/userpublicid/score", payload)
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Set score shouldn't take too long.")
		}, 200)
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
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for user"))
		})

		It("Should not fail in deleting user score from redis if score does not exist", func() {
			res := api.Delete(a, "/l/testkey/users/userpublicid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			_, err := l.GetMember("userpublicid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for user"))
		})

		Measure("it should remove user score", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)
			userID := uuid.NewV4().String()
			_, err := lead.SetUserScore(userID, 100)
			Expect(err).NotTo(HaveOccurred())

			runtime := b.Time("runtime", func() {
				res := api.Delete(a, fmt.Sprintf("/l/%s/users/%s", lead.PublicID, userID))
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Remove member shouldn't take too long.")
		}, 200)
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

		Measure("it should get user score", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)
			userID := uuid.NewV4().String()
			_, err := lead.SetUserScore(userID, 500)
			Expect(err).NotTo(HaveOccurred())

			runtime := b.Time("runtime", func() {
				url := fmt.Sprintf("/l/%s/users/%s", lead.PublicID, userID)
				res := api.Get(a, url)
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.02), "Remove member shouldn't take too long.")
		}, 200)
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

		Measure("it should get user rank", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)
			userID := uuid.NewV4().String()

			_, err := lead.SetUserScore(userID, 500)
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < 10; i++ {
				_, err := lead.SetUserScore(fmt.Sprintf("user-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			runtime := b.Time("runtime", func() {
				url := fmt.Sprintf("/l/%s/users/%s/rank", lead.PublicID, userID)
				res := api.Get(a, url)
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.02), "Remove member shouldn't take too long.")
		}, 200)
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

		It("Should get one page of top users from redis if leaderboard exists but less than pageSize neighbours exist", func() {
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
				Expect(int(user["rank"].(float64))).To(Equal(i + 1))
				Expect(user["publicID"]).To(Equal(fmt.Sprintf("user_%d", i+1)))
				Expect(int(user["score"].(float64))).To(Equal(15 - i))

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

		It("Should get user score and limit neighbours from redis if user score exists and custom limit", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/users/user_50/around", map[string]interface{}{
				"pageSize": 10,
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

		It("Should get user score and limit neighbours from redis if user score exists and repeated scores", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 100)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/users/user_50/around", map[string]interface{}{
				"pageSize": 10,
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(10))
			start := int(users[0].(map[string]interface{})["rank"].(float64))
			for i, userObj := range users {
				user := userObj.(map[string]interface{})
				pos := start + i
				Expect(int(user["rank"].(float64))).To(Equal(pos))
				Expect(int(user["score"].(float64))).To(Equal(100))

				dbUser, err := l.GetMember(user["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbUser.Rank).To(Equal(int(user["rank"].(float64))))
				Expect(dbUser.Score).To(Equal(int(user["score"].(float64))))
				Expect(dbUser.PublicID).To(Equal(user["publicID"]))
			}
		})

		It("Should fail with 404 if score for player does not exist", func() {
			res := api.Get(a, "/l/testkey/users/userpublicid/around")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("User not found."))
		})

		It("Should fail with 400 if bad pageSize provided", func() {
			res := api.Get(a, "/l/testkey/users/user_50/around", map[string]interface{}{
				"pageSize": "notint",
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Invalid pageSize provided: strconv.ParseInt: parsing \"notint\": invalid syntax"))
		})

		It("Should fail with 400 if pageSize provided if bigger than maxPageSizeAllowed", func() {
			pageSize := a.Config.GetInt("api.maxReturnedMembers") + 1
			res := api.Get(a, "/l/testkey/users/user_50/around", map[string]interface{}{
				"pageSize": pageSize,
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Max pageSize allowed: %d. pageSize requested: %d", pageSize-1, pageSize)))
		})

		It("Should get one page of top users from redis if leaderboard exists and user in ranking bottom", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 100-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/users/user_2/around")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(20))
			for i, userObj := range users {
				user := userObj.(map[string]interface{})
				Expect(int(user["rank"].(float64))).To(Equal(i + 1))
				Expect(user["publicID"]).To(Equal(fmt.Sprintf("user_%d", i+1)))
				Expect(int(user["score"].(float64))).To(Equal(100 - i - 1))

				dbUser, err := l.GetMember(user["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbUser.Rank).To(Equal(int(user["rank"].(float64))))
				Expect(dbUser.Score).To(Equal(int(user["score"].(float64))))
				Expect(dbUser.PublicID).To(Equal(user["publicID"]))
			}
		})

		It("Should get one page of top users from redis if leaderboard exists and user in ranking top", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 100-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/users/user_99/around")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(20))
			for i, userObj := range users {
				user := userObj.(map[string]interface{})
				Expect(int(user["rank"].(float64))).To(Equal(80 + i + 1))
				Expect(user["publicID"]).To(Equal(fmt.Sprintf("user_%d", 80+i+1)))
				Expect(int(user["score"].(float64))).To(Equal(100 - 80 - i - 1))

				dbUser, err := l.GetMember(user["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbUser.Rank).To(Equal(int(user["rank"].(float64))))
				Expect(dbUser.Score).To(Equal(int(user["score"].(float64))))
				Expect(dbUser.PublicID).To(Equal(user["publicID"]))
			}
		})

		Measure("it should get around user", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)
			userID := uuid.NewV4().String()

			_, err := lead.SetUserScore(userID, 500)
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < 10; i++ {
				_, err := lead.SetUserScore(fmt.Sprintf("user-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			runtime := b.Time("runtime", func() {
				res := api.Get(a, fmt.Sprintf("/l/%s/users/%s/around", lead.PublicID, userID))
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Getting around user shouldn't take too long.")
		}, 200)
	})

	Describe("Get Total Members Handler", func() {
		It("Should get the number of members in a leaderboard it exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/users-count")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(int(result["count"].(float64))).To(Equal(100))
		})

		It("Should not fail if leaderboard does not exist", func() {
			res := api.Get(a, "/l/testkey/users-count")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(int(result["count"].(float64))).To(Equal(0))
		})

		Measure("it should get total members", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)

			for i := 0; i < 10; i++ {
				_, err := lead.SetUserScore(fmt.Sprintf("user-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			runtime := b.Time("runtime", func() {
				res := api.Get(a, fmt.Sprintf("/l/%s/users-count", lead.PublicID))
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Getting total members shouldn't take too long.")
		}, 200)
	})

	Describe("Get Total Pages Handler", func() {
		It("Should get the number of pages in an existing leaderboard with default pageSize", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/pages")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(int(result["count"].(float64))).To(Equal(5))
		})

		It("Should get the number of pages in an existing leaderboard with custom pageSize", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/pages", map[string]interface{}{
				"pageSize": 10,
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(int(result["count"].(float64))).To(Equal(10))
		})

		It("Should not fail if leaderboard does not exist", func() {
			res := api.Get(a, "/l/testkey/pages")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(int(result["count"].(float64))).To(Equal(0))
		})

		It("Should fail if leaderboard bad pageSize provided", func() {
			res := api.Get(a, "/l/testkey/pages", map[string]interface{}{
				"pageSize": "notint",
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Invalid pageSize provided: strconv.ParseInt: parsing \"notint\": invalid syntax"))
		})

		Measure("it should get total pages", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)

			for i := 0; i < 10; i++ {
				_, err := lead.SetUserScore(fmt.Sprintf("user-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			runtime := b.Time("runtime", func() {
				res := api.Get(a, fmt.Sprintf("/l/%s/pages", lead.PublicID))
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Getting total pages shouldn't take too long.")
		}, 200)
	})

	Describe("Get Top Users Handler", func() {
		It("Should get one page of top users from redis if leaderboard exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/top/1")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(20))
			for i, userObj := range users {
				user := userObj.(map[string]interface{})
				Expect(int(user["rank"].(float64))).To(Equal(i + 1))
				Expect(user["publicID"]).To(Equal(fmt.Sprintf("user_%d", i+1)))
				Expect(int(user["score"].(float64))).To(Equal(100 - i))

				dbUser, err := l.GetMember(user["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbUser.Rank).To(Equal(int(user["rank"].(float64))))
				Expect(dbUser.Score).To(Equal(int(user["score"].(float64))))
				Expect(dbUser.PublicID).To(Equal(user["publicID"]))
			}
		})

		It("Should get one page of top users from redis if leaderboard exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/top/2")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(20))
			start := 20
			for i, userObj := range users {
				pos := start + i
				user := userObj.(map[string]interface{})
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

		It("Should get top users from redis if leaderboard exists with custom pageSize", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/top/1", map[string]interface{}{
				"pageSize": 10,
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(10))
			for i, userObj := range users {
				user := userObj.(map[string]interface{})
				Expect(int(user["rank"].(float64))).To(Equal(i + 1))
				Expect(user["publicID"]).To(Equal(fmt.Sprintf("user_%d", i+1)))
				Expect(int(user["score"].(float64))).To(Equal(100 - i))

				dbUser, err := l.GetMember(user["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbUser.Rank).To(Equal(int(user["rank"].(float64))))
				Expect(dbUser.Score).To(Equal(int(user["score"].(float64))))
				Expect(dbUser.PublicID).To(Equal(user["publicID"]))
			}
		})

		It("Should get empty list if page does not exist", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/top/100000")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(0))
		})

		It("Should get only one page of top users from redis if leaderboard exists and repeated scores", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore("user_"+strconv.Itoa(i), 100)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/top/1", map[string]interface{}{
				"pageSize": 10,
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(10))
			for i, userObj := range users {
				user := userObj.(map[string]interface{})
				Expect(int(user["rank"].(float64))).To(Equal(i + 1))
				Expect(int(user["score"].(float64))).To(Equal(100))

				dbUser, err := l.GetMember(user["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbUser.Rank).To(Equal(int(user["rank"].(float64))))
				Expect(dbUser.Score).To(Equal(int(user["score"].(float64))))
				Expect(dbUser.PublicID).To(Equal(user["publicID"]))
			}
		})

		It("Should fail with 400 if bad pageNumber provided", func() {
			res := api.Get(a, "/l/testkey/top/notint")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Invalid pageNumber provided: strconv.ParseInt: parsing \"notint\": invalid syntax"))
		})

		It("Should fail with 400 if bad pageSize provided", func() {
			res := api.Get(a, "/l/testkey/top/1", map[string]interface{}{
				"pageSize": "notint",
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Invalid pageSize provided: strconv.ParseInt: parsing \"notint\": invalid syntax"))
		})

		It("Should fail with 400 if pageSize provided if bigger than maxPageSizeAllowed", func() {
			pageSize := a.Config.GetInt("api.maxReturnedMembers") + 1
			res := api.Get(a, "/l/testkey/top/1", map[string]interface{}{
				"pageSize": pageSize,
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Max pageSize allowed: %d. pageSize requested: %d", pageSize-1, pageSize)))
		})

		Measure("it should get top users", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)

			for i := 0; i < 100; i++ {
				_, err := lead.SetUserScore(fmt.Sprintf("user-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			runtime := b.Time("runtime", func() {
				res := api.Get(a, fmt.Sprintf("/l/%s/top/10", lead.PublicID))
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Getting top users shouldn't take too long.")
		}, 200)
	})

	Describe("Get Top Percentage Handler", func() {
		It("Should get top users from redis if leaderboard exists", func() {
			leaderboardID := uuid.NewV4().String()
			l = leaderboard.NewLeaderboard(a.RedisClient, leaderboardID, 10, lg)

			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore(fmt.Sprintf("user_%d", i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, fmt.Sprintf("/l/%s/top-percent/10", leaderboardID))
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))

			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(10))

			for i, userObj := range users {
				user := userObj.(map[string]interface{})
				Expect(int(user["rank"].(float64))).To(Equal(i + 1))
				Expect(user["publicID"]).To(Equal(fmt.Sprintf("user_%d", i+1)))
				Expect(int(user["score"].(float64))).To(Equal(100 - i))
			}
		})

		It("Should get top users from redis if leaderboard exists and repeated scores", func() {
			leaderboardID := uuid.NewV4().String()
			l = leaderboard.NewLeaderboard(a.RedisClient, leaderboardID, 10, lg)

			for i := 1; i <= 100; i++ {
				_, err := l.SetUserScore(fmt.Sprintf("user_%d", i), 100)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, fmt.Sprintf("/l/%s/top-percent/10", leaderboardID))
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))

			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			Expect(result["success"]).To(BeTrue())
			users := result["users"].([]interface{})
			Expect(len(users)).To(Equal(10))

			for i, userObj := range users {
				user := userObj.(map[string]interface{})
				Expect(int(user["rank"].(float64))).To(Equal(i + 1))
				Expect(int(user["score"].(float64))).To(Equal(100))
			}
		})

		Measure("it should get top percentage of users", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)

			for i := 0; i < 100; i++ {
				_, err := lead.SetUserScore(fmt.Sprintf("user-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			runtime := b.Time("runtime", func() {
				res := api.Get(a, fmt.Sprintf("/l/%s/top-percent/10", lead.PublicID))
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Getting top percentage users shouldn't take too long.")
		}, 200)
	})

	Describe("Upsert User Score For Several Leaderboards", func() {
		It("Should set correct user score in redis and respond with the correct values", func() {
			payload := map[string]interface{}{
				"score":        100,
				"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
			}
			res := api.PutJSON(a, "/u/userpublicid/scores", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			scores := result["scores"].([]interface{})
			Expect(len(scores)).To(Equal(5))
			for i, scoreObj := range scores {
				score := scoreObj.(map[string]interface{})
				Expect(score["publicID"]).To(Equal("userpublicid"))
				Expect(int(score["score"].(float64))).To(Equal(payload["score"]))
				Expect(int(score["rank"].(float64))).To(Equal(1))
				Expect(score["leaderboardID"]).To(Equal(payload["leaderboards"].([]string)[i]))

				ll := leaderboard.NewLeaderboard(a.RedisClient, score["leaderboardID"].(string), 0, lg)
				user, err := ll.GetMember("userpublicid")
				Expect(err).NotTo(HaveOccurred())
				Expect(user.Rank).To(Equal(1))
				Expect(user.Score).To(Equal(100))
				Expect(user.PublicID).To(Equal("userpublicid"))
			}
		})

		It("Should fail if missing score", func() {
			payload := map[string]interface{}{
				"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
			}
			res := api.PutJSON(a, "/u/userpublicid/scores", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("score is required"))
		})

		It("Should fail if missing leaderboards", func() {
			payload := map[string]interface{}{
				"score": 100,
			}
			res := api.PutJSON(a, "/u/userpublicid/scores", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("leaderboards is required"))
		})

		It("Should fail if invalid payload", func() {
			res := api.PutBody(a, "/u/userpublicid/scores", "invalid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring("While trying to read JSON"))
		})

		Measure("it should set correct user score for all leaderboards", func(b Benchmarker) {
			runtime := b.Time("runtime", func() {
				payload := map[string]interface{}{
					"score":        100,
					"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
				}
				res := api.PutJSON(a, "/u/userpublicid/scores", payload)
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Set score shouldn't take too long.")
		}, 200)
	})
})
