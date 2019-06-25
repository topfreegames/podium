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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"google.golang.org/grpc"

	"github.com/spf13/viper"
	"github.com/topfreegames/podium/leaderboard"

	"github.com/go-redis/redis"

	. "github.com/onsi/gomega"
	extredis "github.com/topfreegames/extensions/redis"
	"github.com/topfreegames/podium/api"
	"github.com/topfreegames/podium/testing"
	"github.com/valyala/fasthttp"

	pb "github.com/topfreegames/podium/proto/podium/api/v1"
)

func GetConnectedRedis() (*extredis.Client, error) {
	config := viper.New()
	config.Set("redis.url", "redis://localhost:1234/0")
	config.Set("redis.connectionTimeout", 200)

	return extredis.NewClient("redis", config)
}

//GetFaultyRedis returns an invalid connection to redis
func GetFaultyRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:38465",
		Password: "",
		DB:       0,
		PoolSize: 20,
	})
}

//Creates an empty context (shortcut for context.Background())
func NewEmptyCtx() context.Context {
	return context.Background()
}

// GetDefaultTestApp returns a new podium API Application bound to 0.0.0.0:8890 for test
func GetDefaultTestApp() *api.App {
	logger := testing.NewMockLogger()
	app, err := api.GetApp("0.0.0.0", 8890, 8900, "../config/test.yaml", false, false, logger)
	Expect(err).NotTo(HaveOccurred())
	return app
}

// GetFaultyTestApp returns a new podium API Application bound to 0.0.0.0:8890 for test but with a failing Redis
func GetDefaultTestAppWithFaultyRedis() *api.App {
	app := GetDefaultTestApp()
	faultyRedisClient, err := GetConnectedRedis()
	Expect(err).NotTo(HaveOccurred())
	faultyRedisClient.Client = GetFaultyRedis()
	app.Leaderboards = leaderboard.NewClientWithRedis(faultyRedisClient)
	return app
}

//Get from server
func Get(app *api.App, url string) (int, string) {
	return doRequest(app, "GET", url, "")
}

//Post to server
func Post(app *api.App, url, body string) (int, string) {
	return doRequest(app, "POST", url, body)
}

//PostJSON to server
func PostJSON(app *api.App, url string, body interface{}) (int, string) {
	result, err := json.Marshal(body)
	if err != nil {
		return 510, "Failed to marshal specified body to JSON format"
	}
	return Post(app, url, string(result))
}

//Put to server
func Put(app *api.App, url, body string) (int, string) {
	return doRequest(app, "PUT", url, body)
}

//PutJSON to server
func PutJSON(app *api.App, url string, body interface{}) (int, string) {
	result, err := json.Marshal(body)
	if err != nil {
		return 510, "Failed to marshal specified body to JSON format"
	}
	return Put(app, url, string(result))
}

//Patch to server
func Patch(app *api.App, url, body string) (int, string) {
	return doRequest(app, "PATCH", url, body)
}

//PatchJSON to server
func PatchJSON(app *api.App, url string, body interface{}) (int, string) {
	result, err := json.Marshal(body)
	if err != nil {
		return 510, "Failed to marshal specified body to JSON format"
	}
	return Patch(app, url, string(result))
}

//Delete from server
func Delete(app *api.App, url string) (int, string) {
	return doRequest(app, "DELETE", url, "")
}

var client *http.Client
var transport *http.Transport

func initializeTestServer(app *api.App) {
	if client == nil {
		transport = &http.Transport{DisableKeepAlives: true}
		client = &http.Client{Transport: transport}
	}
	go func() {
		_ = app.Start()
	}()
	//wait for server to start
	time.Sleep(25 * time.Millisecond)
}

func shutdownTestServer(app *api.App) {
	app.Stop()
}

func getRequest(app *api.App, method, url, body string) *http.Request {
	var bodyBuff io.Reader
	if body != "" {
		bodyBuff = bytes.NewBuffer([]byte(body))
	}
	req, err := http.NewRequest(method, fmt.Sprintf("http://%s%s", app.HTTPEndpoint, url), bodyBuff)
	req.Header.Set("Connection", "close")
	req.Close = true
	Expect(err).NotTo(HaveOccurred())

	return req
}

func performRequest(req *http.Request) (int, string) {
	res, err := client.Do(req)
	Expect(err).NotTo(HaveOccurred())

	b, err := ioutil.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())

	err = res.Body.Close()
	Expect(err).NotTo(HaveOccurred())

	return res.StatusCode, string(b)
}

func doRequest(app *api.App, method, url, body string) (int, string) {
	initializeTestServer(app)
	defer shutdownTestServer(app)
	req := getRequest(app, method, url, body)
	return performRequest(req)
}

func getRoute(ts *httptest.Server, url string) string {
	return fmt.Sprintf("%s%s", ts.URL, url)
}

func fastGet(url string) (int, []byte, error) {
	return fastSendTo("GET", url, nil)
}

func fastDelete(url string) (int, []byte, error) {
	return fastSendTo("DELETE", url, nil)
}

func fastPostTo(url string, payload []byte) (int, []byte, error) {
	return fastSendTo("POST", url, payload)
}

func fastPutTo(url string, payload []byte) (int, []byte, error) {
	return fastSendTo("PUT", url, payload)
}

func fastPatchTo(url string, payload []byte) (int, []byte, error) {
	return fastSendTo("PATCH", url, payload)
}

var fastClient *fasthttp.Client

func fastGetClient() *fasthttp.Client {
	if fastClient == nil {
		fastClient = &fasthttp.Client{}
	}
	return fastClient
}

func fastSendTo(method, url string, payload []byte) (int, []byte, error) {
	c := fastGetClient()
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	if payload != nil {
		req.AppendBody(payload)
	}
	resp := fasthttp.AcquireResponse()
	err := c.Do(req, resp)
	return resp.StatusCode(), resp.Body(), err
}

//sets up the environment for grpc communication, starting the app and creating a connected client
func SetupGRPC(app *api.App, f func(pb.PodiumAPIClient)) {
	go func() {
		_ = app.Start()
	}()
	time.Sleep(25 * time.Millisecond)

	conn, err := grpc.Dial(app.GRPCEndpoint, grpc.WithInsecure())
	Expect(err).NotTo(HaveOccurred())
	defer conn.Close()

	cli := pb.NewPodiumAPIClient(conn)

	f(cli)

	app.Stop()
}
