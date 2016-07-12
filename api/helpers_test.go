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

	"github.com/gavv/httpexpect"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// GetDefaultTestApp returns a new arkadiko API Application bound to 0.0.0.0:8888 for test
func GetDefaultTestApp() *App {
	app := GetApp("0.0.0.0", 8890, "../config/test.yaml", true)
	app.Configure()
	return app
}

// Get returns a test request against specified URL
func Get(app *App, url string) *httpexpect.Response {
	req := sendRequest(app, "GET", url)
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
	handler := app.App.NoListen().Handler

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL: "http://example.com",
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(handler),
			Jar:       httpexpect.NewJar(),
		},
		Reporter: &GinkgoReporter{},
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(&GinkgoPrinter{}, true),
		},
	})

	return e.Request(method, url)
}
