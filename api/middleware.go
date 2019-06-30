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
	"encoding/base64"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/getsentry/raven-go"
	"github.com/topfreegames/podium/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type newRelicContextKey struct {
	key string
}

func (app *App) noAuthMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (app *App) basicAuthMiddleware(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "basic")
	if err != nil {
		return nil, err
	}

	auth := app.Config.GetString("basicauth.username") + ":" + app.Config.GetString("basicauth.password")

	if token != base64.StdEncoding.EncodeToString([]byte(auth)) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
	}
	return ctx, nil
}

func (app *App) loggerMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	l := app.Logger.With(
		zap.String("source", "request"),
	)

	//all except latency to string
	var statusCode int
	var latency time.Duration
	var startTime, endTime time.Time

	startTime = time.Now()

	h, err := handler(ctx, req)

	//no time.Since in order to format it well after
	endTime = time.Now()
	latency = endTime.Sub(startTime)

	_, statusCode = app.getStatusCodeFromError(err)

	method := info.FullMethod
	reqLog := l.With(
		zap.String("method", method),
		zap.Time("endTime", endTime),
		zap.Int("statusCode", statusCode),
		zap.Duration("latency", latency),
	)

	//request failed
	if statusCode > 399 && statusCode < 500 {
		log.D(reqLog, "Request failed.")
		return h, err
	}

	//request is ok, but server failed
	if statusCode > 499 {
		log.D(reqLog, "Response failed.")
		return h, err
	}

	//Everything went ok
	if cm := reqLog.Check(zap.DebugLevel, "Request successful."); cm.OK() {
		cm.Write()
	}

	return h, err
}

//Serve executes on error handler when errors happen
func (app *App) recoveryMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			eError, ok := err.(error)
			if !ok {
				eError = fmt.Errorf(fmt.Sprintf("%v", err))
			}
			app.OnErrorHandler(eError, debug.Stack())
		}
	}()
	return handler(ctx, req)
}

func (app *App) responseTimeMetricsMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	startTime := time.Now()
	h, err := handler(ctx, req)
	_, status := app.getStatusCodeFromError(err)
	method := info.FullMethod

	timeUsed := time.Since(startTime)

	tags := []string{
		fmt.Sprintf("method:%s", method),
		fmt.Sprintf("status:%d", status),
	}

	if err := app.DDStatsD.Timing("response_time_milliseconds", timeUsed, tags...); err != nil {
		app.Logger.Error("DDStatsD Timing", zap.Error(err))
	}

	return h, err
}

type addVersionMiddleware struct {
	Handler http.Handler
}

func addVersionHeaders(w http.ResponseWriter) {
	w.Header().Set("Server", fmt.Sprintf("Podium/v%s", VERSION))
	w.Header().Set("Podium-Server", fmt.Sprintf("Podium/v%s", VERSION))
}

func (m addVersionMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	addVersionHeaders(w)
	m.Handler.ServeHTTP(w, r)
}

func addVersionHandlerFunc(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		addVersionHeaders(w)
		f(w, r)
	}
}

func (app *App) sentryMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	h, err := handler(ctx, req)
	if err != nil {
		_, statusCode := app.getStatusCodeFromError(err)
		if statusCode < 500 {
			return h, err
		}

		tags := map[string]string{
			"source": "app",
			"type":   "Internal server error",
			"method": info.FullMethod,
			"params": fmt.Sprintf("%v", req),
			"status": fmt.Sprintf("%d", statusCode),
		}
		raven.CaptureError(err, tags)
	}
	return h, err
}

func (app *App) newRelicMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	txn := app.NewRelic.StartTransaction(info.FullMethod, nil, nil)
	newCtx := context.WithValue(ctx, newRelicContextKey{"txn"}, txn)
	defer func() {
		txn.End()
	}()
	h, err := handler(newCtx, req)
	if err != nil {
		txn.NoticeError(err)
	}
	return h, err
}

type removeTrailingSlashMiddleware struct {
	Handler http.Handler
}

func (m *removeTrailingSlashMiddleware) removeTrailingSlash(path string) string {
	l := len(path) - 1
	if l > 0 && path != "/" && path[l] == '/' {
		return path[:l]
	}
	return path
}

func (m removeTrailingSlashMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = m.removeTrailingSlash(r.URL.Path)
	m.Handler.ServeHTTP(w, r)
}
