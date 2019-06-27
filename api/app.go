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
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/topfreegames/podium/leaderboard"

	"github.com/getsentry/raven-go"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/engine/fasthttp"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	newrelic "github.com/newrelic/go-agent"
	"github.com/rcrowley/go-metrics"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/echo"
	extechomiddleware "github.com/topfreegames/extensions/echo/middleware"
	"github.com/topfreegames/extensions/jaeger"
	extnethttpmiddleware "github.com/topfreegames/extensions/middleware"
	"github.com/topfreegames/podium/log"
	"go.uber.org/zap"

	api "github.com/topfreegames/podium/proto/podium/api/v1"
)

// JSON type
type JSON map[string]interface{}

// App is a struct that represents a podium Application
type App struct {
	Debug        bool
	Fast         bool
	HTTPPort     int
	GRPCPort     int
	HTTPEndpoint string
	GRPCEndpoint string
	Host         string
	ConfigPath   string
	Errors       metrics.EWMA
	App          *echo.Echo
	Engine       engine.Server
	grpcServer   *grpc.Server
	httpServer   *http.Server
	Config       *viper.Viper
	Logger       zap.Logger
	Leaderboards *leaderboard.Client
	NewRelic     newrelic.Application
	DDStatsD     *extnethttpmiddleware.DogStatsD
}

// GetApp returns a new podium Application
func GetApp(host string, httpPort, grpcPort int, configPath string, debug, fast bool, logger zap.Logger) (*App, error) {
	app := &App{
		Host:         host,
		HTTPPort:     httpPort,
		GRPCPort:     grpcPort,
		HTTPEndpoint: fmt.Sprintf("localhost:%d", httpPort),
		GRPCEndpoint: fmt.Sprintf("localhost:%d", grpcPort),
		ConfigPath:   configPath,
		Config:       viper.New(),
		Debug:        debug,
		Logger:       logger,
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

	err = app.configureStatsD()
	if err != nil {
		return err
	}

	err = app.configureNewRelic()
	if err != nil {
		return err
	}

	err = app.configureApplication()
	if err != nil {
		return err
	}

	//we are customizing the default http error reply
	runtime.HTTPError = func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler,
		w http.ResponseWriter, _ *http.Request, err error) {
		w.Header().Set("Content-Type", marshaler.ContentType())
		var s int

		st, ok := status.FromError(err)
		if !ok {
			s = http.StatusInternalServerError
		} else {
			s = runtime.HTTPStatusFromCode(st.Code())
		}

		w.WriteHeader(s)

		type errorBody struct {
			Success bool   `json:"success"`
			Reason  string `json:"reason"`
		}

		body := &errorBody{
			Success: false,
			Reason:  st.Message(),
		}

		buf, merr := marshaler.Marshal(body)
		if merr != nil {
			app.Logger.Error("Failed to marshal error body,", zap.Error(merr))
		}
		if _, err := w.Write(buf); err != nil {
			app.Logger.Error("Failed to write response.,", zap.Error(err))
		}
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

func (app *App) configureStatsD() error {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureStatsD"),
	)

	ddstatsd, err := extnethttpmiddleware.NewDogStatsD(app.Config)
	if err != nil {
		log.E(l, "Failed to initialize DogStatsD.", func(cm log.CM) {
			cm.Write(zap.Error(err))
		})
		return err
	}
	app.DDStatsD = ddstatsd
	l.Info("Configured StatsD successfully.")

	return nil
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
		Disabled:    app.Config.GetBool("jaeger.disabled"),
		Probability: app.Config.GetFloat64("jaeger.samplingProbability"),
		ServiceName: "podium",
	}

	_, err := jaeger.Configure(opts)
	if err != nil {
		l.Error("Failed to initialize Jaeger.")
	}
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("graceperiod.ms", 5000)
	app.Config.SetDefault("api.maxReturnedMembers", 2000)
	app.Config.SetDefault("api.maxReadBufferSize", 32000)
	app.Config.SetDefault("redis.host", "localhost")
	app.Config.SetDefault("redis.port", 1212)
	app.Config.SetDefault("redis.password", "")
	app.Config.SetDefault("redis.db", 0)
	app.Config.SetDefault("redis.connectionTimeout", 200)
	app.Config.SetDefault("jaeger.disabled", true)
	app.Config.SetDefault("jaeger.samplingProbability", 0.001)
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

	app.Engine = standard.New(fmt.Sprintf("%s:%d", app.Host, app.HTTPPort))
	if app.Fast {
		rb := app.Config.GetInt("api.maxReadBufferSize")
		engine := fasthttp.New(fmt.Sprintf("%s:%d", app.Host, app.HTTPPort))
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
	a.Use(extechomiddleware.NewResponseTimeMetricsMiddleware(app.DDStatsD).Serve)
	a.Use(NewVersionMiddleware().Serve)
	a.Use(NewSentryMiddleware(app).Serve)
	a.Use(NewNewRelicMiddleware(app, app.Logger).Serve)

	a.Put("/m/:memberPublicID/scores", UpsertMemberLeaderboardsScoreHandler(app))
	a.Get("/m/:memberPublicID/scores", GetMemberRankInManyLeaderboardsHandler(app))
	a.Get("/l/:leaderboardID/scores/:score/around", GetAroundScoreHandler(app))

	app.Errors = metrics.NewEWMA15()

	go func() {
		app.Errors.Tick()
		time.Sleep(5 * time.Second)
	}()

	host := app.Config.GetString("redis.host")
	port := app.Config.GetInt("redis.port")
	password := app.Config.GetString("redis.password")
	db := app.Config.GetInt("redis.db")
	connectionTimeout := app.Config.GetInt("redis.connectionTimeout")

	rl := l.With(
		zap.String("url", fmt.Sprintf("redis://:<REDACTED>@%s:%v/%v", host, port, db)),
		zap.Int("connectionTimeout", connectionTimeout),
	)

	rl.Info("Creating leaderboard client.")
	cli, err := leaderboard.NewClient(host, port, password, db, connectionTimeout)
	if err != nil {
		return err
	}
	app.Leaderboards = cli
	rl.Info("Leaderboard client creation successfull.")

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
		zap.String("host", app.Host),
		zap.Int("port", app.HTTPPort),
	)

	//errch is the channel for retrieving errors from server goroutines.
	errch := make(chan error, 2)

	go app.startGRPCServer(errch)
	go app.startHTTPServer(errch)

	log.I(l, "app started")
	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	// stop server
	select {
	case s := <-sg:
		graceperiod := app.Config.GetInt("graceperiod.ms")
		log.I(l, "shutting down", func(cm log.CM) {
			cm.Write(zap.String("signal", fmt.Sprintf("%v", s)),
				zap.Int("graceperiod", graceperiod))
		})
		time.Sleep(time.Duration(graceperiod) * time.Millisecond)
	case err := <-errch:
		return err
	}
	log.I(l, "app stopped")
	return nil
}

func (app *App) startGRPCServer(errch chan<- error) {
	lis, err := net.Listen("tcp", app.GRPCEndpoint)
	if err != nil {
		errch <- fmt.Errorf("error trying to listen for connections: %v", err)
		return
	}

	app.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			grpc.UnaryServerInterceptor(app.newRelicMiddleware),
		),
	))
	api.RegisterPodiumAPIServer(app.grpcServer, app)

	if err := app.grpcServer.Serve(lis); err != nil {
		errch <- fmt.Errorf("error trying to serve with grpc server: %v", err)
	}
}

func (app *App) startHTTPServer(errch chan<- error) {
	gatewayMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard,
		&runtime.JSONPb{EmitDefaults: true}))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	endpoint := fmt.Sprintf("localhost:%d", app.GRPCPort)

	if err := api.RegisterPodiumAPIHandlerFromEndpoint(context.Background(), gatewayMux, endpoint, opts); err != nil {
		errch <- fmt.Errorf("error registering multiplexer for grpc gateway: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gatewayMux)
	mux.HandleFunc("/healthcheck", app.healthCheckHandler)
	mux.HandleFunc("/status", app.statusHandler)

	app.httpServer = &http.Server{
		Addr:    app.HTTPEndpoint,
		Handler: mux,
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	list, err := net.Listen("tcp", app.HTTPEndpoint)
	if err != nil {
		errch <- fmt.Errorf("error listening on HTTPEndpoint: %v", err)
		return
	}

	if err := app.httpServer.Serve(list); err != http.ErrServerClosed {
		errch <- fmt.Errorf("error listening and serving http requests: %v", err)
	}
}

func (app *App) Stop() {
	if app.grpcServer != nil {
		app.grpcServer.GracefulStop()
	}
	if app.httpServer != nil {
		if err := app.httpServer.Shutdown(context.Background()); err != nil {
			app.Logger.Error("HTTP server Shutdown.", zap.Error(err))
		}
	}
}

func (app *App) closeAll(ctx context.Context) {
	if app.httpServer != nil {
		_ = app.httpServer.Shutdown(ctx)
		_ = app.httpServer.Close()
	}
	if app.grpcServer != nil {
		app.grpcServer.GracefulStop()
	}
}
