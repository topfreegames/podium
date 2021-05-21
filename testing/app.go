package testing

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/api"
	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/database/redis"
	"github.com/topfreegames/podium/leaderboard/v2/service"
	"github.com/topfreegames/podium/leaderboard/v2/testing"
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

	if !serverInitialized[app.ID.String()] {
		serverInitialized[app.ID.String()] = true
		go func() {
			err := app.Start(context.Background())
			if err != nil {
				panic(err)
			}
		}()
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

	leaderboard := service.NewService(database.NewRedisDatabase(database.RedisOptions{
		ClusterEnabled: app.Config.GetBool("faultyRedis.clusterEnabled"),
		Addrs:          app.Config.GetStringSlice("faultyRedis.addrs"),
		Host:           app.Config.GetString("faultyRedis.host"),
		Port:           app.Config.GetInt("faultyRedis.port"),
		Password:       app.Config.GetString("faultyRedis.password"),
		DB:             app.Config.GetInt("faultyRedis.db"),
	}))
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

// GetTestingRedis creates a redis instance based on the default app configuration
func GetTestingRedis(app *api.App) (redis.Client, error) {
	config, err := testing.GetDefaultConfig("../config/test.yaml")
	if err != nil {
		return nil, err
	}
	shouldRunOnCluster := config.GetBool("redis.cluster.enabled")
	password := config.GetString("redis.password")
	if shouldRunOnCluster {
		addrs := config.GetStringSlice("redis.addrs")
		return redis.NewClusterClient(redis.ClusterOptions{
			Addrs:    addrs,
			Password: password,
		}), nil
	}

	host := config.GetString("redis.host")
	port := config.GetInt("redis.port")
	db := config.GetInt("redis.db")

	return redis.NewStandaloneClient(redis.StandaloneOptions{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	}), nil
}
