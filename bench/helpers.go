// podium
// https://github.com/topfreegames/podium
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/redis"
	"github.com/topfreegames/podium/leaderboard"
	test "github.com/topfreegames/podium/testing"
)

func getRedis() *redis.Client {
	config := viper.New()
	config.Set("redis.url", "redis://localhost:1224/0")
	config.Set("redis.connectionTimeout", 200)

	redisClient, err := redis.NewClient("redis", config)
	if err != nil {
		panic(err.Error())
	}

	return redisClient
}

func getRoute(url string) string {
	return fmt.Sprintf("http://localhost:8888%s", url)
}

func get(url string) (int, string, error) {
	return sendTo("GET", url, nil)
}

func postTo(url string, payload map[string]interface{}) (int, string, error) {
	return sendTo("POST", url, payload)
}

func putTo(url string, payload map[string]interface{}) (int, string, error) {
	return sendTo("PUT", url, payload)
}

func patchTo(url string, payload map[string]interface{}) (int, string, error) {
	return sendTo("PATCH", url, payload)
}

func sendTo(method, url string, payload map[string]interface{}) (int, string, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return -1, "", err
	}

	var req *http.Request

	if payload != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(payloadJSON))
		if err != nil {
			return -1, "", err
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return -1, "", err
		}
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return -1, "", err
	}

	body, _ := ioutil.ReadAll(resp.Body)

	return resp.StatusCode, string(body), nil
}

func validateResp(statusCode int, body string, err error) {
	if err != nil {
		panic(err)
	}
	if statusCode != 200 {
		fmt.Printf("Request failed with status code %d\n", statusCode)
		panic(body)
	}
}

func generateNMembers(amount int) *leaderboard.Leaderboard {
	logger := test.NewMockLogger()
	redisClient := getRedis()

	lbID := "leaderboard-0"

	l := leaderboard.NewLeaderboard(redisClient.Client, lbID, 10, logger)

	for i := 0; i < amount; i++ {
		l.SetMemberScore(fmt.Sprintf("bench-member-%d", i), 100+i, false, "inf")
	}

	return l
}
