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
	"os"
	"strings"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/fasthttp"
	"github.com/labstack/echo/engine/standard"
	"github.com/rcrowley/go-metrics"
	"github.com/spf13/viper"
	"github.com/topfreegames/podium/log"
	"github.com/topfreegames/podium/util"
	"github.com/uber-go/zap"
)

// JSON type
type JSON map[string]interface{}

// App is a struct that represents a podium Application
type App struct {
	Debug       bool
	Fast        bool
	Port        int
	Host        string
	ConfigPath  string
	Errors      metrics.EWMA
	App         *echo.Echo
	Engine      engine.Server
	Config      *viper.Viper
	Logger      zap.Logger
	RedisClient *util.RedisClient
}

// GetApp returns a new podium Application
func GetApp(host string, port int, configPath string, debug, fast bool, logger zap.Logger) (*App, error) {
	app := &App{
		Host:       host,
		Port:       port,
		ConfigPath: configPath,
		Config:     viper.New(),
		Debug:      debug,
		Logger:     logger,
	}
	err := app.Configure()
	if err != nil {
		return nil, err
	}
	return app, nil
}

// Configure instantiates the required dependencies for podium Application
func (app *App) Configure() error {
	app.setConfigurationDefaults()

	err := app.loadConfiguration()
	if err != nil {
		return err
	}

	app.configureSentry()

	err = app.configureApplication()
	if err != nil {
		return err
	}

	return nil
}

func (app *App) configureSentry() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureSentry"),
	)
	sentryURL := app.Config.GetString("sentry.url")
	l.Info(fmt.Sprintf("Configuring sentry with URL %s", sentryURL))
	raven.SetDSN(sentryURL)
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("api.maxReturnedMembers", 2000)
	app.Config.SetDefault("api.maxReadBufferSize", 32000)
	app.Config.SetDefault("redis.host", "localhost")
	app.Config.SetDefault("redis.port", 1212)
	app.Config.SetDefault("redis.password", "")
	app.Config.SetDefault("redis.db", 0)
	app.Config.SetDefault("redis.maxPoolSize", 20)
}

func (app *App) loadConfiguration() error {
	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.SetEnvPrefix("podium")
	app.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	app.Config.AutomaticEnv()

	if err := app.Config.ReadInConfig(); err == nil {
		app.Logger.Info("Loaded config file.", zap.String("configFile", app.Config.ConfigFileUsed()))
	} else {
		return fmt.Errorf("Could not load configuration file from: %s", app.ConfigPath)
	}

	return nil
}

//OnErrorHandler handles panics
func (app *App) OnErrorHandler(err error, stack []byte) {
	app.Logger.Error(
		"Panic occurred.",
		zap.Object("panicText", err),
		zap.String("stack", string(stack)),
	)

	tags := map[string]string{
		"source": "app",
		"type":   "panic",
	}
	raven.CaptureError(err, tags)
}

func (app *App) configureApplication() error {
	l := app.Logger.With(
		zap.String("operation", "configureApplication"),
	)

	app.Engine = standard.New(fmt.Sprintf("%s:%d", app.Host, app.Port))
	if app.Fast {
		rb := app.Config.GetInt("api.maxReadBufferSize")
		engine := fasthttp.New(fmt.Sprintf("%s:%d", app.Host, app.Port))
		engine.ReadBufferSize = rb
		app.Engine = engine
	}
	app.App = echo.New()
	a := app.App

	_, w, _ := os.Pipe()
	a.SetLogOutput(w)

	a.Use(NewLoggerMiddleware(app.Logger).Serve)
	a.Use(NewRecoveryMiddleware(app.OnErrorHandler).Serve)
	a.Use(NewVersionMiddleware().Serve)
	a.Use(NewSentryMiddleware(app).Serve)

	a.Get("/healthcheck", HealthCheckHandler(app))
	a.Get("/status", StatusHandler(app))
	//a.Delete("/l/:leaderboardID", RemoveLeaderboardHandler(app))
	//a.Put("/l/:leaderboardID/members/:memberPublicID/score", UpsertMemberScoreHandler(app))
	//a.Get("/l/:leaderboardID/members/:memberPublicID", GetMemberHandler(app))
	//a.Get("/l/:leaderboardID/members", GetMembersHandler(app))
	//a.Delete("/l/:leaderboardID/members", RemoveMembersHandler(app))
	//a.Delete("/l/:leaderboardID/members/:memberPublicID", RemoveMemberHandler(app))
	//a.Get("/l/:leaderboardID/members/:memberPublicID/rank", GetMemberRankHandler(app))
	//a.Get("/l/:leaderboardID/members/:memberPublicID/around", GetAroundMemberHandler(app))
	//a.Get("/l/:leaderboardID/members-count", GetTotalMembersHandler(app))
	//a.Get("/l/:leaderboardID/top/:pageNumber", GetTopMembersHandler(app))
	//a.Get("/l/:leaderboardID/top-percent/:percentage", GetTopPercentageHandler(app))
	//a.Put("/m/:memberPublicID/scores", UpsertMemberLeaderboardsScoreHandler(app))
	//a.Get("/m/:memberPublicID/scores", GetMemberRankInManyLeaderboardsHandler(app))

	app.Errors = metrics.NewEWMA15()

	go func() {
		app.Errors.Tick()
		time.Sleep(5 * time.Second)
	}()

	redisHost := app.Config.GetString("redis.host")
	redisPort := app.Config.GetInt("redis.port")
	redisPass := app.Config.GetString("redis.password")
	redisDB := app.Config.GetInt("redis.db")
	redisMaxPoolSize := app.Config.GetInt("redis.maxPoolSize")

	rl := l.With(
		zap.String("host", redisHost),
		zap.Int("port", redisPort),
		zap.Int("db", redisDB),
		zap.Int("maxPoolSize", redisMaxPoolSize),
	)
	rl.Debug("Connecting to redis...")
	cli, err := util.GetRedisClient(redisHost, redisPort, redisPass, redisDB, redisMaxPoolSize, app.Logger)
	if err != nil {
		return err
	}
	app.RedisClient = cli
	rl.Info("Connected to redis successfully.")
	return nil
}

//AddError rate statistics
func (app *App) AddError() {
	app.Errors.Update(1)
}

// Start starts listening for web requests at specified host and port
func (app *App) Start() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "Start"),
	)

	log.D(l, "App started.", func(cm log.CM) {
		cm.Write(zap.String("host", app.Host), zap.Int("port", app.Port))
	})

	app.App.Run(app.Engine)
}
