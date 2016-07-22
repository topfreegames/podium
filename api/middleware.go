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

	"github.com/getsentry/raven-go"
	"github.com/kataras/iris"
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
