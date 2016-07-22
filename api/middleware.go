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
	"runtime/debug"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/kataras/iris"
	"github.com/uber-go/zap"
)

//VersionMiddleware automatically adds a version header to response
type VersionMiddleware struct {
	App *App
}

// Serve automatically adds a version header to response
func (m *VersionMiddleware) Serve(c *iris.Context) {
	c.SetHeader("PODIUM-VERSION", VERSION)
	c.Next()
}

//RecoveryMiddleware recovers from errors in Iris
type RecoveryMiddleware struct {
	OnError func(interface{}, []byte)
}

//Serve executes on error handler when errors happen
func (r RecoveryMiddleware) Serve(ctx *iris.Context) {
	defer func() {
		if err := recover(); err != nil {
			if r.OnError != nil {
				r.OnError(err, debug.Stack())
			}
			ctx.Panic()
		}
	}()
	ctx.Next()
}

//LoggerMiddleware is responsible for logging to Zap all requests
type LoggerMiddleware struct {
	Logger zap.Logger
}

// Serve serves the middleware
func (l *LoggerMiddleware) Serve(ctx *iris.Context) {
	log := l.Logger.With(
		zap.String("source", "request"),
	)

	//all except latency to string
	var ip, method, path string
	var status int
	var latency time.Duration
	var startTime, endTime time.Time

	path = ctx.PathString()
	method = ctx.MethodString()

	startTime = time.Now()

	ctx.Next()

	//no time.Since in order to format it well after
	endTime = time.Now()
	latency = endTime.Sub(startTime)

	status = ctx.Response.StatusCode()
	ip = ctx.RemoteAddr()

	reqLog := log.With(
		zap.Time("endTime", endTime),
		zap.Int("statusCode", status),
		zap.Duration("latency", latency),
		zap.String("ip", ip),
		zap.String("method", method),
		zap.String("path", path),
	)

	//request failed
	if status > 399 && status < 500 {
		reqLog.Warn("Request failed.")
		return
	}

	//request is ok, but server failed
	if status > 499 {
		reqLog.Error("Response failed.")
		return
	}

	//Everything went ok
	reqLog.Info("Request successful.")
}

// NewLoggerMiddleware returns the logger middleware
func NewLoggerMiddleware(theLogger zap.Logger) iris.HandlerFunc {
	l := &LoggerMiddleware{Logger: theLogger}
	return l.Serve
}

//SentryMiddleware is responsible for sending all exceptions to sentry
type SentryMiddleware struct {
	App *App
}

// Serve serves the middleware
func (l *SentryMiddleware) Serve(ctx *iris.Context) {
	ctx.Next()

	if ctx.Response.StatusCode() > 499 {
		tags := map[string]string{
			"source": "app",
			"type":   "Internal server error",
			"url":    ctx.Request.URI().String(),
		}
		raven.CaptureError(fmt.Errorf("%s", string(ctx.Response.Body())), tags)
	}
}
