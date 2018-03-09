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
	"github.com/labstack/echo/middleware"
	newrelic "github.com/newrelic/go-agent"
	"github.com/rcrowley/go-metrics"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/jaeger"
	"github.com/topfreegames/extensions/redis"
	"github.com/topfreegames/podium/log"
	"go.uber.org/zap"
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
	RedisClient *redis.Client
	NewRelic    newrelic.Application
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

	app.configureJaeger()
	app.configureSentry()
	err = app.configureNewRelic()
	if err != nil {
		return err
	}

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
	raven.SetDSN(sentryURL)
	l.Info("Configured sentry successfully.")
}

func (app *App) configureNewRelic() error {
	newRelicKey := app.Config.GetString("newrelic.key")

	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureNewRelic"),
	)

	config := newrelic.NewConfig("Podium", newRelicKey)
	if newRelicKey == "" {
		l.Info("New Relic is not enabled..")
		config.Enabled = false
	}
	nr, err := newrelic.NewApplication(config)
	if err != nil {
		l.Error("Failed to initialize New Relic.", zap.Error(err))
		return err
	}

	app.NewRelic = nr
	l.Info("Initialized New Relic successfully.")

	return nil
}

func (app *App) configureJaeger() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureJaeger"),
	)

	opts := jaeger.Options{
		Disabled:    false,
		Probability: 1.0,
		ServiceName: "podium",
	}

	_, err := jaeger.Configure(opts)
	if err != nil {
		l.Error("Failed to initialize Jaeger.")
	}
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("api.maxReturnedMembers", 2000)
	app.Config.SetDefault("api.maxReadBufferSize", 32000)
	app.Config.SetDefault("redis.url", "redis://localhost:1212/0")
	app.Config.SetDefault("redis.connectionTimeout", 200)
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

	basicAuthUser := app.Config.GetString("basicauth.username")
	if basicAuthUser != "" {
		basicAuthPass := app.Config.GetString("basicauth.password")

		a.Use(middleware.BasicAuth(func(username, password string) bool {
			return username == basicAuthUser && password == basicAuthPass
		}))
	}

	a.Pre(middleware.RemoveTrailingSlash())
	a.Use(NewLoggerMiddleware(app.Logger).Serve)
	a.Use(NewRecoveryMiddleware(app.OnErrorHandler).Serve)
	a.Use(NewVersionMiddleware().Serve)
	a.Use(NewSentryMiddleware(app).Serve)
	a.Use(NewNewRelicMiddleware(app, app.Logger).Serve)

	jaeger.InstrumentEcho(a)

	a.Get("/healthcheck", HealthCheckHandler(app))
	a.Get("/status", StatusHandler(app))
	a.Delete("/l/:leaderboardID", RemoveLeaderboardHandler(app))
	a.Put("/l/:leaderboardID/members/:memberPublicID/score", UpsertMemberScoreHandler(app))
	a.Patch("/l/:leaderboardID/members/:memberPublicID/score", IncrementMemberScoreHandler(app))
	a.Get("/l/:leaderboardID/members/:memberPublicID", GetMemberHandler(app))
	a.Get("/l/:leaderboardID/members", GetMembersHandler(app))
	a.Delete("/l/:leaderboardID/members", RemoveMembersHandler(app))
	a.Delete("/l/:leaderboardID/members/:memberPublicID", RemoveMemberHandler(app))
	a.Get("/l/:leaderboardID/members/:memberPublicID/rank", GetMemberRankHandler(app))
	a.Get("/l/:leaderboardID/members/:memberPublicID/around", GetAroundMemberHandler(app))
	a.Get("/l/:leaderboardID/members-count", GetTotalMembersHandler(app))
	a.Get("/l/:leaderboardID/top/:pageNumber", GetTopMembersHandler(app))
	a.Get("/l/:leaderboardID/top-percent/:percentage", GetTopPercentageHandler(app))
	a.Put("/m/:memberPublicID/scores", UpsertMemberLeaderboardsScoreHandler(app))
	a.Get("/m/:memberPublicID/scores", GetMemberRankInManyLeaderboardsHandler(app))

	app.Errors = metrics.NewEWMA15()

	go func() {
		app.Errors.Tick()
		time.Sleep(5 * time.Second)
	}()

	redisURL := app.Config.GetString("redis.url")
	redisConnectionTimeout := app.Config.GetString("redis.connectionTimeout")

	rl := l.With(
		zap.String("url", redisURL),
		zap.String("connectionTimeout", redisConnectionTimeout),
	)
	rl.Debug("Connecting to redis...")
	cli, err := redis.NewClient("redis", app.Config)
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
func (app *App) Start() error {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "Start"),
	)

	err := app.App.Run(app.Engine)
	if err != nil {
		log.E(l, "App failed to start.", func(cm log.CM) {
			cm.Write(
				zap.String("host", app.Host),
				zap.Int("port", app.Port),
				zap.Error(err),
			)
		})
		return err
	}

	log.I(l, "App started.", func(cm log.CM) {
		cm.Write(zap.String("host", app.Host), zap.Int("port", app.Port))
	})
	return nil
}
