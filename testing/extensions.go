// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package testing

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/engine/standard"
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	"github.com/onsi/gomega"
	"github.com/topfreegames/podium/api"
	"go.uber.org/zap"
)

//BeforeOnce runs the before each block only once
func BeforeOnce(beforeBlock func()) {
	hasRun := false

	ginkgo.BeforeEach(func() {
		if !hasRun {
			beforeBlock()
			hasRun = true
		}
	})
}

var client *http.Client
var transport *http.Transport

func initClient() {
	if client == nil {
		transport = &http.Transport{DisableKeepAlives: true}
		client = &http.Client{Transport: transport}
	}
}
func initializeTestServer(app *api.App) *httptest.Server {
	initClient()
	app.Engine.SetHandler(app.App)
	return httptest.NewServer(app.Engine.(*standard.Server))
}

// GetDefaultTestApp returns a new podium API Application bound to 0.0.0.0:8890 for test
func getDefaultTestApp() *api.App {
	logger := zap.New(
		zap.NewJSONEncoder(),
		zap.FatalLevel,
	)
	app, err := api.GetApp("0.0.0.0", 8890, "../config/test.yaml", false, false, logger)
	if err != nil {
		panic(fmt.Sprintf("Could not get app: %s\n", err.Error()))
	}
	app.Configure()
	return app
}

//HTTPMeasure runs the specified specs in an http test
func HTTPMeasure(description string, setup func(map[string]interface{}), f func(*httptest.Server, map[string]interface{}), timeout float64) bool {
	return measure(description, setup, f, timeout, types.FlagTypeNone)
}

//FHTTPMeasure runs the specified specs in an http test
func FHTTPMeasure(description string, setup func(map[string]interface{}), f func(*httptest.Server, map[string]interface{}), timeout float64) bool {
	return measure(description, setup, f, timeout, types.FlagTypeFocused)
}

//XHTTPMeasure runs the specified specs in an http test
func XHTTPMeasure(description string, setup func(map[string]interface{}), f func(*httptest.Server, map[string]interface{}), timeout float64) bool {
	return measure(description, setup, f, timeout, types.FlagTypePending)
}

func measure(description string, setup func(map[string]interface{}), f func(*httptest.Server, map[string]interface{}), timeout float64, flagType types.FlagType) bool {
	app := getDefaultTestApp()
	d := func(t string, f func()) { ginkgo.Describe(t, f) }
	if flagType == types.FlagTypeFocused {
		d = func(t string, f func()) { ginkgo.FDescribe(t, f) }
	}
	if flagType == types.FlagTypePending {
		d = func(t string, f func()) { ginkgo.XDescribe(t, f) }
	}

	d("Measure", func() {
		var ts *httptest.Server
		var loops int
		var ctx map[string]interface{}

		BeforeOnce(func() {
			ts = initializeTestServer(app)
			ctx = map[string]interface{}{"app": app}
			setup(ctx)
		})

		ginkgo.AfterEach(func() {
			loops++
			if loops == 200 {
				transport.CloseIdleConnections()
				ts.Close()
			}
		})

		ginkgo.Measure(description, func(b ginkgo.Benchmarker) {
			runtime := b.Time("runtime", func() {
				f(ts, ctx)
			})
			gomega.Expect(runtime.Seconds()).Should(
				gomega.BeNumerically("<", timeout),
				fmt.Sprintf("%s shouldn't take too long.", description),
			)
		}, 200)
	})

	return true
}
