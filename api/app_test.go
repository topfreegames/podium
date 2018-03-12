package api_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/api"
	. "github.com/topfreegames/podium/testing"
	"go.uber.org/zap"
)

var _ = Describe("App", func() {
	var logger *MockLogger
	BeforeEach(func() {
		logger = NewMockLogger()
	})

	Describe("App creation", func() {
		It("should create new app", func() {
			app, err := api.GetApp("127.0.0.1", 9999, "../config/test.yaml", true, false, logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(app).NotTo(BeNil())
			Expect(app.Host).To(Equal("127.0.0.1"))
			Expect(app.Port).To(Equal(9999))
			Expect(app.Debug).To(BeTrue())
			Expect(app.Config).NotTo(BeNil())
		})
	})

	Describe("App Load Configuration", func() {
		It("Should load configuration from file", func() {
			app, err := api.GetApp("127.0.0.1", 9999, "../config/test.yaml", false, false, logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Config).NotTo(BeNil())
			expected := app.Config.GetString("healthcheck.workingText")
			Expect(expected).To(Equal("WORKING"))
		})

		It("Should fail if configuration file does not exist", func() {
			app, err := api.GetApp("127.0.0.1", 9999, "../config/invalid.yaml", false, false, logger)
			Expect(app).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not load configuration file from: ../config/invalid.yaml"))
		})
	})

	Describe("App Connect To Redis", func() {
		It("Should faild if invalid redis connection", func() {
			app, err := api.GetApp("127.0.0.1", 9999, "../config/invalid-redis.yaml", false, false, logger)
			Expect(app).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("timed out"))
		})
	})

	Describe("Error Handler", func() {
		It("should handle errors and send to raven", func() {
			app, err := api.GetApp("127.0.0.1", 9999, "../config/test.yaml", false, false, logger)
			Expect(err).NotTo(HaveOccurred())

			app.OnErrorHandler(fmt.Errorf("some other error occurred"), []byte("stack"))
			Expect(logger).To(HaveLogMessage(
				zap.ErrorLevel, "Panic occurred.",
				"panicText", "some other error occurred",
				"stack", "stack",
			))

		})
	})

	Describe("Error metrics", func() {
		It("should add error rate", func() {
			app, err := api.GetApp("127.0.0.1", 9999, "../config/test.yaml", false, false, logger)
			Expect(err).NotTo(HaveOccurred())

			app.AddError()
			app.Errors.Tick()
			Expect(app.Errors.Rate()).To(BeNumerically(">", 0))
		})
	})
})
