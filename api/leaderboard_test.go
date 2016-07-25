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

	Describe("Upsert Member Score", func() {
		It("Should set correct member score in redis and respond with the correct values", func() {
			payload := map[string]interface{}{
				"score": 100,
			}
			res := api.PutJSON(a, "/l/testkey/members/memberpublicid/score", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int(result["score"].(float64))).To(Equal(payload["score"]))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			member, err := l.GetMember("memberpublicid")
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(100))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should fail if missing parameters", func() {
			payload := map[string]interface{}{
				"notscore": 100,
			}
			res := api.PutJSON(a, "/l/testkey/members/memberpublicid/score", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("score is required"))
		})

		It("Should fail if invalid payload", func() {
			res := api.PutBody(a, "/l/testkey/members/memberpublicid/score", "invalid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring("While trying to read JSON"))
		})

		Measure("it should set correct member score", func(b Benchmarker) {
			runtime := b.Time("runtime", func() {
				payload := map[string]interface{}{
					"score": 100,
				}
				res := api.PutJSON(a, "/l/testkey/members/memberpublicid/score", payload)
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Set score shouldn't take too long.")
		}, 200)
	})

	Describe("Remove Member Score", func() {
		It("Should delete member score from redis if score exists", func() {
			_, err := l.SetMemberScore("memberpublicid", 100)
			Expect(err).NotTo(HaveOccurred())

			res := api.Delete(a, "/l/testkey/members/memberpublicid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			_, err = l.GetMember("memberpublicid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member"))
		})

		It("Should not fail in deleting member score from redis if score does not exist", func() {
			res := api.Delete(a, "/l/testkey/members/memberpublicid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			_, err := l.GetMember("memberpublicid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member"))
		})

		Measure("it should remove member score", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)
			memberID := uuid.NewV4().String()
			_, err := lead.SetMemberScore(memberID, 100)
			Expect(err).NotTo(HaveOccurred())

			runtime := b.Time("runtime", func() {
				res := api.Delete(a, fmt.Sprintf("/l/%s/members/%s", lead.PublicID, memberID))
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Remove member shouldn't take too long.")
		}, 200)
	})

	Describe("Get Member", func() {
		It("Should get member score from redis if score exists", func() {
			_, err := l.SetMemberScore("memberpublicid", 100)
			Expect(err).NotTo(HaveOccurred())

			res := api.Get(a, "/l/testkey/members/memberpublicid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int(result["score"].(float64))).To(Equal(100))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			member, err := l.GetMember("memberpublicid")
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(100))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should fail with 404 if score does not exist", func() {
			res := api.Get(a, "/l/testkey/members/memberpublicid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Member not found."))
		})

		Measure("it should get member score", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)
			memberID := uuid.NewV4().String()
			_, err := lead.SetMemberScore(memberID, 500)
			Expect(err).NotTo(HaveOccurred())

			runtime := b.Time("runtime", func() {
				url := fmt.Sprintf("/l/%s/members/%s", lead.PublicID, memberID)
				res := api.Get(a, url)
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.02), "Remove member shouldn't take too long.")
		}, 200)
	})

	Describe("Get Member Rank", func() {
		It("Should get member score from redis if score exists", func() {
			_, err := l.SetMemberScore("memberpublicid", 100)
			Expect(err).NotTo(HaveOccurred())

			res := api.Get(a, "/l/testkey/members/memberpublicid/rank")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			member, err := l.GetMember("memberpublicid")
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(100))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should fail with 404 if score does not exist", func() {
			res := api.Get(a, "/l/testkey/members/memberpublicid/rank")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Member not found."))
		})

		Measure("it should get member rank", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)
			memberID := uuid.NewV4().String()

			_, err := lead.SetMemberScore(memberID, 500)
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < 10; i++ {
				_, err := lead.SetMemberScore(fmt.Sprintf("member-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			runtime := b.Time("runtime", func() {
				url := fmt.Sprintf("/l/%s/members/%s/rank", lead.PublicID, memberID)
				res := api.Get(a, url)
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.02), "Remove member shouldn't take too long.")
		}, 200)
	})

	Describe("Get Around Member Handler", func() {
		It("Should get member score and neighbours from redis if member score exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/members/member_50/around")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			start := 50 - 20/2
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				pos := start + i
				Expect(int(member["rank"].(float64))).To(Equal(pos + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", pos+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - pos))

				dbMember, err := l.GetMember(member["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get one page of top members from redis if leaderboard exists but less than pageSize neighbours exist", func() {
			for i := 1; i <= 15; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 16-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/members/member_10/around")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(15))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", i+1)))
				Expect(int(member["score"].(float64))).To(Equal(15 - i))

				dbMember, err := l.GetMember(member["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get member score and default limit neighbours from redis if member score and less than limit neighbours exist", func() {
			for i := 1; i <= 15; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 16-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/members/member_10/around")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(15))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				pos := i
				Expect(int(member["rank"].(float64))).To(Equal(pos + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", pos+1)))
				Expect(int(member["score"].(float64))).To(Equal(15 - pos))

				dbMember, err := l.GetMember(member["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get member score and limit neighbours from redis if member score exists and custom limit", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/members/member_50/around", map[string]interface{}{
				"pageSize": 10,
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(10))
			start := 50 - 10/2
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				pos := start + i
				Expect(int(member["rank"].(float64))).To(Equal(pos + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", pos+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - pos))

				dbMember, err := l.GetMember(member["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get member score and limit neighbours from redis if member score exists and repeated scores", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 100)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/members/member_50/around", map[string]interface{}{
				"pageSize": 10,
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(10))
			start := int(members[0].(map[string]interface{})["rank"].(float64))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				pos := start + i
				Expect(int(member["rank"].(float64))).To(Equal(pos))
				Expect(int(member["score"].(float64))).To(Equal(100))

				dbMember, err := l.GetMember(member["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should fail with 404 if score for member does not exist", func() {
			res := api.Get(a, "/l/testkey/members/memberpublicid/around")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusNotFound))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Member not found."))
		})

		It("Should fail with 400 if bad pageSize provided", func() {
			res := api.Get(a, "/l/testkey/members/member_50/around", map[string]interface{}{
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
			res := api.Get(a, "/l/testkey/members/member_50/around", map[string]interface{}{
				"pageSize": pageSize,
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Max pageSize allowed: %d. pageSize requested: %d", pageSize-1, pageSize)))
		})

		It("Should get one page of top members from redis if leaderboard exists and member in ranking bottom", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 100-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/members/member_2/around")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", i+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - i - 1))

				dbMember, err := l.GetMember(member["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get one page of top members from redis if leaderboard exists and member in ranking top", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 100-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/members/member_99/around")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(80 + i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", 80+i+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - 80 - i - 1))

				dbMember, err := l.GetMember(member["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		Measure("it should get around member", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)
			memberID := uuid.NewV4().String()

			_, err := lead.SetMemberScore(memberID, 500)
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < 10; i++ {
				_, err := lead.SetMemberScore(fmt.Sprintf("member-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			runtime := b.Time("runtime", func() {
				res := api.Get(a, fmt.Sprintf("/l/%s/members/%s/around", lead.PublicID, memberID))
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Getting around member shouldn't take too long.")
		}, 200)
	})

	Describe("Get Total Members Handler", func() {
		It("Should get the number of members in a leaderboard it exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/members-count")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(int(result["count"].(float64))).To(Equal(100))
		})

		It("Should not fail if leaderboard does not exist", func() {
			res := api.Get(a, "/l/testkey/members-count")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(int(result["count"].(float64))).To(Equal(0))
		})

		Measure("it should get total members", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)

			for i := 0; i < 10; i++ {
				_, err := lead.SetMemberScore(fmt.Sprintf("member-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			runtime := b.Time("runtime", func() {
				res := api.Get(a, fmt.Sprintf("/l/%s/members-count", lead.PublicID))
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Getting total members shouldn't take too long.")
		}, 200)
	})

	Describe("Get Total Pages Handler", func() {
		It("Should get the number of pages in an existing leaderboard with default pageSize", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 101-i)
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
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 101-i)
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
				_, err := lead.SetMemberScore(fmt.Sprintf("member-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			runtime := b.Time("runtime", func() {
				res := api.Get(a, fmt.Sprintf("/l/%s/pages", lead.PublicID))
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Getting total pages shouldn't take too long.")
		}, 200)
	})

	Describe("Get Top Members Handler", func() {
		It("Should get one page of top members from redis if leaderboard exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/top/1")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", i+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - i))

				dbMember, err := l.GetMember(member["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get one page of top members from redis if leaderboard exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/top/2")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			start := 20
			for i, memberObj := range members {
				pos := start + i
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(pos + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", pos+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - pos))

				dbMember, err := l.GetMember(member["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get top members from redis if leaderboard exists with custom pageSize", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/top/1", map[string]interface{}{
				"pageSize": 10,
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(10))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", i+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - i))

				dbMember, err := l.GetMember(member["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get empty list if page does not exist", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/top/100000")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(0))
		})

		It("Should get only one page of top members from redis if leaderboard exists and repeated scores", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 100)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/top/1", map[string]interface{}{
				"pageSize": 10,
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(10))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(int(member["score"].(float64))).To(Equal(100))

				dbMember, err := l.GetMember(member["publicID"].(string))
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
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

		Measure("it should get top members", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)

			for i := 0; i < 100; i++ {
				_, err := lead.SetMemberScore(fmt.Sprintf("member-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			runtime := b.Time("runtime", func() {
				res := api.Get(a, fmt.Sprintf("/l/%s/top/10", lead.PublicID))
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Getting top members shouldn't take too long.")
		}, 200)
	})

	Describe("Get Top Percentage Handler", func() {
		It("Should get top members from redis if leaderboard exists", func() {
			leaderboardID := uuid.NewV4().String()
			l = leaderboard.NewLeaderboard(a.RedisClient, leaderboardID, 10, lg)

			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore(fmt.Sprintf("member_%d", i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, fmt.Sprintf("/l/%s/top-percent/10", leaderboardID))
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))

			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(10))

			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", i+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - i))
			}
		})

		It("Should get top members from redis if leaderboard exists and repeated scores", func() {
			leaderboardID := uuid.NewV4().String()
			l = leaderboard.NewLeaderboard(a.RedisClient, leaderboardID, 10, lg)

			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore(fmt.Sprintf("member_%d", i), 100)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, fmt.Sprintf("/l/%s/top-percent/10", leaderboardID))
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))

			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(10))

			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(int(member["score"].(float64))).To(Equal(100))
			}
		})

		Measure("it should get top percentage of members", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)

			for i := 0; i < 100; i++ {
				_, err := lead.SetMemberScore(fmt.Sprintf("member-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			runtime := b.Time("runtime", func() {
				res := api.Get(a, fmt.Sprintf("/l/%s/top-percent/10", lead.PublicID))
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Getting top percentage members shouldn't take too long.")
		}, 200)
	})

	Describe("Upsert Member Score For Several Leaderboards", func() {
		It("Should set correct member score in redis and respond with the correct values", func() {
			payload := map[string]interface{}{
				"score":        100,
				"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
			}
			res := api.PutJSON(a, "/m/memberpublicid/scores", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			scores := result["scores"].([]interface{})
			Expect(len(scores)).To(Equal(5))
			for i, scoreObj := range scores {
				score := scoreObj.(map[string]interface{})
				Expect(score["publicID"]).To(Equal("memberpublicid"))
				Expect(int(score["score"].(float64))).To(Equal(payload["score"]))
				Expect(int(score["rank"].(float64))).To(Equal(1))
				Expect(score["leaderboardID"]).To(Equal(payload["leaderboards"].([]string)[i]))

				ll := leaderboard.NewLeaderboard(a.RedisClient, score["leaderboardID"].(string), 0, lg)
				member, err := ll.GetMember("memberpublicid")
				Expect(err).NotTo(HaveOccurred())
				Expect(member.Rank).To(Equal(1))
				Expect(member.Score).To(Equal(100))
				Expect(member.PublicID).To(Equal("memberpublicid"))
			}
		})

		It("Should fail if missing score", func() {
			payload := map[string]interface{}{
				"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
			}
			res := api.PutJSON(a, "/m/memberpublicid/scores", payload)
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
			res := api.PutJSON(a, "/m/memberpublicid/scores", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("leaderboards is required"))
		})

		It("Should fail if invalid payload", func() {
			res := api.PutBody(a, "/m/memberpublicid/scores", "invalid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring("While trying to read JSON"))
		})

		Measure("it should set correct member score for all leaderboards", func(b Benchmarker) {
			runtime := b.Time("runtime", func() {
				payload := map[string]interface{}{
					"score":        100,
					"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
				}
				res := api.PutJSON(a, "/m/memberpublicid/scores", payload)
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.03), "Set score shouldn't take too long.")
		}, 200)
	})
})
