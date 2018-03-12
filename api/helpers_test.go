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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-redis/redis"

	"github.com/labstack/echo/engine/standard"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/api"
	"github.com/topfreegames/podium/testing"
	"github.com/valyala/fasthttp"
)

//GetFaultyRedis returns an invalid connection to redis
func GetFaultyRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:38465",
		Password: "",
		DB:       0,
		PoolSize: 20,
	})
}

// GetDefaultTestApp returns a new podium API Application bound to 0.0.0.0:8890 for test
func GetDefaultTestApp() *api.App {
	logger := testing.NewMockLogger()
	app, err := api.GetApp("0.0.0.0", 8890, "../config/test.yaml", false, false, logger)
	Expect(err).NotTo(HaveOccurred())
	app.Configure()
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

func initClient() {
	if client == nil {
		transport = &http.Transport{DisableKeepAlives: true}
		client = &http.Client{Transport: transport}
	}
}

func InitializeTestServer(app *api.App) *httptest.Server {
	initClient()
	app.Engine.SetHandler(app.App)
	return httptest.NewServer(app.Engine.(*standard.Server))
}

func GetRequest(app *api.App, ts *httptest.Server, method, url, body string) *http.Request {
	var bodyBuff io.Reader
	if body != "" {
		bodyBuff = bytes.NewBuffer([]byte(body))
	}
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", ts.URL, url), bodyBuff)
	req.Header.Set("Connection", "close")
	req.Close = true
	Expect(err).NotTo(HaveOccurred())

	return req
}

func PerformRequest(ts *httptest.Server, req *http.Request) (int, string) {
	res, err := client.Do(req)
	//Wait for port of httptest to be reclaimed by OS
	time.Sleep(50 * time.Millisecond)
	Expect(err).NotTo(HaveOccurred())

	b, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	Expect(err).NotTo(HaveOccurred())

	return res.StatusCode, string(b)
}

func doRequest(app *api.App, method, url, body string) (int, string) {
	ts := InitializeTestServer(app)
	defer transport.CloseIdleConnections()
	defer ts.Close()

	req := GetRequest(app, ts, method, url, body)
	return PerformRequest(ts, req)
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
