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
	"github.com/kataras/iris"
	"github.com/topfreegames/podium/leaderboard"
)

type setScorePayload struct {
	Score int
}

// UpsertUserScoreHandler is the handler responsible for creating or updating the user score
func UpsertUserScoreHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		leaderboardID := c.Param("leaderboardID")
		userPublicID := c.Param("userPublicID")

		var payload setScorePayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0)

		res, err := l.SetUserScore(
			userPublicID,
			payload.Score,
		)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"publicID": res.PublicID,
			"score":    res.Score,
			"rank":     res.Rank,
		}, c)
	}
}
