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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/topfreegames/extensions/redis/interfaces"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/podium/api"
	. "github.com/topfreegames/podium/testing"

	pb "github.com/topfreegames/podium/proto/podium/api/v1"
)

var _ = Describe("Leaderboard Handler", func() {
	var a *api.App
	var redisClient interfaces.RedisClient
	const testLeaderboardID = "testkey"

	BeforeSuite(func() {
		a = GetDefaultTestApp()
	})

	BeforeEach(func() {
		extRedisClient, err := GetConnectedRedis()
		Expect(err).NotTo(HaveOccurred())
		redisClient = extRedisClient.Client
		redisClient.Del("testkey")
		redisClient.Del("testkey1")
		redisClient.Del("testkey2")
		redisClient.Del("testkey3")
		redisClient.Del("testkey4")
		redisClient.Del("testkey5")
	})

	Describe("When leaderboard has expired", func() {
		var (
			maybePrefixWithZero = func(i int) string {
				if i < 10 {
					return fmt.Sprintf("0%d", i)
				}
				return fmt.Sprintf("%d", i)
			}
			year, week               = time.Now().UTC().AddDate(0, 0, -15).ISOWeek()
			lastQuarter, quarterYear = func() (int, int) {
				quarter := int(time.Now().UTC().Month()-1)/3 + 1
				quarterYear := time.Now().UTC().Year()
				if quarter-2 <= 0 {
					quarterYear--
					quarter = 4 + (quarter - 2)
				} else {
					quarter -= 2
				}
				return quarter, quarterYear
			}()
			keys = []string{
				fmt.Sprintf(
					"testkey-from%dto%d",
					time.Now().UTC().Add(time.Duration(-2)*time.Second).Unix(),
					time.Now().UTC().Add(time.Duration(-1)*time.Second).Unix(),
				),
				fmt.Sprintf("testkey-from20180101to20180105"),
				fmt.Sprintf(
					"testkey-year%d",
					time.Now().UTC().AddDate(-2, 0, 0).Year(),
				),
				fmt.Sprintf("testkey-year%dweek%s", year, maybePrefixWithZero(week)),
				fmt.Sprintf(
					"testkey-year%dmonth%s",
					time.Now().UTC().AddDate(0, -2, 0).Year(),
					maybePrefixWithZero(int(time.Now().UTC().AddDate(0, -2, 0).Month())),
				),
				fmt.Sprintf(
					"testkey-year%dquarter0%d",
					quarterYear,
					lastQuarter,
				),
			}
		)

		checkBody := func(key, body string) {
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(
				Equal(
					fmt.Sprintf("Leaderboard %s has already expired", key),
				),
			)
		}

		It("PUT upsert score", func() {
			payload := map[string]interface{}{"score": int64(100)}
			for _, k := range keys {
				httpPath := fmt.Sprintf("/l/%s/members/memberpublicid/score", k)
				status, body := PutJSON(a, httpPath, payload)
				Expect(status).To(Equal(http.StatusBadRequest))
				checkBody(k, body)
			}
		})

		It("PUT upsert members scores", func() {
			payload := map[string]interface{}{"members": []map[string]interface{}{
				{"publicID": "memberpublicid", "score": 0},
			}}
			for _, k := range keys {
				httpPath := fmt.Sprintf("/l/%s/scores", k)
				status, body := PutJSON(a, httpPath, payload)
				Expect(status).To(Equal(http.StatusBadRequest))
				checkBody(k, body)
			}
		})

		It("PATCH increment score", func() {
			payload := map[string]interface{}{"increment": 100}
			for _, k := range keys {
				httpPath := fmt.Sprintf("/l/%s/members/memberpublicid/score", k)
				status, body := PatchJSON(a, httpPath, payload)
				Expect(status).To(Equal(http.StatusBadRequest))
				checkBody(k, body)
			}
		})
	})

	Describe("Bulk Upsert Members Score", func() {
		It("Should set correct members score in redis and respond with the correct values", func() {
			payload := map[string]interface{}{"members": []map[string]interface{}{
				{"publicID": "memberpublicid1", "score": int64(150)},
				{"publicID": "memberpublicid2", "score": int64(100)},
			}}
			status, body := PutJSON(a, "/l/testkey/scores", payload)
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			for i, memberObj := range result["members"].([]interface{}) {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(int64(member["score"].(float64))).To(Equal(payload["members"].([]map[string]interface{})[i]["score"].(int64)))
				Expect(member["publicID"]).To(Equal(payload["members"].([]map[string]interface{})[i]["publicID"].(string)))
			}

			member1, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid1", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member1.Rank).To(Equal(1))
			Expect(member1.Score).To(Equal(int64(150)))
			Expect(member1.PublicID).To(Equal("memberpublicid1"))

			member2, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid2", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member2.Rank).To(Equal(2))
			Expect(member2.Score).To(Equal(int64(100)))
			Expect(member2.PublicID).To(Equal("memberpublicid2"))
		})

		It("Should set correct members scores in redis and respond with the correct values if bigger than int", func() {
			bigScore1 := int64(15584657100002)
			bigScore2 := int64(15584657100001)
			payload := map[string]interface{}{"members": []map[string]interface{}{
				{"publicID": "memberpublicid1", "score": bigScore1},
				{"publicID": "memberpublicid2", "score": bigScore2},
			}}
			status, body := PutJSON(a, "/l/testkey/scores", payload)
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			for i, memberObj := range result["members"].([]interface{}) {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(int64(member["score"].(float64))).To(Equal(payload["members"].([]map[string]interface{})[i]["score"].(int64)))
				Expect(member["publicID"]).To(Equal(payload["members"].([]map[string]interface{})[i]["publicID"].(string)))
			}

			member1, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid1", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member1.Rank).To(Equal(1))
			Expect(member1.Score).To(Equal(bigScore1))
			Expect(member1.PublicID).To(Equal("memberpublicid1"))

			member2, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid2", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member2.Rank).To(Equal(2))
			Expect(member2.Score).To(Equal(bigScore2))
			Expect(member2.PublicID).To(Equal("memberpublicid2"))
		})

		It("Should insert successfully with expiration if scoreTTL argument is sent", func() {
			ttl := 100
			lbName := "testkey"

			payload := map[string]interface{}{"members": []map[string]interface{}{
				{"publicID": "memberpublicid1", "score": int64(150)},
				{"publicID": "memberpublicid2", "score": int64(100)},
			}}

			status, body := PutJSON(a, fmt.Sprintf("/l/%s/scores?scoreTTL=%d", lbName, ttl), payload)
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			for i, memberObj := range result["members"].([]interface{}) {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(int64(member["score"].(float64))).To(Equal(payload["members"].([]map[string]interface{})[i]["score"].(int64)))
				Expect(member["publicID"]).To(Equal(payload["members"].([]map[string]interface{})[i]["publicID"].(string)))
				Expect(int(member["expireAt"].(float64))).To(BeNumerically("~", time.Now().Unix()+int64(ttl), 1))

				memb, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", true)
				Expect(err).NotTo(HaveOccurred())
				Expect(memb.Rank).To(Equal(i + 1))
				Expect(memb.Score).To(Equal(payload["members"].([]map[string]interface{})[i]["score"].(int64)))
				Expect(memb.PublicID).To(Equal(member["publicID"]))
				Expect(memb.ExpireAt).To(BeNumerically("~", time.Now().Unix()+int64(ttl), 1))
			}

			redisLBExpirationKey := fmt.Sprintf("%s:ttl", lbName)
			result2, err := redisClient.Exists(redisLBExpirationKey).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result2).To(Equal(int64(1)))
			redisExpirationSetKey := "expiration-sets"
			result2, err = redisClient.Exists(redisExpirationSetKey).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result2).To(Equal(int64(1)))
			result3, err := redisClient.SMembers(redisExpirationSetKey).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result3).To(ContainElement(redisLBExpirationKey))
			result4, err := redisClient.ZScore(redisLBExpirationKey, "memberpublicid1").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result4).To(BeNumerically("~", time.Now().Unix()+int64(ttl), 1))
			result5, err := redisClient.ZScore(redisLBExpirationKey, "memberpublicid2").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result5).To(BeNumerically("~", time.Now().Unix()+int64(ttl), 1))

		})

		It("Should set correct members scores in redis and respond with previous rank", func() {
			payload1 := map[string]interface{}{"members": []map[string]interface{}{
				{"publicID": "memberpublicid1", "score": int64(200)},
				{"publicID": "memberpublicid2", "score": int64(150)},
				{"publicID": "memberpublicid3", "score": int64(100)},
			}}
			payload2 := map[string]interface{}{"members": []map[string]interface{}{
				{"publicID": "memberpublicid2", "score": int64(300)},
				{"publicID": "memberpublicid3", "score": int64(250)},
			}}

			status, body := PutJSON(a, "/l/testkey/scores", payload1)
			Expect(status).To(Equal(http.StatusOK), body)
			status, body = PutJSON(a, "/l/testkey/scores?prevRank=true", payload2)
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			for i, memberObj := range result["members"].([]interface{}) {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(int64(member["score"].(float64))).To(Equal(payload2["members"].([]map[string]interface{})[i]["score"].(int64)))
				Expect(member["publicID"]).To(Equal(payload2["members"].([]map[string]interface{})[i]["publicID"].(string)))
				Expect(int(member["previousRank"].(float64))).To(Equal(i + 2))

				memb, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(memb.Rank).To(Equal(i + 1))
				Expect(memb.Score).To(Equal(int64(member["score"].(float64))))
				Expect(memb.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should work when setting scores to 0", func() {
			payload := map[string]interface{}{"members": []map[string]interface{}{
				{"publicID": "memberpublicid1", "score": int64(0)},
				{"publicID": "memberpublicid2", "score": int64(0)},
			}}
			status, body := PutJSON(a, "/l/testkey/scores", payload)
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			for i, memberObj := range result["members"].([]interface{}) {
				member := memberObj.(map[string]interface{})
				Expect(int64(member["score"].(float64))).To(Equal(int64(0)))
				Expect(member["publicID"]).To(Equal(payload["members"].([]map[string]interface{})[i]["publicID"].(string)))

				memb, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(memb.Score).To(Equal(int64(0)))
				Expect(memb.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should fail if wrong type for score", func() {
			payload := map[string]interface{}{"members": []map[string]interface{}{
				{"publicID": "memberpublicid1", "score": "hundred"},
				{"publicID": "memberpublicid2", "score": "fifty"},
			}}
			status, body := PutJSON(a, "/l/testkey/scores", payload)
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring("invalid character"))
		})

		It("Should fail if missing parameters", func() {
			payload := map[string]interface{}{"members": []map[string]interface{}{
				{"score": 100},
			}}
			status, body := PutJSON(a, "/l/testkey/scores", payload)
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("publicID is required"))
		})

		It("Should fail if invalid payload", func() {
			status, body := Put(a, "/l/testkey/scores", "invalid")
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring("invalid character"))
		})

		It("Should fail if error updating score", func() {
			payload := map[string]interface{}{"members": []map[string]interface{}{
				{"publicID": "memberpublicid1", "score": int64(0)},
				{"publicID": "memberpublicid2", "score": int64(0)},
			}}
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := PutJSON(app, "/l/testkey/scores", payload)
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		HTTPMeasure("it should update member score", func(ctx map[string]interface{}) {
			payload := map[string]interface{}{"members": []map[string]interface{}{
				{"publicID": "memberpublicid1", "score": int64(150)},
				{"publicID": "memberpublicid2", "score": int64(100)},
			}}
			payloadJSON, err := json.Marshal(payload)
			Expect(err).NotTo(HaveOccurred())
			ctx["payload"] = payloadJSON
		}, func(ts *httptest.Server, ctx map[string]interface{}) {
			url := getRoute(ts, "/l/testkey/scores")
			status, body, err := fastPutTo(url, ctx["payload"].([]byte))
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.05)
	})

	Describe("Upsert Member Score", func() {
		It("Should set correct member score in redis and respond with the correct values", func() {
			payload := map[string]interface{}{
				"score": int64(100),
			}
			status, body := PutJSON(a, "/l/testkey/members/memberpublicid/score", payload)
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int64(result["score"].(float64))).To(Equal(payload["score"]))
			Expect(int(result["rank"].(float64))).To(Equal(1))
			Expect(result).NotTo(HaveKey("previousRank"))

			member, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(int64(100)))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should set correct member score in redis and respond with the correct values if bigger than int", func() {
			bigScore := int64(15584657100001)
			payload := map[string]interface{}{
				"score": bigScore,
			}
			status, body := PutJSON(a, "/l/testkey/members/memberpublicid/score", payload)
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int64(result["score"].(float64))).To(Equal(payload["score"]))
			Expect(int(result["rank"].(float64))).To(Equal(1))
			Expect(result).NotTo(HaveKey("previousRank"))

			member, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(bigScore))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should insert successfully with expiration if scoreTTL argument is sent", func() {
			ttl := 100
			lbName := "testkey"

			payload := map[string]interface{}{
				"score": int64(100),
			}
			status, body := PutJSON(a, fmt.Sprintf("/l/%s/members/memberpublicid/score?scoreTTL=%d", lbName, ttl), payload)
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int64(result["score"].(float64))).To(Equal(payload["score"]))
			Expect(int(result["rank"].(float64))).To(Equal(1))
			Expect(int(result["expireAt"].(float64))).To(BeNumerically("~", time.Now().Unix()+int64(ttl), 1))

			member, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", true)
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(int64(100)))
			Expect(member.PublicID).To(Equal("memberpublicid"))
			Expect(member.ExpireAt).To(BeNumerically("~", time.Now().Unix()+int64(ttl), 1))

			redisLBExpirationKey := fmt.Sprintf("%s:ttl", lbName)
			result2, err := redisClient.Exists(redisLBExpirationKey).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result2).To(Equal(int64(1)))
			redisExpirationSetKey := "expiration-sets"
			result2, err = redisClient.Exists(redisExpirationSetKey).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result2).To(Equal(int64(1)))
			result3, err := redisClient.SMembers(redisExpirationSetKey).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result3).To(ContainElement(redisLBExpirationKey))
			result4, err := redisClient.ZScore(redisLBExpirationKey, "memberpublicid").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result4).To(BeNumerically("~", time.Now().Unix()+int64(ttl), 1))
		})

		It("Should set correct member score in redis and respond with previous rank", func() {
			payload1 := map[string]interface{}{
				"score": int64(100),
			}
			payload2 := map[string]interface{}{
				"score": int64(50),
			}
			payload3 := map[string]interface{}{
				"score": int64(10),
			}
			status, body := PutJSON(a, "/l/testkey/members/memberpublicid/score", payload1)
			Expect(status).To(Equal(http.StatusOK), body)
			status, body = PutJSON(a, "/l/testkey/members/memberpublicid2/score", payload2)
			Expect(status).To(Equal(http.StatusOK), body)
			status, body = PutJSON(a, "/l/testkey/members/memberpublicid/score?prevRank=true", payload3)
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int64(result["score"].(float64))).To(Equal(payload3["score"]))
			Expect(int(result["rank"].(float64))).To(Equal(2))
			Expect(int(result["previousRank"].(float64))).To(Equal(1))

			member, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(2))
			Expect(member.Score).To(Equal(int64(10)))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should work when setting score to 0", func() {
			payload := map[string]interface{}{
				"score": int64(0),
			}
			status, body := PutJSON(a, "/l/testkey/members/memberpublicid/score", payload)
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int64(result["score"].(float64))).To(Equal(payload["score"]))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			member, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(int64(0)))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should fail if wrong type for score", func() {
			payload := map[string]interface{}{
				"score": "100",
			}
			status, body := PutJSON(a, "/l/testkey/members/memberpublicid/score", payload)
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("invalid type for score: string"))
		})

		It("Should fail if missing parameters", func() {
			payload := map[string]interface{}{
				"notscore": 100,
			}
			status, body := PutJSON(a, "/l/testkey/members/memberpublicid/score", payload)
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("score is required"))
		})

		It("Should fail if invalid payload", func() {
			status, body := Put(a, "/l/testkey/members/memberpublicid/score", "invalid")
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring("score is required"))
		})

		It("Should fail if error updating score", func() {
			payload := map[string]interface{}{
				"score": int64(100),
			}
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := PutJSON(app, "/l/testkey/members/memberpublicid/score", payload)
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		HTTPMeasure("it should update member score", func(ctx map[string]interface{}) {
			payload := map[string]interface{}{
				"score": int64(100),
			}
			payloadJSON, err := json.Marshal(payload)
			Expect(err).NotTo(HaveOccurred())
			ctx["payload"] = payloadJSON
		}, func(ts *httptest.Server, ctx map[string]interface{}) {
			url := getRoute(ts, "/l/testkey/members/memberpublicid/score")
			status, body, err := fastPutTo(url, ctx["payload"].([]byte))
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.05)
	})

	Describe("Increment Member Score", func() {
		It("Should increment correct member score in redis and respond with the correct values", func() {
			payload := map[string]interface{}{
				"increment": 10,
			}

			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "memberpublicid", 100, false, "")
			Expect(err).NotTo(HaveOccurred())

			status, body := PatchJSON(a, "/l/testkey/members/memberpublicid/score", payload)
			Expect(status).To(Equal(http.StatusOK), body)

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int64(result["score"].(float64))).To(Equal(int64(110)))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			member, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(int64(110)))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should increment correct member score when member does not exist", func() {
			payload := map[string]interface{}{
				"increment": 10,
			}

			status, body := PatchJSON(a, "/l/testkey/members/memberpublicid/score", payload)
			Expect(status).To(Equal(http.StatusOK), body)

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int64(result["score"].(float64))).To(Equal(int64(10)))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			member, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(int64(10)))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should not work when incrementing by 0", func() {
			payload := map[string]interface{}{
				"increment": 0,
			}
			status, body := PatchJSON(a, "/l/testkey/members/memberpublicid/score", payload)
			Expect(status).To(Equal(http.StatusBadRequest), body)

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("increment is required"))
		})

		It("Should fail if missing parameters", func() {
			payload := map[string]interface{}{
				"notscore": 100,
			}
			status, body := PatchJSON(a, "/l/testkey/members/memberpublicid/score", payload)
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("increment is required"))
		})

		It("Should fail if invalid payload", func() {
			status, body := Patch(a, "/l/testkey/members/memberpublicid/score", "invalid")
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring("parse error: syntax error"))
		})

		It("Should fail if error updating score", func() {
			payload := map[string]interface{}{
				"increment": 100,
			}
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := PatchJSON(app, "/l/testkey/members/memberpublicid/score", payload)
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		HTTPMeasure("it should update member score", func(ctx map[string]interface{}) {
			payload := map[string]interface{}{
				"increment": 100,
			}
			payloadJSON, err := json.Marshal(payload)
			Expect(err).NotTo(HaveOccurred())
			ctx["payload"] = payloadJSON
		}, func(ts *httptest.Server, ctx map[string]interface{}) {
			url := getRoute(ts, "/l/testkey/members/memberpublicid/score")
			status, body, err := fastPatchTo(url, ctx["payload"].([]byte))
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.05)
	})

	Describe("Remove Member Score", func() {
		It("Should delete member score from redis if score exists", func() {
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "memberpublicid", 100, false, "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Delete(a, "/l/testkey/members?ids=memberpublicid")
			Expect(status).To(Equal(http.StatusOK), body, body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			_, err = a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member"))
		})

		It("Should delete member score from redis if score exists", func() {
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "memberpublicid", 100, false, "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Delete(a, "/l/testkey/members/memberpublicid")
			Expect(status).To(Equal(http.StatusOK), body, body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			_, err = a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member"))
		})

		It("Should delete many member score from redis if they exists", func() {
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "memberpublicid", 100, false, "")
			_, err = a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "memberpublicid2", 100, false, "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Delete(a, "/l/testkey/members?ids=memberpublicid,memberpublicid2")
			Expect(status).To(Equal(http.StatusOK), body, body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			_, err = a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member"))
			_, err = a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid2", "desc", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member"))
		})

		It("Should fail if error removing score", func() {
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "memberpublicid", 100, false, "")
			Expect(err).NotTo(HaveOccurred())

			app := GetDefaultTestAppWithFaultyRedis()

			status, body := Delete(app, "/l/testkey/members?ids=memberpublicid")
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		It("Should not fail in deleting member score from redis if score does not exist", func() {
			status, body := Delete(a, "/l/testkey/members?ids=memberpublicid")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())

			_, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member"))
		})

		HTTPMeasure("it should remove member score", func(ctx map[string]interface{}) {
			lbID := uuid.NewV4().String()
			memberID := uuid.NewV4().String()
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), lbID, memberID, 100, false, "")
			Expect(err).NotTo(HaveOccurred())
			ctx["lead"] = lbID
			ctx["memberID"] = memberID
		}, func(ts *httptest.Server, ctx map[string]interface{}) {
			lead := ctx["lead"].(string)
			memberID := ctx["memberID"].(string)
			url := getRoute(ts, fmt.Sprintf("/l/%s/members?ids=%s", lead, memberID))
			status, body, err := fastDelete(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.05)
	})

	Describe("Get Member", func() {
		It("Should get member score from redis if score exists", func() {
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "memberpublicid", 100, false, "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, "/l/testkey/members/memberpublicid")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int64(result["score"].(float64))).To(Equal(int64(100)))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			member, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(int64(100)))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should get member score from redis if greater than int", func() {
			bigScore := int64(15584657100001)
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "memberpublicid", bigScore, false, "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, "/l/testkey/members/memberpublicid")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int64(result["score"].(float64))).To(Equal(bigScore))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			member, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(bigScore))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should get member score from redis if score exists including expiration timestamp", func() {
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "memberpublicid", 100, false, "15")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, "/l/testkey/members/memberpublicid?scoreTTL=true")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int64(result["score"].(float64))).To(Equal(int64(100)))
			Expect(int(result["rank"].(float64))).To(Equal(1))
			Expect(int(result["expireAt"].(float64))).To(BeNumerically("~", time.Now().Unix()+15, 1))

			member, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(int64(100)))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should get member score from redis if score exists including expiration timestamp if no ttl", func() {
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "memberpublicidnottl", 100, false, "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, "/l/testkey/members/memberpublicidnottl?scoreTTL=true")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicidnottl"))
			Expect(int64(result["score"].(float64))).To(Equal(int64(100)))
			Expect(int(result["rank"].(float64))).To(Equal(1))
			Expect(int(result["expireAt"].(float64))).To(Equal(0))
		})

		It("Should fail with 404 if score does not exist", func() {
			status, body := Get(a, "/l/testkey/members/memberpublicid")
			Expect(status).To(Equal(http.StatusNotFound), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Member not found."))
		})

		It("Should fail if error in Redis", func() {
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := Get(app, "/l/testkey/members/member_99")
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		HTTPMeasure("it should get member", func(ctx map[string]interface{}) {
			lbID := uuid.NewV4().String()
			memberID := uuid.NewV4().String()
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), lbID, memberID, 500, false, "")
			Expect(err).NotTo(HaveOccurred())

			ctx["lead"] = lbID
			ctx["memberID"] = memberID
		}, func(ts *httptest.Server, ctx map[string]interface{}) {
			lead := ctx["lead"].(string)
			memberID := ctx["memberID"].(string)
			url := getRoute(ts, fmt.Sprintf("/l/%s/members/%s", lead, memberID))
			status, body, err := fastGet(url)
			Expect(status).To(Equal(http.StatusOK), string(body))
			Expect(err).NotTo(HaveOccurred())
		}, 0.05)
	})

	Describe("Get Member Rank", func() {
		It("Should get member score from redis if score exists", func() {
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "memberpublicid", 100, false, "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, "/l/testkey/members/memberpublicid/rank")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			member, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(int64(100)))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should get member score from redis if score exists and order is asc", func() {
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "memberpublicid", 100, false, "")
			Expect(err).NotTo(HaveOccurred())

			status, body := Get(a, "/l/testkey/members/memberpublicid/rank?order=asc")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(result["publicID"]).To(Equal("memberpublicid"))
			Expect(int(result["rank"].(float64))).To(Equal(1))

			member, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "memberpublicid", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Rank).To(Equal(1))
			Expect(member.Score).To(Equal(int64(100)))
			Expect(member.PublicID).To(Equal("memberpublicid"))
		})

		It("Should fail with 404 if score does not exist", func() {
			status, body := Get(a, "/l/testkey/members/memberpublicid/rank")
			Expect(status).To(Equal(http.StatusNotFound), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Member not found."))
		})

		It("Should fail if error in Redis", func() {
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := Get(app, "/l/testkey/members/member_99/rank")
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		HTTPMeasure("it should get member rank", func(ctx map[string]interface{}) {
			lbID := uuid.NewV4().String()
			memberID := uuid.NewV4().String()
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), lbID, memberID, 500, false, "")
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < 10; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), lbID, fmt.Sprintf("member-%d", i), 500, false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			ctx["lead"] = lbID
			ctx["memberID"] = memberID
		}, func(ts *httptest.Server, ctx map[string]interface{}) {
			lead := ctx["lead"].(string)
			memberID := ctx["memberID"].(string)
			url := getRoute(ts, fmt.Sprintf("/l/%s/members/%s/rank", lead, memberID))
			status, body, err := fastGet(url)
			Expect(status).To(Equal(http.StatusOK), string(body))
			Expect(err).NotTo(HaveOccurred())
		}, 0.05)
	})

	Describe("Get Around Member Handler", func() {
		It("Should get member score and neighbours from redis if member score exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/members/member_50/around")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
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

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get member score and neighbours from redis in reverse order if member score exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/members/member_50/around?order=asc")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
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
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(16-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/members/member_10/around")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(15))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", i+1)))
				Expect(int(member["score"].(float64))).To(Equal(15 - i))

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get member score and default limit neighbours from redis if member score and less than limit neighbours exist", func() {
			for i := 1; i <= 15; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(16-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/members/member_10/around")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(15))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				pos := i
				Expect(int(member["rank"].(float64))).To(Equal(pos + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", pos+1)))
				Expect(int(member["score"].(float64))).To(Equal(15 - pos))

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get member score and limit neighbours from redis if member score exists and custom limit", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/members/member_50/around?pageSize=10")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
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

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get member score and limit neighbours from redis if member score exists and repeated scores", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/members/member_50/around?pageSize=10")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(10))
			start := int(members[0].(map[string]interface{})["rank"].(float64))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				pos := start + i
				Expect(int(member["rank"].(float64))).To(Equal(pos))
				Expect(int64(member["score"].(float64))).To(Equal(int64(100)))

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get last positions if not in ranking", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(100-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/members/member_999/around?pageSize=20&getLastIfNotFound=true")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			start := int(members[0].(map[string]interface{})["rank"].(float64))
			Expect(start).To(Equal(81))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				pos := start + i
				Expect(int(member["rank"].(float64))).To(Equal(pos))
				Expect(int(member["score"].(float64))).To(Equal(100 - pos))

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should fail with 404 if score for member does not exist", func() {
			status, body := Get(a, "/l/testkey/members/memberpublicid/around")
			Expect(status).To(Equal(http.StatusNotFound), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Member not found."))
		})

		It("Should fail with 400 if bad pageSize provided", func() {
			status, body := Get(a, "/l/testkey/members/member_50/around?pageSize=notint")
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			err := json.Unmarshal([]byte(body), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Failed to process param pageSize: notint"))
		})

		It("Should fail with 400 if pageSize provided if bigger than maxPageSizeAllowed", func() {
			pageSize := a.Config.GetInt("api.maxReturnedMembers") + 1
			status, body := Get(a, fmt.Sprintf("/l/testkey/members/member_50/around?pageSize=%d", pageSize))
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Max pageSize allowed: %d. pageSize requested: %d", pageSize-1, pageSize)))
		})

		It("Should get one page of top members from redis if leaderboard exists and member in ranking bottom", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(100-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/members/member_2/around")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", i+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - i - 1))

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get one page of top members from redis if leaderboard exists and member in ranking top", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(100-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/members/member_99/around")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(80 + i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", 80+i+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - 80 - i - 1))

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should fail if error in Redis", func() {
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := Get(app, "/l/testkey/members/member_99/around")
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		It("Should fail if error in Redis", func() {
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := Get(app, "/l/testkey/members/member_99/around?getLastIfNotFound=true")
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		HTTPMeasure("it should get around member", func(ctx map[string]interface{}) {
			lead := uuid.NewV4().String()
			memberID := uuid.NewV4().String()
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), lead, memberID, 500, false, "")
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < 10; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), lead, fmt.Sprintf("member-%d", i), 500, false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			ctx["lead"] = lead
			ctx["memberID"] = memberID
		}, func(ts *httptest.Server, ctx map[string]interface{}) {
			lead := ctx["lead"].(string)
			memberID := ctx["memberID"].(string)
			url := getRoute(ts, fmt.Sprintf("/l/%s/members/%s/around", lead, memberID))
			status, body, err := fastGet(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.05)
	})

	Describe("Get Around Score Handler", func() {
		It("Should get score neighbours from redis if score is sent", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			score := 50
			status, body := Get(a, fmt.Sprintf("/l/testkey/scores/%d/around", score))
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			rank := score + 1
			start := rank - 20/2
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				pos := start + i
				Expect(int(member["rank"].(float64))).To(Equal(pos + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", pos+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - pos))

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get rank neighbours from redis in reverse order if score is sent", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			score := 50
			status, body := Get(a, fmt.Sprintf("/l/testkey/scores/%d/around?order=asc", score))
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			rank := score + 1
			start := rank + 20/2
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				pos := start - i
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", pos-1)))
			}
		})

		It("Should get one page of top members from redis if leaderboard exists but less than pageSize neighbours exist", func() {
			for i := 1; i <= 15; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(16-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			score := 10
			status, body := Get(a, fmt.Sprintf("/l/testkey/scores/%d/around", score))
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(15))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", i+1)))
				Expect(int(member["score"].(float64))).To(Equal(15 - i))

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should limit neighbours from redis if score is sent and custom limit", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			score := 50
			status, body := Get(a, fmt.Sprintf("/l/testkey/scores/%d/around?pageSize=10", score))
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(10))
			start := 50 + 1 - 10/2
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				pos := start + i
				Expect(int(member["rank"].(float64))).To(Equal(pos + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", pos+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - pos))

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should fail with 400 if bad score is provided", func() {
			status, body := Get(a, "/l/testkey/scores/badScore/around")
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			err := json.Unmarshal([]byte(body), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Score not sent or wrongly formatted"))
		})

		It("Should fail with 400 if bad pageSize provided", func() {
			status, body := Get(a, "/l/testkey/scores/50/around?pageSize=notint")
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			err := json.Unmarshal([]byte(body), &result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Failed to process param pageSize: notint"))
		})

		It("Should fail with 400 if pageSize provided if bigger than maxPageSizeAllowed", func() {
			pageSize := a.Config.GetInt("api.maxReturnedMembers") + 1
			status, body := Get(a, fmt.Sprintf("/l/testkey/scores/50/around?pageSize=%d", pageSize))
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Max pageSize allowed: %d. pageSize requested: %d", pageSize-1, pageSize)))
		})

		It("Should get one page of top members from redis if leaderboard exists and score <= 0", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(100-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			scores := []int{-2, 0}

			for _, score := range scores {
				status, body := Get(a, fmt.Sprintf("/l/testkey/scores/%d/around", score))
				Expect(status).To(Equal(http.StatusOK), body)
				var result map[string]interface{}
				json.Unmarshal([]byte(body), &result)
				Expect(result["success"]).To(BeTrue())
				members := result["members"].([]interface{})
				Expect(len(members)).To(Equal(20))

				start := 100 - 20
				for i, memberObj := range members {
					member := memberObj.(map[string]interface{})
					pos := start + i + 1
					Expect(int(member["rank"].(float64))).To(Equal(pos))
					Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", pos)))
					Expect(int(member["score"].(float64))).To(Equal(20 - i - 1))

					dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
					Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
					Expect(dbMember.PublicID).To(Equal(member["publicID"]))
				}
			}
		})

		It("Should get one page of top members from redis if leaderboard exists and score in ranking top", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(100-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			scores := []int{99, 100, 500}

			for _, score := range scores {
				status, body := Get(a, fmt.Sprintf("/l/testkey/scores/%d/around", score))
				Expect(status).To(Equal(http.StatusOK), body)
				var result map[string]interface{}
				json.Unmarshal([]byte(body), &result)
				Expect(result["success"]).To(BeTrue())
				members := result["members"].([]interface{})
				Expect(len(members)).To(Equal(20))
				for i, memberObj := range members {
					member := memberObj.(map[string]interface{})
					pos := i + 1
					Expect(int(member["rank"].(float64))).To(Equal(pos))
					Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", pos)))
					Expect(int(member["score"].(float64))).To(Equal(100 - i - 1))

					dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
					Expect(err).NotTo(HaveOccurred())
					Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
					Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
					Expect(dbMember.PublicID).To(Equal(member["publicID"]))
				}
			}
		})

		It("Should fail if error in Redis", func() {
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := Get(app, "/l/testkey/scores/50/around")
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})
	})

	Describe("Get Total Members Handler", func() {
		It("Should get the number of members in a leaderboard it exists (http)", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, fmt.Sprintf("/l/%s/members-count", testLeaderboardID))
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(int(result["count"].(float64))).To(Equal(100))
		})

		It("Should get the number of members in a leaderboard it exists (grpc)", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			SetupGRPC(a, func(cli pb.PodiumAPIClient) {
				resp, err := cli.TotalMembers(context.Background(),
					&pb.TotalMembersRequest{LeaderboardID: testLeaderboardID})

				Expect(err).NotTo(HaveOccurred())
				Expect(int(resp.Count)).To(Equal(100))
			})
		})

		It("Should not fail if leaderboard does not exist", func() {
			status, body := Get(a, "/l/testkey/members-count")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			Expect(int(result["count"].(float64))).To(Equal(0))
		})

		It("Should fail if error in Redis", func() {
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := Get(app, "/l/testkey/members-count")
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		HTTPMeasure("it should get total members", func(ctx map[string]interface{}) {
			lead := uuid.NewV4().String()
			memberID := uuid.NewV4().String()
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), lead, memberID, 500, false, "")
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < 10; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), lead, fmt.Sprintf("member-%d", i), 500, false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			ctx["lead"] = lead
		}, func(ts *httptest.Server, ctx map[string]interface{}) {
			lead := ctx["lead"].(string)
			url := getRoute(ts, fmt.Sprintf("/l/%s/members-count", lead))
			status, body, err := fastGet(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.05)
	})

	Describe("Get Top Members Handler", func() {
		It("Should get one page of top members from redis if leaderboard exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/top/1")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(20))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", i+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - i))

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get one page of top members in reverse order from redis if leaderboard exists", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/top/1?order=asc")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
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
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/top/2")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
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

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get top members from redis if leaderboard exists with custom pageSize", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/top/1?pageSize=10")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(10))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(member["publicID"]).To(Equal(fmt.Sprintf("member_%d", i+1)))
				Expect(int(member["score"].(float64))).To(Equal(100 - i))

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should get empty list if page does not exist", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/top/100000")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(0))
		})

		It("Should get only one page of top members from redis if leaderboard exists and repeated scores", func() {
			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, "/l/testkey/top/1?pageSize=10")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(10))
			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(int64(member["score"].(float64))).To(Equal(int64(100)))

				dbMember, err := a.Leaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, member["publicID"].(string), "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(dbMember.Rank).To(Equal(int(member["rank"].(float64))))
				Expect(dbMember.Score).To(Equal(int64(member["score"].(float64))))
				Expect(dbMember.PublicID).To(Equal(member["publicID"]))
			}
		})

		It("Should fail with 400 if bad pageNumber provided", func() {
			status, body := Get(a, "/l/testkey/top/notint")
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Failed to process route param pageNumber: notint"))
		})

		It("Should fail with 400 if bad pageSize provided", func() {
			status, body := Get(a, "/l/testkey/top/1?pageSize=notint")
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Failed to process param pageSize: notint"))
		})

		It("Should fail with 400 if pageSize provided if bigger than maxPageSizeAllowed", func() {
			pageSize := a.Config.GetInt("api.maxReturnedMembers") + 1
			status, body := Get(a, fmt.Sprintf("/l/testkey/top/1?pageSize=%d", pageSize))
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal(fmt.Sprintf("Max pageSize allowed: %d. pageSize requested: %d", pageSize-1, pageSize)))
		})

		It("Should fail if error getting top members from Redis", func() {
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := Get(app, "/l/testkey/top/1")
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		HTTPMeasure("it should get top members", func(ctx map[string]interface{}) {
			lead := uuid.NewV4().String()
			memberID := uuid.NewV4().String()
			_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), lead, memberID, 500, false, "")
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), lead, fmt.Sprintf("member-%d", i), 500, false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			ctx["lead"] = lead
		}, func(ts *httptest.Server, ctx map[string]interface{}) {
			lead := ctx["lead"].(string)
			url := getRoute(ts, fmt.Sprintf("/l/%s/top/10", lead))
			status, body, err := fastGet(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.05)
	})

	Describe("Get Top Percentage Handler", func() {
		It("Should get top members from redis if leaderboard exists", func() {
			leaderboardID := uuid.NewV4().String()

			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("member_%d", i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, fmt.Sprintf("/l/%s/top-percent/10", leaderboardID))
			Expect(status).To(Equal(http.StatusOK), body)

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

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

			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("member_%d", i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(a, fmt.Sprintf("/l/%s/top-percent/10", leaderboardID))
			Expect(status).To(Equal(http.StatusOK), body)

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeTrue())
			members := result["members"].([]interface{})
			Expect(len(members)).To(Equal(10))

			for i, memberObj := range members {
				member := memberObj.(map[string]interface{})
				Expect(int(member["rank"].(float64))).To(Equal(i + 1))
				Expect(int64(member["score"].(float64))).To(Equal(int64(100)))
			}
		})

		It("Should fail if invalid percentage", func() {
			leaderboardID := uuid.NewV4().String()

			status, body := Get(a, fmt.Sprintf("/l/%s/top-percent/l", leaderboardID))
			Expect(status).To(Equal(http.StatusBadRequest), body)
			Expect(body).To(ContainSubstring("Invalid percentage provided"))
		})

		It("Should fail if percentage greater than 100", func() {
			leaderboardID := uuid.NewV4().String()
			status, body := Get(a, fmt.Sprintf("/l/%s/top-percent/120", leaderboardID))
			Expect(status).To(Equal(http.StatusBadRequest), body)
			Expect(body).To(ContainSubstring("Percentage must be a valid integer between 1 and 100."))
		})

		It("Should fail if percentage lesser than 1", func() {
			leaderboardID := uuid.NewV4().String()
			status, body := Get(a, fmt.Sprintf("/l/%s/top-percent/0", leaderboardID))
			Expect(status).To(Equal(http.StatusBadRequest), body)
			Expect(body).To(ContainSubstring("Percentage must be a valid integer between 1 and 100."))
		})

		It("Should fail if error in Redis", func() {
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := Get(app, "/l/testkey/top-percent/10")
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		HTTPMeasure("it should get top percentage of members", func(ctx map[string]interface{}) {
			lead := uuid.NewV4().String()

			for i := 0; i < 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), lead, fmt.Sprintf("member-%d", i), 500, false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			ctx["lead"] = lead
		}, func(ts *httptest.Server, ctx map[string]interface{}) {
			lead := ctx["lead"].(string)
			url := getRoute(ts, fmt.Sprintf("/l/%s/top-percent/10", lead))
			status, body, err := fastGet(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.05)
	})

	Describe("Get member score in many leaderboads", func() {
		It("Should get member score in many leaderboards", func() {
			payload := map[string]interface{}{
				"score":        100,
				"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
			}
			status, body := PutJSON(a, "/m/memberpublicid/scores", payload)
			Expect(status).To(Equal(http.StatusOK), body)
			status, body = Get(a, "/m/memberpublicid/scores?leaderboardIds=testkey1,testkey2,testkey3,testkey4,testkey5")
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
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
			status, body := PutJSON(a, "/m/memberpublicid/scores", payload)
			Expect(status).To(Equal(http.StatusOK), body)
			status, body = Get(a, "/m/memberpublicid/scores?leaderboardIds=testkey1,testkey2,testkey3,testkey4,testkey6")
			Expect(status).To(Equal(http.StatusNotFound), body)
		})

	})

	Describe("Upsert Member Score For Several Leaderboards", func() {
		It("Should set correct member score in redis and respond with the correct values", func() {
			payload := map[string]interface{}{
				"score":        100,
				"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
			}
			status, body := PutJSON(a, "/m/memberpublicid/scores?prevRank=true", payload)
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
			scores := result["scores"].([]interface{})
			Expect(len(scores)).To(Equal(5))
			for i, scoreObj := range scores {
				score := scoreObj.(map[string]interface{})
				Expect(score["publicID"]).To(Equal("memberpublicid"))
				Expect(int(score["score"].(float64))).To(Equal(payload["score"]))
				Expect(int(score["rank"].(float64))).To(Equal(1))
				Expect(int(score["previousRank"].(float64))).To(Equal(-1))
				Expect(score["leaderboardID"]).To(Equal(payload["leaderboards"].([]string)[i]))

				member, err := a.Leaderboards.GetMember(NewEmptyCtx(), score["leaderboardID"].(string),
					"memberpublicid", "desc", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(member.Rank).To(Equal(1))
				Expect(member.Score).To(Equal(int64(100)))
				Expect(member.PublicID).To(Equal("memberpublicid"))
			}
		})

		It("Should fail if missing score", func() {
			payload := map[string]interface{}{
				"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
			}
			status, body := PutJSON(a, "/m/memberpublicid/scores", payload)
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("score is required"))
		})

		It("Should fail if missing leaderboards", func() {
			payload := map[string]interface{}{
				"score": int64(100),
			}
			status, body := PutJSON(a, "/m/memberpublicid/scores", payload)
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("leaderboards is required"))
		})

		It("Should fail if invalid payload", func() {
			status, body := Put(a, "/m/memberpublicid/scores", "invalid")
			Expect(status).To(Equal(http.StatusBadRequest), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(ContainSubstring("score is required"))
		})

		It("Should fail if error in Redis when upserting many leaderboards", func() {
			app := GetDefaultTestAppWithFaultyRedis()

			payload := map[string]interface{}{
				"score":        100,
				"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
			}
			status, body := PutJSON(app, "/m/memberpublicid/scores", payload)
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		HTTPMeasure("it should set correct member score for all leaderboards", func(ctx map[string]interface{}) {
			payload := map[string]interface{}{
				"score":        100,
				"leaderboards": []string{"testkey1", "testkey2", "testkey3", "testkey4", "testkey5"},
			}
			payloadJSON, err := json.Marshal(payload)
			Expect(err).NotTo(HaveOccurred())

			ctx["payload"] = payloadJSON
		}, func(ts *httptest.Server, ctx map[string]interface{}) {
			url := getRoute(ts, "/m/memberpublicid/scores")
			status, body, err := fastPutTo(url, ctx["payload"].([]byte))
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.05)
	})

	Describe("Remove Leaderboard", func() {
		It("should remove a leaderboard", func() {
			leaderboardID := uuid.NewV4().String()

			for i := 0; i < 10; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("member-%d", i), 500, false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Delete(a, fmt.Sprintf("/l/%s", leaderboardID))
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
		})

		It("should remove a leaderboard that does not exist", func() {
			status, body := Delete(a, fmt.Sprintf("/l/%s", uuid.NewV4().String()))
			Expect(status).To(Equal(http.StatusOK), body)
			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)
			Expect(result["success"]).To(BeTrue())
		})

		It("Should fail if error in Redis", func() {
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := Delete(app, fmt.Sprintf("/l/%s", uuid.NewV4().String()))
			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})
	})

	Describe("Get Members Handler", func() {
		It("should get several members from leaderboard", func() {
			leaderboardID := uuid.NewV4().String()

			for i := 1; i <= 100; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("member_%d", i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(
				a,
				fmt.Sprintf("/l/%s/members?ids=member_10,member_20,member_30", leaderboardID),
			)
			Expect(status).To(Equal(http.StatusOK), body)

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

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
			status, body := Get(
				a,
				"/l/invalid-leaderboard/members/?ids=member_10,member_20,member_30",
			)
			Expect(status).To(Equal(http.StatusOK), body)

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

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

			for i := 1; i <= 10; i++ {
				_, err := a.Leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("member_%d", i), int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			status, body := Get(
				a,
				fmt.Sprintf("/l/%s/members?ids=member_1,invalid_member", leaderboardID),
			)
			Expect(status).To(Equal(http.StatusOK), body)

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

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

			status, body := Get(a, fmt.Sprintf("/l/%s/members/", leaderboardID))
			Expect(status).To(Equal(http.StatusBadRequest), body)

			var result map[string]interface{}
			json.Unmarshal([]byte(body), &result)

			Expect(result["success"]).To(BeFalse())
			Expect(result["reason"]).To(Equal("Member IDs are required using the 'ids' querystring parameter"))
		})

		It("Should fail if error in Redis", func() {
			app := GetDefaultTestAppWithFaultyRedis()

			status, body := Get(
				app,
				"/l/invalid-redis/members?ids=member_10,member_20,member_30",
			)

			Expect(status).To(Equal(500), body)
			Expect(body).To(ContainSubstring("connection refused"))
		})

		HTTPMeasure("it should get members", func(ctx map[string]interface{}) {
			app := ctx["app"].(*api.App)

			memberIDs := []string{}
			for i := 1; i <= 1000; i++ {
				memberID := fmt.Sprintf("member_%d", i)
				memberIDs = append(memberIDs, memberID)
				_, err := app.Leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, memberID, int64(101-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			ctx["mIDs"] = strings.Join(memberIDs, ",")
		}, func(ts *httptest.Server, ctx map[string]interface{}) {
			url := getRoute(ts, fmt.Sprintf("/l/%s/members?ids=%s", testLeaderboardID, ctx["mIDs"].(string)))
			status, body, err := fastGet(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.1)
	})
})
