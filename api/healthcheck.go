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
	"fmt"
	"strings"

	"github.com/kataras/iris"
)

// HealthCheckHandler is the handler responsible for validating that the app is still up
func HealthCheckHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		workingString := app.Config.GetString("healthcheck.workingText")
		res, err := app.RedisClient.GetConnection().Do("ping")
		if res != "PONG" || err != nil {
			c.Write(fmt.Sprintf("Error connecting to redis: %s", err))
			c.SetStatusCode(500)
			return
		}

		c.SetStatusCode(iris.StatusOK)
		workingString = strings.TrimSpace(workingString)
		c.Write(workingString)
	}
}
