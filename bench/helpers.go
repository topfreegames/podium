// podium
// https://github.com/topfreegames/podium
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package bench

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/gomega"
	"github.com/topfreegames/extensions/redis"
	"github.com/topfreegames/podium/config"
	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/service"
)

func getRedis() *redis.Client {
	config, err := config.GetDefaultConfig("../config/default.yaml")
	Expect(err).NotTo(HaveOccurred())

	redisHost := config.GetString("redis.host")
	redisPort := config.GetInt("redis.port")
	redisDB := config.GetInt("redis.db")

	redisURL := fmt.Sprintf("redis://%s:%d/%d", redisHost, redisPort, redisDB)

	config.SetDefault("redis.url", redisURL)
	config.SetDefault("redis.connectionTimeout", 200)

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
	if err != nil {
		return -1, "", err
	}
	defer resp.Body.Close()

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

func generateNMembers(amount int) string {
	config, err := config.GetDefaultConfig("../config/default.yaml")
	Expect(err).NotTo(HaveOccurred())

	client := service.NewService(
		database.NewRedisDatabase(database.RedisOptions{
			ClusterEnabled: config.GetBool("redis.clusterEnabled"),
			Addrs:          config.GetStringSlice("redis.addrs"),
			Host:           config.GetString("redis.host"),
			Port:           config.GetInt("redis.port"),
			Password:       config.GetString("redis.password"),
			DB:             config.GetInt("redis.db"),
		}),
	)

	lbID := "leaderboard-0"

	for i := 0; i < amount; i++ {
		client.SetMemberScore(context.Background(), lbID, fmt.Sprintf("bench-member-%d", i), int64(100+i), false, "inf")
	}

	return lbID
}
