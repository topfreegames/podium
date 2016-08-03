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
	"strings"

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

			member, err := l.GetMember("memberpublicid", "desc")
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

		It("Should fail if error updating score", func() {
			payload := map[string]interface{}{
				"score": 100,
			}
			app := api.GetDefaultTestApp()
			app.RedisClient = api.GetFaultyRedis(a.Logger)

			res := api.PutJSON(app, "/l/testkey/members/memberpublicid/score", payload)
			Expect(res.Raw().StatusCode).To(Equal(500))
			Expect(res.Body().Raw()).To(ContainSubstring("connection refused"))
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

			res := api.DeleteWithQuery(a, "/l/testkey/members", "ids", "memberpublicid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK), res.Body().Raw())
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			_, err = l.GetMember("memberpublicid", "desc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member"))
		})

		It("Should delete many member score from redis if they exists", func() {
			_, err := l.SetMemberScore("memberpublicid", 100)
			_, err = l.SetMemberScore("memberpublicid2", 100)
			Expect(err).NotTo(HaveOccurred())

			res := api.DeleteWithQuery(a, "/l/testkey/members", "ids", "memberpublicid,memberpublicid2")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK), res.Body().Raw())
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			_, err = l.GetMember("memberpublicid", "desc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member"))
			_, err = l.GetMember("memberpublicid2", "desc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member"))
		})

		It("Should fail if error removing score", func() {
			_, err := l.SetMemberScore("memberpublicid", 100)
			Expect(err).NotTo(HaveOccurred())

			app := api.GetDefaultTestApp()
			app.RedisClient = api.GetFaultyRedis(a.Logger)

			res := api.DeleteWithQuery(app, "/l/testkey/members", "ids", "memberpublicid")
			Expect(res.Raw().StatusCode).To(Equal(500))
			Expect(res.Body().Raw()).To(ContainSubstring("connection refused"))
		})

		It("Should not fail in deleting member score from redis if score does not exist", func() {
			res := api.DeleteWithQuery(a, "/l/testkey/members", "ids", "memberpublicid")
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())

			_, err := l.GetMember("memberpublicid", "desc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member"))
		})

		Measure("it should remove member score", func(b Benchmarker) {
			lead := leaderboard.NewLeaderboard(a.RedisClient, uuid.NewV4().String(), 0, lg)
			memberID := uuid.NewV4().String()
			_, err := lead.SetMemberScore(memberID, 100)
			Expect(err).NotTo(HaveOccurred())

			runtime := b.Time("runtime", func() {
				res := api.DeleteWithQuery(a, fmt.Sprintf("/l/%s/members", lead.PublicID), "ids", memberID)
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

			member, err := l.GetMember("memberpublicid", "desc")
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

		It("Should fail if error in Redis", func() {
			app := api.GetDefaultTestApp()
			app.RedisClient = api.GetFaultyRedis(a.Logger)

			res := api.Get(app, "/l/testkey/members/member_99")
			Expect(res.Raw().StatusCode).To(Equal(500))
			Expect(res.Body().Raw()).To(ContainSubstring("connection refused"))
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

			member, err := l.GetMember("memberpublicid", "desc")
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(100))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should get member score from redis if score exists and order is asc", func() {
			_, err := l.SetMemberScore("memberpublicid", 100)
			Expect(err).NotTo(HaveOccurred())

			res := api.Get(a, "/l/testkey/members/memberpublicid/rank", map[string]interface{}{
				"order": "asc",
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			member, err := l.GetMember("memberpublicid", "desc")
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

		It("Should fail if error in Redis", func() {
			app := api.GetDefaultTestApp()
			app.RedisClient = api.GetFaultyRedis(a.Logger)

			res := api.Get(app, "/l/testkey/members/member_99/rank")
			Expect(res.Raw().StatusCode).To(Equal(500))
			Expect(res.Body().Raw()).To(ContainSubstring("connection refused"))
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

				dbMember, err := l.GetMember(member["publicID"].(string), "desc")
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get member score and neighbours from redis in reverse order if member score exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/members/member_50/around", map[string]interface{}{
				"order": "asc",
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			start := 50 + 20/2
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				pos := start - i
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", pos-1)))
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

				dbMember, err := l.GetMember(member["publicID"].(string), "desc")
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

				dbMember, err := l.GetMember(member["publicID"].(string), "desc")
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

				dbMember, err := l.GetMember(member["publicID"].(string), "desc")
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

				dbMember, err := l.GetMember(member["publicID"].(string), "desc")
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

				dbMember, err := l.GetMember(member["publicID"].(string), "desc")
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

				dbMember, err := l.GetMember(member["publicID"].(string), "desc")
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should fail if error in Redis", func() {
			app := api.GetDefaultTestApp()
			app.RedisClient = api.GetFaultyRedis(a.Logger)

			res := api.Get(app, "/l/testkey/members/member_99/around")
			Expect(res.Raw().StatusCode).To(Equal(500))
			Expect(res.Body().Raw()).To(ContainSubstring("connection refused"))
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

		It("Should fail if error in Redis", func() {
			app := api.GetDefaultTestApp()
			app.RedisClient = api.GetFaultyRedis(a.Logger)

			res := api.Get(app, "/l/testkey/members-count")
			Expect(res.Raw().StatusCode).To(Equal(500))
			Expect(res.Body().Raw()).To(ContainSubstring("connection refused"))
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

				dbMember, err := l.GetMember(member["publicID"].(string), "desc")
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get one page of top members in reverse order from redis if leaderboard exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore("member_"+strconv.Itoa(i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, "/l/testkey/top/1", map[string]interface{}{
				"order": "asc",
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", 100-i)))
				Expect(int(member["score"].(float64))).To(Equal(i + 1))
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

				dbMember, err := l.GetMember(member["publicID"].(string), "desc")
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

				dbMember, err := l.GetMember(member["publicID"].(string), "desc")
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

				dbMember, err := l.GetMember(member["publicID"].(string), "desc")
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

		It("Should fail if error getting top members from Redis", func() {
			app := api.GetDefaultTestApp()
			app.RedisClient = api.GetFaultyRedis(a.Logger)

			res := api.Get(app, "/l/testkey/top/1")
			Expect(res.Raw().StatusCode).To(Equal(500))
			Expect(res.Body().Raw()).To(ContainSubstring("connection refused"))
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

		It("Should fail if invalid percentage", func() {
			leaderboardID := uuid.NewV4().String()
			l = leaderboard.NewLeaderboard(a.RedisClient, leaderboardID, 10, lg)

			res := api.Get(a, fmt.Sprintf("/l/%s/top-percent/l", leaderboardID))
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			Expect(res.Body().Raw()).To(ContainSubstring("Invalid percentage provided"))
		})

		It("Should fail if percentage greater than 100", func() {
			leaderboardID := uuid.NewV4().String()
			res := api.Get(a, fmt.Sprintf("/l/%s/top-percent/120", leaderboardID))
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			Expect(res.Body().Raw()).To(ContainSubstring("Percentage must be a valid integer between 1 and 100."))
		})

		It("Should fail if percentage lesser than 1", func() {
			leaderboardID := uuid.NewV4().String()
			res := api.Get(a, fmt.Sprintf("/l/%s/top-percent/0", leaderboardID))
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))
			Expect(res.Body().Raw()).To(ContainSubstring("Percentage must be a valid integer between 1 and 100."))
		})

		It("Should fail if error in Redis", func() {
			app := api.GetDefaultTestApp()
			app.RedisClient = api.GetFaultyRedis(a.Logger)

			res := api.Get(app, "/l/testkey/top-percent/10")
			Expect(res.Raw().StatusCode).To(Equal(500))
			Expect(res.Body().Raw()).To(ContainSubstring("connection refused"))
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

	Describe("Get member score in many leaderboads", func() {
		It("Should get member score in many leaderboards", func() {
			payload := map[string]interface{}{
				"score":        100,
				"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
			}
			res := api.PutJSON(a, "/m/memberpublicid/scores", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			res = api.Get(a, "/m/memberpublicid/scores", map[string]interface{}{
				"leaderboardIds": "testkey1,testkey2,testkey3,testkey4,testkey5",
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			scores := result["scores"].([]interface{})
			for i, scoreObj := range scores {
				score := scoreObj.(map[string]interface{})
				Expect(int(score["score"].(float64))).To(Equal(payload["score"]))
				Expect(int(score["rank"].(float64))).To(Equal(1))
				Expect(score["leaderboardID"]).To(Equal(payload["leaderboards"].([]string)[i]))
			}
		})

		It("Should fail if pass a leaderboard that does not exist", func() {
			payload := map[string]interface{}{
				"score":        100,
				"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
			}
			res := api.PutJSON(a, "/m/memberpublicid/scores", payload)
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			res = api.Get(a, "/m/memberpublicid/scores", map[string]interface{}{
				"leaderboardIds": "testkey1,testkey2,testkey3,testkey4,testkey6",
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusNotFound))
		})

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
				member, err := ll.GetMember("memberpublicid", "desc")
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

		It("Should fail if error in Redis when upserting many leaderboards", func() {
			app := api.GetDefaultTestApp()
			app.RedisClient = api.GetFaultyRedis(a.Logger)

			payload := map[string]interface{}{
				"score":        100,
				"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
			}
			res := api.PutJSON(app, "/m/memberpublicid/scores", payload)
			Expect(res.Raw().StatusCode).To(Equal(500))
			Expect(res.Body().Raw()).To(ContainSubstring("connection refused"))
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

	Describe("Remove Leaderboard", func() {
		It("should remove a leaderboard", func() {
			leaderboardID := uuid.NewV4().String()
			lead := leaderboard.NewLeaderboard(a.RedisClient, leaderboardID, 0, lg)

			for i := 0; i < 10; i++ {
				_, err := lead.SetMemberScore(fmt.Sprintf("member-%d", i), 500)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Delete(a, fmt.Sprintf("/l/%s", leaderboardID))
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
		})

		It("should remove a leaderboard that does not exist", func() {
			res := api.Delete(a, fmt.Sprintf("/l/%s", uuid.NewV4().String()))
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)
			Expect(result["success"]).To(BeTrue())
		})

		It("Should fail if error in Redis", func() {
			app := api.GetDefaultTestApp()
			app.RedisClient = api.GetFaultyRedis(a.Logger)

			res := api.Delete(app, fmt.Sprintf("/l/%s", uuid.NewV4().String()))
			Expect(res.Raw().StatusCode).To(Equal(500))
			Expect(res.Body().Raw()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("Get Members Handler", func() {
		It("should get several members from leaderboard", func() {
			leaderboardID := uuid.NewV4().String()
			l = leaderboard.NewLeaderboard(a.RedisClient, leaderboardID, 10, lg)

			for i := 1; i <= 100; i++ {
				_, err := l.SetMemberScore(fmt.Sprintf("member_%d", i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, fmt.Sprintf("/l/%s/members/", l.PublicID), map[string]interface{}{
				"ids": "member_10,member_20,member_30",
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))

			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			Expect(result["success"]).To(BeTrue())
			Expect(result["notFound"]).To(BeEmpty())
			members := result["members"].([]interface{})
			Expect(members).To(HaveLen(3))

			for i := 0; i < 3; i++ {
				By(fmt.Sprintf("Member %d", i))
				member := members[i].(map[string]interface{})
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", (i+1)*10)))
				Expect(member["rank"]).To(BeEquivalentTo((i + 1) * 10))
				Expect(member["score"]).To(BeEquivalentTo(101 - (i+1)*10))
				Expect(member["position"]).To(BeEquivalentTo(i))
			}
		})

		It("should return empty list if invalid leaderboard", func() {
			res := api.Get(a, "/l/invalid-leaderboard/members/", map[string]interface{}{
				"ids": "member_10,member_20,member_30",
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))

			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			Expect(result["success"]).To(BeTrue())
			Expect(result["members"]).To(HaveLen(0))
			Expect(result["notFound"]).To(HaveLen(3))

			members := result["notFound"].([]interface{})
			Expect(members).To(HaveLen(3))
			Expect(members[0]).To(Equal("member_10"))
			Expect(members[1]).To(Equal("member_20"))
			Expect(members[2]).To(Equal("member_30"))
		})

		It("should return not found members", func() {
			leaderboardID := uuid.NewV4().String()
			l = leaderboard.NewLeaderboard(a.RedisClient, leaderboardID, 10, lg)

			for i := 1; i <= 10; i++ {
				_, err := l.SetMemberScore(fmt.Sprintf("member_%d", i), 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			res := api.Get(a, fmt.Sprintf("/l/%s/members/", l.PublicID), map[string]interface{}{
				"ids": "member_1,invalid_member",
			})
			Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))

			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(members).To(HaveLen(1))

			member := members[0].(map[string]interface{})
			Expect(member["publicID"]).To(Equal("member_1"))
			Expect(member["rank"]).To(BeEquivalentTo(1))
			Expect(member["score"]).To(BeEquivalentTo(100))

			members = result["notFound"].([]interface{})
			Expect(members).To(HaveLen(1))
			Expect(members[0]).To(Equal("invalid_member"))
		})

		It("should fail if no public ids sent", func() {
			leaderboardID := uuid.NewV4().String()
			l = leaderboard.NewLeaderboard(a.RedisClient, leaderboardID, 10, lg)

			res := api.Get(a, fmt.Sprintf("/l/%s/members/", l.PublicID))
			Expect(res.Raw().StatusCode).To(Equal(http.StatusBadRequest))

			var result map[string]interface{}
			json.Unmarshal([]byte(res.Body().Raw()), &result)

			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Member IDs are required using the 'ids' querystring parameter"))
		})

		It("Should fail if error in Redis", func() {
			app := api.GetDefaultTestApp()
			app.RedisClient = api.GetFaultyRedis(a.Logger)

			res := api.Get(app, "/l/invalid-redis/members/", map[string]interface{}{
				"ids": "member_10,member_20,member_30",
			})

			Expect(res.Raw().StatusCode).To(Equal(500))
			Expect(res.Body().Raw()).To(ContainSubstring("connection refused"))
		})

		Measure("it should get members", func(b Benchmarker) {
			leaderboardID := uuid.NewV4().String()
			l = leaderboard.NewLeaderboard(a.RedisClient, leaderboardID, 10, lg)

			memberIDs := []string{}
			for i := 1; i <= 1000; i++ {
				memberID := fmt.Sprintf("member_%d", i)
				memberIDs = append(memberIDs, memberID)
				_, err := l.SetMemberScore(memberID, 101-i)
				Expect(err).NotTo(HaveOccurred())
			}

			mIDs := strings.Join(memberIDs, ",")

			runtime := b.Time("runtime", func() {
				res := api.Get(a, fmt.Sprintf("/l/%s/members/", l.PublicID), map[string]interface{}{
					"ids": mIDs,
				})
				Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.08), "Getting members shouldn't take too long.")
		}, 200)
	})
})
