package testing

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/api"
	"github.com/topfreegames/podium/leaderboard"
	"github.com/topfreegames/podium/leaderboard/database/redis"
	"github.com/topfreegames/podium/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var serverInitialized map[string]bool = map[string]bool{}
var defaultApp *api.App
var defaultFaultyRedisApp *api.App

// GetDefaultTestApp returns a testing app
func GetDefaultTestApp() *api.App {
	if defaultApp != nil {
		return defaultApp
	}

	logger := log.CreateLoggerWithLevel(zap.FatalLevel, log.LoggerOptions{WriteSyncer: os.Stdout, RemoveTimestamp: true})
	app, err := api.New("127.0.0.1", 0, 0, "../config/test.yaml", false, logger)
	if err != nil {
		panic(fmt.Sprintf("Could not get app: %s\n", err.Error()))
	}
	defaultApp = app
	return app
}

// ShutdownDefaultTestApp turn off default test app
func ShutdownDefaultTestApp() {
	if defaultApp != nil {
		defaultApp.GracefullStop()
	}
}

// InitializeTestServer starts the test server
func InitializeTestServer(app *api.App) {
	if client == nil {
		transport = &http.Transport{DisableKeepAlives: true}
		client = &http.Client{Transport: transport}
	}

	if !serverInitialized[app.HTTPEndpoint] {
		go func() {
			err := app.Start(context.Background())
			fmt.Printf("Starting test server. HTTP: %s; GRPC: %s\n", app.HTTPEndpoint, app.GRPCEndpoint)
			if err != nil {
				panic(err)
			}
		}()
		serverInitialized[app.HTTPEndpoint] = true
		err := app.WaitForReady(500 * time.Millisecond)
		Expect(err).NotTo(HaveOccurred())
	}
}

// GetDefaultTestAppWithFaultyRedis returns a new podium API Application bound to 0.0.0.0:8890 for test but with a failing Redis
func GetDefaultTestAppWithFaultyRedis() *api.App {
	if defaultFaultyRedisApp != nil {
		return defaultFaultyRedisApp
	}

	logger := log.CreateLoggerWithLevel(zapcore.DebugLevel, log.LoggerOptions{WriteSyncer: os.Stdout, RemoveTimestamp: true})
	app, err := api.New("127.0.0.1", 8082, 8083, "../config/test.yaml", false, logger)
	Expect(err).NotTo(HaveOccurred())

	leaderboard := leaderboard.NewClient(
		app.Config.GetString("faultyRedis.host"),
		app.Config.GetInt("faultyRedis.port"),
		app.Config.GetString("faultyRedis.password"),
		app.Config.GetInt("faultyRedis.db"),
	)
	app.Leaderboards = leaderboard

	defaultFaultyRedisApp = app
	return app
}

// ShutdownDefaultTestAppWithFaltyRedis turn off default test app
func ShutdownDefaultTestAppWithFaltyRedis() {
	if defaultFaultyRedisApp != nil {
		defaultFaultyRedisApp.GracefullStop()
	}
}

// GetAppRedis creates a redis instance based on the app configuration
func GetAppRedis(app *api.App) redis.Redis {
	shouldRunOnCluster := app.Config.GetBool("redis.cluster.enabled")
	password := app.Config.GetString("redis.password")
	if shouldRunOnCluster {
		addrs := app.Config.GetStringSlice("redis.addrs")
		return redis.NewClusterClient(redis.ClusterOptions{
			Addrs:    addrs,
			Password: password,
		})
	}

	host := app.Config.GetString("redis.host")
	port := app.Config.GetInt("redis.port")
	db := app.Config.GetInt("redis.db")

	return redis.NewStandaloneClient(redis.StandaloneOptions{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	})
}
