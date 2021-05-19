package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/api"
	"github.com/valyala/fasthttp"
)

var client *http.Client
var transport *http.Transport

// Get from server
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

func doRequest(app *api.App, method, url, body string) (int, string) {
	InitializeTestServer(app)
	req := getRequest(app, method, url, body)
	return performRequest(req)
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

// GetRoute returns the endpoint for accessing an url in an app
func GetRoute(httpEndPoint string, url string) string {
	return fmt.Sprintf("http://%s%s", httpEndPoint, url)
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

// FastGet performs a GET on an URL
func FastGet(url string) (int, []byte, error) {
	return fastSendTo("GET", url, nil)
}

// FastDelete performs a DELETE on an URL
func FastDelete(url string) (int, []byte, error) {
	return fastSendTo("DELETE", url, nil)
}

// FastPostTo performs a POST on an URL
func FastPostTo(url string, payload []byte) (int, []byte, error) {
	return fastSendTo("POST", url, payload)
}

// FastPutTo performs a PUT on an URL
func FastPutTo(url string, payload []byte) (int, []byte, error) {
	return fastSendTo("PUT", url, payload)
}

// FastPatchTo performs a PATCH on an URL
func FastPatchTo(url string, payload []byte) (int, []byte, error) {
	return fastSendTo("PATCH", url, payload)
}
