// go-leaderboard
// https://github.com/topfreegames/go-leaderboard
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package api

import (
	"runtime/debug"

	"github.com/kataras/iris"
)

//VersionMiddleware automatically adds a version header to response
type VersionMiddleware struct {
	App *App
}

// Serve automatically adds a version header to response
func (m *VersionMiddleware) Serve(c *iris.Context) {
	c.SetHeader("GO-LEADERBOARD-VERSION", VERSION)
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
