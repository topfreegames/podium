package api_test

import (
	"encoding/json"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/api"
	"github.com/topfreegames/podium/log"
	"github.com/topfreegames/podium/testing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ = Describe("App", func() {
	var logger *zap.Logger
	var app *api.App
	var err error
	BeforeEach(func() {
		logger = log.CreateLoggerWithLevel(zapcore.FatalLevel, log.LoggerOptions{WriteSyncer: os.Stdout, RemoveTimestamp: true})
	})

	AfterSuite(func() {
		testing.ShutdownDefaultTestApp()
		testing.ShutdownDefaultTestAppWithFaltyRedis()
	})

	Describe("App creation", func() {
		It("should create new app", func() {
			app, err = api.New("127.0.0.1", 9999, 10000, "../config/test.yaml", true, logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(app).NotTo(BeNil())
			Expect(app.HTTPEndpoint).To(Equal("127.0.0.1:9999"))
			Expect(app.GRPCEndpoint).To(Equal("127.0.0.1:10000"))
			Expect(app.Debug).To(BeTrue())
			Expect(app.Config).NotTo(BeNil())
		})
	})

	Describe("App Load Configuration", func() {
		It("Should load configuration from file", func() {
			app, err = api.New("127.0.0.1", 9999, 10000, "../config/test.yaml", false, logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Config).NotTo(BeNil())
			expected := app.Config.GetString("healthcheck.workingText")
			Expect(expected).To(Equal("WORKING"))
		})

		It("Should fail if configuration file does not exist", func() {
			app, err = api.New("127.0.0.1", 9999, 10000, "../config/invalid.yaml", false, logger)
			Expect(app).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not load configuration file from: ../config/invalid.yaml"))
		})
	})

	Describe("Error Handler", func() {
		var sink *testing.TestBuffer
		BeforeEach(func() {
			sink = &testing.TestBuffer{}
			logger = log.CreateLoggerWithLevel(zapcore.ErrorLevel, log.LoggerOptions{WriteSyncer: sink, RemoveTimestamp: true})
		})

		It("should handle errors and send to raven", func() {
			app, err = api.New("127.0.0.1", 9999, 10000, "../config/test.yaml", false, logger)
			Expect(err).NotTo(HaveOccurred())

			app.OnErrorHandler(fmt.Errorf("some other error occurred"), []byte("stack"))
			result := sink.Buffer.String()
			var obj map[string]interface{}
			err = json.Unmarshal([]byte(result), &obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(obj["level"]).To(Equal("error"))
			Expect(obj["panicText"]).To(Equal("some other error occurred"))
			Expect(obj["stack"]).To(Equal("stack"))
		})
	})

	Describe("Error metrics", func() {
		It("should add error rate", func() {
			app, err = api.New("127.0.0.1", 9999, 10000, "../config/test.yaml", false, logger)
			Expect(err).NotTo(HaveOccurred())

			app.AddError()
			app.Errors.Tick()
			Expect(app.Errors.Rate()).To(BeNumerically(">", 0))
		})
	})
})
