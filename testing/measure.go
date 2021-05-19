package testing

import (
	"fmt"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

//HTTPMeasure runs the specified specs in an http test
func HTTPMeasure(description string, setup func(map[string]interface{}), f func(string, map[string]interface{}), timeout float64) bool {
	return measure(description, setup, f, timeout, types.FlagTypeNone)
}

//FHTTPMeasure runs the specified specs in an http test
func FHTTPMeasure(description string, setup func(map[string]interface{}), f func(string, map[string]interface{}), timeout float64) bool {
	return measure(description, setup, f, timeout, types.FlagTypeFocused)
}

//XHTTPMeasure runs the specified specs in an http test
func XHTTPMeasure(description string, setup func(map[string]interface{}), f func(string, map[string]interface{}), timeout float64) bool {
	return measure(description, setup, f, timeout, types.FlagTypePending)
}

func measure(description string, setup func(map[string]interface{}), f func(string, map[string]interface{}), timeout float64, flagType types.FlagType) bool {
	app := GetDefaultTestApp()

	d := func(t string, f func()) { ginkgo.Describe(t, f) }
	if flagType == types.FlagTypeFocused {
		d = func(t string, f func()) { ginkgo.FDescribe(t, f) }
	}
	if flagType == types.FlagTypePending {
		d = func(t string, f func()) { ginkgo.XDescribe(t, f) }
	}

	d("Measure", func() {
		var loops int
		var ctx map[string]interface{}

		BeforeOnce(func() {
			InitializeTestServer(app)
			ctx = map[string]interface{}{"app": app}
			setup(ctx)
		})

		ginkgo.AfterEach(func() {
			loops++
			if loops == 200 {
				transport.CloseIdleConnections()
			}
		})

		ginkgo.Measure(description, func(b ginkgo.Benchmarker) {
			runtime := b.Time("runtime", func() {
				f(app.HTTPEndpoint, ctx)
			})
			Expect(runtime.Seconds()).Should(
				BeNumerically("<", timeout),
				fmt.Sprintf("%s shouldn't take too long.", description),
			)
		}, 200)
	})

	return true
}
