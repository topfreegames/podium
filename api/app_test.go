package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/api"
	"github.com/topfreegames/podium/testing"
)

var _ = Describe("App", func() {
	var logger *testing.MockLogger
	BeforeEach(func() {
		logger = testing.NewMockLogger()
	})

	Describe("App creation", func() {
		It("should create new app", func() {
			app := api.GetApp("127.0.0.1", 9999, "../config/test.yaml", true, logger)
			Expect(app).NotTo(BeNil())
			Expect(app.Host).To(Equal("127.0.0.1"))
			Expect(app.Port).To(Equal(9999))
			Expect(app.Debug).To(BeTrue())
			Expect(app.Config).NotTo(BeNil())
		})
	})

	Describe("App Load Configuration", func() {
		It("Should load configuration from file", func() {
			app := api.GetApp("127.0.0.1", 9999, "../config/test.yaml", false, logger)
			Expect(app.Config).NotTo(BeNil())
			expected := app.Config.GetString("healthcheck.workingText")
			Expect(expected).To(Equal("WORKING"))
		})
	})
})
