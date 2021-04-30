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
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/topfreegames/podium/api"
	"github.com/topfreegames/podium/leaderboard"
	"github.com/topfreegames/podium/log"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"

	extredis "github.com/topfreegames/extensions/redis"
	pb "github.com/topfreegames/podium/proto/podium/api/v1"
)

var serverInitialized map[string]bool = map[string]bool{}
var defaultApp *api.App
var defaultFaultyRedisApp *api.App

func GetConnectedRedis(app *api.App) (*extredis.Client, error) {
	redisURL := url.URL{
		Scheme: "redis",
		User:   url.UserPassword("", app.Config.GetString("redis.password")),
		Host:   fmt.Sprintf("%s:%d", app.Config.GetString("redis.host"), app.Config.GetInt("redis.port")),
		Path:   app.Config.GetString("redis.db"),
	}
	app.Config.Set("redis.url", redisURL.String())
	app.Config.Set("redis.connectionTimeout", 200)

	return extredis.NewClient("redis", app.Config)
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

// GetDefaultTestApp returns a new podium API Application bound to random ports for test
func GetDefaultTestApp() *api.App {
	if defaultApp != nil {
		return defaultApp
	}
	logger := log.CreateLoggerWithLevel(zapcore.DebugLevel, log.LoggerOptions{WriteSyncer: os.Stdout, RemoveTimestamp: true})
	app, err := api.New("127.0.0.1", 8080, 8081, "../config/test.yaml", false, logger)
	Expect(err).NotTo(HaveOccurred())

	defaultApp = app
	return app
}

// GetFaultyTestApp returns a new podium API Application bound to 0.0.0.0:8890 for test but with a failing Redis
func GetDefaultTestAppWithFaultyRedis() *api.App {
	if defaultFaultyRedisApp != nil {
		return defaultFaultyRedisApp
	}

	logger := log.CreateLoggerWithLevel(zapcore.DebugLevel, log.LoggerOptions{WriteSyncer: os.Stdout, RemoveTimestamp: true})
	app, err := api.New("127.0.0.1", 8082, 8083, "../config/test.yaml", false, logger)
	Expect(err).NotTo(HaveOccurred())

	faultyRedisClient, err := GetConnectedRedis(app)
	Expect(err).NotTo(HaveOccurred())
	faultyRedisClient.Client = GetFaultyRedis()
	leaderboardClient, err := leaderboard.NewClientWithRedis(faultyRedisClient)
	Expect(err).NotTo(HaveOccurred())

	app.Leaderboards = leaderboardClient
	defaultFaultyRedisApp = app
	return app
}

// ShutdownDefaultTestApp turn off default test app
func ShutdownDefaultTestApp() {
	defaultApp.GracefullStop()
}

// ShutdownDefaultTestWithFaultyApp turn off default test app
func ShutdownDefaultTestAppWithFaltyRedis() {
	defaultFaultyRedisApp.GracefullStop()
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
	if !serverInitialized[app.HTTPEndpoint] {
		go func() {
			_ = app.Start(context.Background())
		}()
		serverInitialized[app.HTTPEndpoint] = true
		err := app.WaitForReady(500 * time.Millisecond)
		Expect(err).NotTo(HaveOccurred())
	}
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
	req := getRequest(app, method, url, body)
	return performRequest(req)
}

func getRoute(httpEndPoint string, url string) string {
	return fmt.Sprintf("http://%s%s", httpEndPoint, url)
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
	if fastClient != nil {
		return fastClient
	}

	fastClient = &fasthttp.Client{}
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
func SetupGRPC(app *api.App, f func(pb.PodiumClient)) {
	initializeTestServer(app)

	conn, err := grpc.Dial(app.GRPCEndpoint, grpc.WithInsecure())
	Expect(err).NotTo(HaveOccurred())
	defer func() {
		_ = conn.Close()
	}()

	cli := pb.NewPodiumClient(conn)

	f(cli)
}

//TestBuffer is a mock buffer
type TestBuffer struct {
	bytes.Buffer
}

//Sync does nothing
func (b *TestBuffer) Sync() error {
	return nil
}

//Lines returns all lines of log
func (b *TestBuffer) Lines() []string {
	output := strings.Split(b.String(), "\n")
	return output[:len(output)-1]
}

//Stripped removes new lines
func (b *TestBuffer) Stripped() string {
	return strings.TrimRight(b.String(), "\n")
}

//ResetStdout back to os.Stdout
var ResetStdout func()

//ReadStdout value
var ReadStdout func() string

//MockStdout to read it's value later
func MockStdout() {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	os.Stdout = w

	ReadStdout = func() string {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		r.Close()
		return buf.String()
	}

	ResetStdout = func() {
		w.Close()
		os.Stdout = stdout
	}
}
