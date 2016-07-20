// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gavv/httpexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/valyala/fasthttp"
)

// GetDefaultTestApp returns a new podium API Application bound to 0.0.0.0:8890 for test
func GetDefaultTestApp() *App {
	app := GetApp("0.0.0.0", 8890, "../config/test.yaml", false)
	app.Configure()
	return app
}

// Get returns a test request against specified URL
func Get(app *App, url string, queryString ...map[string]interface{}) *httpexpect.Response {
	req := sendRequest(app, "GET", url)
	if len(queryString) == 1 {
		for k, v := range queryString[0] {
			req = req.WithQuery(k, v)
		}
	}
	return req.Expect()
}

// PostBody returns a test request against specified URL
func PostBody(app *App, url string, payload string) *httpexpect.Response {
	return sendBody(app, "POST", url, payload)
}

// PutBody returns a test request against specified URL
func PutBody(app *App, url string, payload string) *httpexpect.Response {
	return sendBody(app, "PUT", url, payload)
}

func sendBody(app *App, method string, url string, payload string) *httpexpect.Response {
	req := sendRequest(app, method, url)
	return req.WithBytes([]byte(payload)).Expect()
}

// PostJSON returns a test request against specified URL
func PostJSON(app *App, url string, payload map[string]interface{}) *httpexpect.Response {
	return sendJSON(app, "POST", url, payload)
}

// PutJSON returns a test request against specified URL
func PutJSON(app *App, url string, payload map[string]interface{}) *httpexpect.Response {
	return sendJSON(app, "PUT", url, payload)
}

func sendJSON(app *App, method, url string, payload map[string]interface{}) *httpexpect.Response {
	req := sendRequest(app, method, url)
	return req.WithJSON(payload).Expect()
}

// Delete returns a test request against specified URL
func Delete(app *App, url string) *httpexpect.Response {
	req := sendRequest(app, "DELETE", url)
	return req.Expect()
}

//GinkgoReporter implements tests for httpexpect
type GinkgoReporter struct {
}

// Errorf implements Reporter.Errorf.
func (g *GinkgoReporter) Errorf(message string, args ...interface{}) {
	Expect(false).To(BeTrue(), fmt.Sprintf(message, args...))
}

//GinkgoPrinter reports errors to stdout
type GinkgoPrinter struct{}

//Logf reports to stdout
func (g *GinkgoPrinter) Logf(source string, args ...interface{}) {
	fmt.Printf(source, args...)
}

func sendRequest(app *App, method, url string) *httpexpect.Request {
	api := app.App
	srv := api.Servers.Main()

	if srv == nil { // maybe the user called this after .Listen/ListenTLS/ListenUNIX, the tester can be used as standalone (with no running iris instance) or inside a running instance/app
		srv = api.ListenVirtual(api.Config.Tester.ListeningAddr)
	}

	opened := api.Servers.GetAllOpened()
	h := srv.Handler
	baseURL := srv.FullHost()
	if len(opened) > 1 {
		baseURL = ""
		//we have more than one server, so we will create a handler here and redirect by registered listening addresses
		h = func(reqCtx *fasthttp.RequestCtx) {
			for _, s := range opened {
				if strings.HasPrefix(reqCtx.URI().String(), s.FullHost()) { // yes on :80 should be passed :80 also, this is inneed for multiserver testing
					s.Handler(reqCtx)
					break
				}
			}
		}
	}

	if api.Config.Tester.ExplicitURL {
		baseURL = ""
	}

	testConfiguration := httpexpect.Config{
		BaseURL: baseURL,
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(h),
			Jar:       httpexpect.NewJar(),
		},
		Reporter: &GinkgoReporter{},
	}
	if api.Config.Tester.Debug {
		testConfiguration.Printers = []httpexpect.Printer{
			httpexpect.NewDebugPrinter(&GinkgoPrinter{}, true),
		}
	}

	return httpexpect.WithConfig(testConfiguration).Request(method, url)
}
