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
	"strings"
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/config"
	"github.com/rcrowley/go-metrics"
	"github.com/spf13/viper"
	"github.com/topfreegames/podium/util"
	"github.com/uber-go/zap"
)

// JSON type
type JSON map[string]interface{}

// App is a struct that represents a podium Application
type App struct {
	Debug       bool
	Port        int
	Host        string
	ConfigPath  string
	Errors      metrics.EWMA
	App         *iris.Framework
	Config      *viper.Viper
	Logger      zap.Logger
	RedisClient *util.RedisClient
}

// GetApp returns a new podium Application
func GetApp(host string, port int, configPath string, debug bool, logger zap.Logger) *App {
	app := &App{
		Host:       host,
		Port:       port,
		ConfigPath: configPath,
		Config:     viper.New(),
		Debug:      debug,
		Logger:     logger,
	}
	app.Configure()
	return app
}

// Configure instantiates the required dependencies for podium Application
func (app *App) Configure() {
	app.setConfigurationDefaults()
	app.loadConfiguration()
	app.configureApplication()
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("api.maxReturnedMembers", 2000)
	app.Config.SetDefault("redis.host", "localhost")
	app.Config.SetDefault("redis.port", 1212)
	app.Config.SetDefault("redis.password", "")
	app.Config.SetDefault("redis.db", 0)
}

func (app *App) loadConfiguration() {
	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.SetEnvPrefix("podium")
	app.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	app.Config.AutomaticEnv()

	if err := app.Config.ReadInConfig(); err == nil {
		app.Logger.Info("Loaded config file.", zap.String("configFile", app.Config.ConfigFileUsed()))
	} else {
		panic(fmt.Sprintf("Could not load configuration file from: %s", app.ConfigPath))
	}
}

func (app *App) onErrorHandler(err interface{}, stack []byte) {
	app.Logger.Error(
		"Panic occurred.",
		zap.Object("panicText", err),
		zap.String("stack", string(stack)),
	)
}

func (app *App) configureApplication() {
	l := app.Logger.With(
		zap.String("operation", "configureApplication"),
	)

	c := config.Iris{
		DisableBanner: true,
	}

	app.App = iris.New(c)
	a := app.App

	a.Use(&RecoveryMiddleware{OnError: app.onErrorHandler})
	a.Use(&VersionMiddleware{App: app})

	a.Get("/healthcheck", HealthCheckHandler(app))
	a.Put("/l/:leaderboardID/users/:userPublicID/score", UpsertUserScoreHandler(app))
	a.Get("/l/:leaderboardID/users/:userPublicID", GetUserHandler(app))
	a.Delete("/l/:leaderboardID/users/:userPublicID", RemoveUserHandler(app))
	a.Get("/l/:leaderboardID/users/:userPublicID/rank", GetUserRankHandler(app))
	a.Get("/l/:leaderboardID/users/:userPublicID/around", GetAroundUserHandler(app))
	a.Get("/l/:leaderboardID/users-count", GetTotalMembersHandler(app))
	a.Get("/l/:leaderboardID/pages", GetTotalPagesHandler(app))
	a.Get("/l/:leaderboardID/top/:pageNumber", GetTopUsersHandler(app))
	a.Get("/l/:leaderboardID/top-percent/:percentage", GetTopPercentageHandler(app))

	app.Errors = metrics.NewEWMA15()

	go func() {
		app.Errors.Tick()
		time.Sleep(5 * time.Second)
	}()

	redisHost := app.Config.GetString("redis.host")
	redisPort := app.Config.GetInt("redis.port")
	redisPass := app.Config.GetString("redis.password")
	redisDB := app.Config.GetInt("redis.db")

	rl := l.With(
		zap.String("host", redisHost),
		zap.Int("port", redisPort),
		zap.Int("db", redisDB),
	)
	rl.Debug("Connecting to redis...")
	cli, err := util.GetRedisClient(redisHost, redisPort, redisPass, redisDB, app.Logger)
	if err != nil {
		panic(fmt.Sprintf("Could not connect to redis: %s", err))
	}
	app.RedisClient = cli
	rl.Info("Connected to redis successfully.")
}

func (app *App) addError() {
	app.Errors.Update(1)
}

// Start starts listening for web requests at specified host and port
func (app *App) Start() {
	app.App.Listen(fmt.Sprintf("%s:%d", app.Host, app.Port))
}
