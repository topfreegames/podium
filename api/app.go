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

	"github.com/topfreegames/podium/config"
	"github.com/topfreegames/podium/leaderboard/v2/enriching"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/getsentry/raven-go"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/rcrowley/go-metrics"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/service"
	lservice "github.com/topfreegames/podium/leaderboard/v2/service"
	"github.com/topfreegames/podium/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	extnethttpmiddleware "github.com/topfreegames/extensions/middleware"
	api "github.com/topfreegames/podium/proto/podium/api/v1"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

// JSON type
type JSON map[string]interface{}

// App is a struct that represents a podium Application
type App struct {
	api.UnimplementedPodiumServer

	ConfigPath string
	Debug      bool

	// HTTP endpoint for HTTP requests. Built after calling Start. Format: 127.0.0.1:8080
	HTTPEndpoint string

	// GRPC endpoint for GRPC requests. Built after calling Start. Format: 127.0.0.1:8081
	GRPCEndpoint string

	httpReady, grpcReady chan bool

	Config       *viper.Viper
	ParsedConfig *config.PodiumConfig
	DDStatsD     *extnethttpmiddleware.DogStatsD
	Enricher     enriching.Enricher
	Errors       metrics.EWMA
	grpcServer   *grpc.Server
	httpServer   *http.Server
	ID           uuid.UUID
	Logger       *zap.Logger
	Leaderboards lservice.Leaderboard
}

// New returns a new podium Application.
// If httpPort is sent as zero, a random port will be selected (the same will happen for grpcPort)
func New(host string, httpPort, grpcPort int, configPath string, debug bool, logger *zap.Logger) (*App, error) {
	app := &App{
		HTTPEndpoint: fmt.Sprintf("%s:%d", host, httpPort),
		GRPCEndpoint: fmt.Sprintf("%s:%d", host, grpcPort),
		httpReady:    make(chan bool, 1),
		grpcReady:    make(chan bool, 1),
		ConfigPath:   configPath,
		Config:       viper.New(),
		Debug:        debug,
		Logger:       logger,
		ID:           uuid.NewV4(),
	}
	err := app.configure()
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
func (app *App) configure() error {
	app.setConfigurationDefaults()

	err := app.loadConfiguration()
	if err != nil {
		return err
	}

	app.configureJaeger()

	app.configureEnrichment()

	err = app.configureStatsD()
	if err != nil {
		return err
	}

	err = app.configureApplication()
	if err != nil {
		return err
	}

	return nil
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

func (app *App) configureJaeger() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureJaeger"),
	)

	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		l.Error("Could not parse Jaeger env vars", zap.Error(err))
		return
	}

	if cfg.ServiceName == "" {
		cfg.ServiceName = "podium"
	}

	tracer, _, err := cfg.NewTracer()
	if err != nil {
		l.Error("Could not initialize jaeger tracer", zap.Error(err))
		return
	}

	opentracing.SetGlobalTracer(tracer)
	l.Info(fmt.Sprintf("Tracer configured for %s", cfg.Reporter.LocalAgentHostPort))
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("graceperiod.ms", 50)
	app.Config.SetDefault("api.maxReturnedMembers", 2000)
	app.Config.SetDefault("api.maxReadBufferSize", 32000)
	app.Config.SetDefault("redis.host", "localhost")
	app.Config.SetDefault("redis.port", 6379)
	app.Config.SetDefault("redis.password", "")
	app.Config.SetDefault("redis.db", 0)
	app.Config.SetDefault("redis.connectionTimeout", 200)
	app.Config.SetDefault("redis.cluster.enabled", false)
}

func (app *App) loadConfiguration() error {
	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.SetEnvPrefix("podium")
	app.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	app.Config.AutomaticEnv()

	if err := app.Config.ReadInConfig(); err == nil {
		app.Logger.Info("Loaded config file.", zap.String("configFile", app.Config.ConfigFileUsed()))
	} else {
		return fmt.Errorf("could not load configuration file from: %s", app.ConfigPath)
	}

	app.ParsedConfig = &config.PodiumConfig{}
	if err := app.Config.Unmarshal(app.ParsedConfig, config.DecodeHook()); err != nil {
		return fmt.Errorf("could not parse configuration file: %w", err)
	}

	return nil
}

func (app *App) configureEnrichment() {
	enricher := enriching.NewEnricher(app.ParsedConfig.Enrichment, app.Logger)
	app.Enricher = enriching.NewInstrumentedEnricher(enricher, app.DDStatsD)
}

// OnErrorHandler handles panics
func (app *App) OnErrorHandler(err error, stack []byte) {
	app.Logger.Error(
		"Panic occurred.",
		zap.Any("panicText", err),
		zap.String("stack", string(stack)),
	)

	tags := map[string]string{
		"source": "app",
		"type":   "panic",
	}
	raven.CaptureError(err, tags)
}

func (app *App) configureApplication() error {
	app.Errors = metrics.NewEWMA15()

	go func() {
		app.Errors.Tick()
		time.Sleep(5 * time.Second)
	}()

	client, err := app.createAndConfigureLeaderboardClient()
	if err != nil {
		return err
	}
	app.Leaderboards = client

	return nil
}

func (app *App) createAndConfigureLeaderboardClient() (lservice.Leaderboard, error) {
	client := app.createLeaderboardClient()
	err := client.Healthcheck(context.Background())

	if err != nil {
		return nil, err
	}

	app.Logger.Info("Leaderboard client creation successfull.")
	return client, nil
}

func (app *App) createLeaderboardClient() lservice.Leaderboard {
	shouldRunOnCluster := app.Config.GetBool("redis.cluster.enabled")
	addrs := app.Config.GetStringSlice("redis.addrs")
	password := app.Config.GetString("redis.password")
	host := app.Config.GetString("redis.host")
	port := app.Config.GetInt("redis.port")
	db := app.Config.GetInt("redis.db")

	logger := app.Logger.With(
		zap.String("operation", "createLeaderboardClient"),
		zap.Strings("addrs", addrs),
		zap.Bool("cluster", shouldRunOnCluster),
		zap.String("url", fmt.Sprintf("redis://:<REDACTED>@%s:%v/%v", host, port, db)),
	)

	leaderboardService := service.NewService(database.NewRedisDatabase(database.RedisOptions{
		ClusterEnabled: shouldRunOnCluster,
		Addrs:          addrs,
		Host:           host,
		Password:       password,
		Port:           port,
		DB:             db,
	}))

	logger.Info("Creating leaderboard client.")

	return leaderboardService
}

// AddError rate statistics
func (app *App) AddError() {
	app.Errors.Update(1)
}

// Start starts listening for web requests at specified host and port
func (app *App) Start(ctx context.Context) error {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "Start"),
		zap.String("HTTPEndpoint", app.HTTPEndpoint),
		zap.String("GRPCEndPoint", app.GRPCEndpoint),
	)

	grpcLis, err := net.Listen("tcp", app.GRPCEndpoint)
	if err != nil {
		return fmt.Errorf("error trying to listen for connections: %v", err)
	}
	app.GRPCEndpoint = grpcLis.Addr().String()

	httpLis, err := net.Listen("tcp", app.HTTPEndpoint)
	if err != nil {
		return fmt.Errorf("error listening on HTTPEndpoint: %v", err)
	}
	app.HTTPEndpoint = httpLis.Addr().String()

	//errch is the channel for retrieving errors from server goroutines.
	errch := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := app.startGRPCServer(grpcLis); err != nil {
			errch <- err
		}
	}()

	go func() {
		defer wg.Done()
		if err := app.startHTTPServer(ctx, httpLis); err != nil {
			errch <- err
		}
	}()

	stopped := make(chan bool, 1)
	go func() {
		wg.Wait()
		stopped <- true
	}()

	log.I(l, "app started")
	sg := make(chan os.Signal)
	//TODO verify that capturing SIGKILL actually works. Signal handling should be moved outside of Start.
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	// stop server
	select {
	case s := <-sg:
		graceperiod := app.Config.GetInt("graceperiod.ms")
		log.I(l, "shutting down", func(cm log.CM) {
			cm.Write(zap.String("signal", fmt.Sprintf("%v", s)),
				zap.Int("graceperiod", graceperiod))
		})
		app.GracefullStop()
		time.Sleep(time.Duration(graceperiod) * time.Millisecond)
	case err := <-errch:
		app.Logger.Error("Err on start server", zap.Error(err))
		return err
	case <-stopped:
	}
	log.I(l, "app stopped")
	return nil
}

func (app *App) startGRPCServer(lis net.Listener) error {
	var basicAuthInterceptor grpc.UnaryServerInterceptor

	basicAuthUser := app.Config.GetString("basicauth.username")
	if basicAuthUser == "" {
		basicAuthInterceptor = app.noAuthMiddleware
	} else {
		basicAuthInterceptor = grpcauth.UnaryServerInterceptor(app.basicAuthMiddleware)
	}

	app.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(
		grpcmiddleware.ChainUnaryServer(
			basicAuthInterceptor,
			otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer()),
			app.loggerMiddleware,
			app.recoveryMiddleware,
			app.responseTimeMetricsMiddleware,
		),
	))
	api.RegisterPodiumServer(app.grpcServer, app)

	app.grpcReady <- true
	if err := app.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("error trying to serve with grpc server: %v", err)
	}

	return nil
}

func (app *App) applicationErrorHandler(_ context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, rpcErr error) {

	w.Header().Set("Content-Type", "application/json")
	st, s := app.getStatusCodeFromError(rpcErr)

	w.WriteHeader(s)

	type errorBody struct {
		Success bool   `json:"success"`
		Reason  string `json:"reason"`
	}

	body := &errorBody{
		Success: false,
		Reason:  st.Message(),
	}

	buf, err := marshaler.Marshal(body)
	if err != nil {
		app.Logger.Error("Failed to marshal error body,", zap.Error(err))
		return
	}
	if _, err := w.Write(buf); err != nil {
		app.Logger.Error("Failed to write response.,", zap.Error(err))
	}
}

func (app *App) startHTTPServer(ctx context.Context, lis net.Listener) error {
	gatewayMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{MarshalOptions: protojson.MarshalOptions{EmitUnpopulated: true}}),
		runtime.WithErrorHandler(app.applicationErrorHandler),
		runtime.WithIncomingHeaderMatcher(customHeadersMatcher),
	)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
	}

	if err := api.RegisterPodiumHandlerFromEndpoint(ctx, gatewayMux, app.GRPCEndpoint, opts); err != nil {
		return fmt.Errorf("error registering multiplexer for grpc gateway: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", removeTrailingSlashMiddleware{addVersionMiddleware{gatewayMux}})
	mux.HandleFunc("/healthcheck", addVersionHandlerFunc(app.healthCheckHandler))
	mux.HandleFunc("/status", addVersionHandlerFunc(app.statusHandler))
	attachSpan := func(span opentracing.Span, r *http.Request) {
		_ = opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	}

	muxWithTracing := nethttp.Middleware(opentracing.GlobalTracer(), mux, nethttp.MWSpanObserver(attachSpan))

	app.httpServer = &http.Server{
		Addr:    app.HTTPEndpoint,
		Handler: muxWithTracing,
	}

	app.httpReady <- true
	if err := app.httpServer.Serve(lis); err != http.ErrServerClosed {
		return fmt.Errorf("error listening and serving http requests: %v", err)
	}

	return nil
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

// GracefullStop attempts to stop the server.
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

func customHeadersMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case "tenant-id":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
