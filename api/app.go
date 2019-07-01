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
	"sync"
	"syscall"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/topfreegames/podium/leaderboard"

	"github.com/getsentry/raven-go"
	newrelic "github.com/newrelic/go-agent"
	"github.com/rcrowley/go-metrics"
	"github.com/spf13/viper"
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
	Debug                bool
	Fast                 bool
	HTTPPort             int
	GRPCPort             int
	HTTPEndpoint         string
	GRPCEndpoint         string
	httpReady, grpcReady chan bool
	Host                 string
	ConfigPath           string
	Errors               metrics.EWMA
	grpcServer           *grpc.Server
	httpServer           *http.Server
	Config               *viper.Viper
	Logger               zap.Logger
	Leaderboards         *leaderboard.Client
	NewRelic             newrelic.Application
	DDStatsD             *extnethttpmiddleware.DogStatsD
}

// GetApp returns a new podium Application.
// If httpPort is sent as zero, a random port will be selected (the same will happen for grpcPort)
func GetApp(host string, httpPort, grpcPort int, configPath string, debug bool, logger zap.Logger) (*App, error) {
	app := &App{
		Host:       host,
		HTTPPort:   httpPort,
		GRPCPort:   grpcPort,
		httpReady:  make(chan bool, 1),
		grpcReady:  make(chan bool, 1),
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

func (app *App) getStatusCodeFromError(err error) (*status.Status, int) {
	var statusCode int
	st, ok := status.FromError(err)
	if !ok {
		statusCode = http.StatusInternalServerError
	} else {
		statusCode = runtime.HTTPStatusFromCode(st.Code())
	}
	return st, statusCode
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

		st, s := app.getStatusCodeFromError(err)

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

func buildAddress(port int) string {
	if port == 0 {
		return "localhost:"
	} else {
		return fmt.Sprintf("localhost:%d", port)
	}
}

// Start starts listening for web requests at specified host and port
func (app *App) Start() error {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "Start"),
		zap.String("host", app.Host),
		zap.Int("port", app.HTTPPort),
	)

	grpcLis, err := net.Listen("tcp", buildAddress(app.GRPCPort))
	if err != nil {
		return fmt.Errorf("error trying to listen for connections: %v", err)
	}
	app.GRPCEndpoint = grpcLis.Addr().String()

	httpLis, err := net.Listen("tcp", buildAddress(app.HTTPPort))
	if err != nil {
		return fmt.Errorf("error listening on HTTPEndpoint: %v", err)
	}
	app.HTTPEndpoint = httpLis.Addr().String()

	//errch is the channel for retrieving errors from server goroutines.
	errch := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		app.startGRPCServer(grpcLis, errch)
		wg.Done()
	}()
	go func() {
		app.startHTTPServer(httpLis, errch)
		wg.Done()
	}()

	stopped := make(chan bool, 1)
	go func() {
		wg.Wait()
		stopped <- true
	}()

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
	case <-stopped:
	}
	log.I(l, "app stopped")
	return nil
}

func (app *App) startGRPCServer(lis net.Listener, errch chan<- error) {
	var basicAuthInterceptor grpc.UnaryServerInterceptor

	basicAuthUser := app.Config.GetString("basicauth.username")
	if basicAuthUser == "" {
		basicAuthInterceptor = grpc.UnaryServerInterceptor(app.noAuthMiddleware)
	} else {
		basicAuthInterceptor = grpc_auth.UnaryServerInterceptor(app.basicAuthMiddleware)
	}

	app.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			basicAuthInterceptor,
			grpc.UnaryServerInterceptor(app.loggerMiddleware),
			grpc.UnaryServerInterceptor(app.recoveryMiddleware),
			grpc.UnaryServerInterceptor(app.responseTimeMetricsMiddleware),
			grpc.UnaryServerInterceptor(app.sentryMiddleware),
			grpc.UnaryServerInterceptor(app.newRelicMiddleware),
		),
	))
	api.RegisterPodiumAPIServer(app.grpcServer, app)

	app.grpcReady <- true
	if err := app.grpcServer.Serve(lis); err != nil {
		errch <- fmt.Errorf("error trying to serve with grpc server: %v", err)
	}
}

func (app *App) startHTTPServer(lis net.Listener, errch chan<- error) {
	gatewayMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard,
		&runtime.JSONPb{EmitDefaults: true}))
	opts := []grpc.DialOption{grpc.WithInsecure()}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := api.RegisterPodiumAPIHandlerFromEndpoint(ctx, gatewayMux, app.GRPCEndpoint, opts); err != nil {
		errch <- fmt.Errorf("error registering multiplexer for grpc gateway: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", removeTrailingSlashMiddleware{addVersionMiddleware{gatewayMux}})
	mux.HandleFunc("/healthcheck", addVersionHandlerFunc(app.healthCheckHandler))
	mux.HandleFunc("/status", addVersionHandlerFunc(app.statusHandler))

	app.httpServer = &http.Server{
		Addr:    app.HTTPEndpoint,
		Handler: mux,
	}

	app.httpReady <- true
	if err := app.httpServer.Serve(lis); err != http.ErrServerClosed {
		errch <- fmt.Errorf("error listening and serving http requests: %v", err)
	}
}

// WaitForReady blocks until App is ready to serve requests or the timeout is reached.
// An error is returned on timeout.
func (app *App) WaitForReady(d time.Duration) error {
	isReady := func(c chan bool) bool {
		select {
		case <-c:
			return true
		case <-time.After(d):
			return false
		}
	}

	if isReady(app.grpcReady) && isReady(app.httpReady) {
		return nil
	}
	return fmt.Errorf("timed out waiting for endpoints")
}

// GracefulStop attempts to stop the server.
func (app *App) GracefullStop() {
	if app.grpcServer != nil {
		app.grpcServer.GracefulStop()
	}
	if app.httpServer != nil {
		if err := app.httpServer.Shutdown(context.Background()); err != nil {
			app.Logger.Error("HTTP server Shutdown.", zap.Error(err))
		}
	}
}
