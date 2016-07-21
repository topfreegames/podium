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
	"strings"

	"github.com/kataras/iris"
	"github.com/topfreegames/podium/leaderboard"
	"github.com/uber-go/zap"
)

var notFoundError = "Could not find data for user"
var noPageSizeProvidedError = "strconv.ParseInt: parsing \"\": invalid syntax"
var defaultPageSize = 20

type setScorePayload struct {
	Score int
}

func serializeUser(user *leaderboard.User) map[string]interface{} {
	return map[string]interface{}{
		"publicID": user.PublicID,
		"score":    user.Score,
		"rank":     user.Rank,
	}
}

func serializeUsers(users []*leaderboard.User) []map[string]interface{} {
	serializedUsers := make([]map[string]interface{}, len(users))
	for i, user := range users {
		serializedUsers[i] = serializeUser(user)
	}
	return serializedUsers
}

// UpsertUserScoreHandler is the handler responsible for creating or updating the user score
func UpsertUserScoreHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "UpsertUserScoreHandler"),
		)
		leaderboardID := c.Param("leaderboardID")
		userPublicID := c.Param("userPublicID")

		var payload setScorePayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		user, err := l.SetUserScore(userPublicID, payload.Score)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(serializeUser(user), c)
	}
}

// RemoveUserHandler is the handler responsible for removing a user score
func RemoveUserHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "RemoveUserHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		userPublicID := c.Param("userPublicID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		_, err := l.RemoveMember(userPublicID)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}

// GetUserHandler is the handler responsible for retrieving a user score and rank
func GetUserHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetUserHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		userPublicID := c.Param("userPublicID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		user, err := l.GetMember(userPublicID)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			FailWith(404, "User not found.", c)
			return
		} else if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(serializeUser(user), c)
	}
}

// GetUserRankHandler is the handler responsible for retrieving a user rank
func GetUserRankHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetUserRankHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		userPublicID := c.Param("userPublicID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		rank, err := l.GetRank(userPublicID)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			FailWith(404, "User not found.", c)
			return
		} else if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"publicID": userPublicID,
			"rank":     rank,
		}, c)
	}
}

// GetAroundUserHandler retrieves a list of user score and rank centered in the given user
func GetAroundUserHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetAroundUserHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		userPublicID := c.Param("userPublicID")
		pageSize, err := c.URLParamInt("pageSize")
		if err != nil && err.Error() == noPageSizeProvidedError {
			pageSize = defaultPageSize
		} else if err != nil {
			FailWith(400, fmt.Sprintf("Invalid pageSize provided: %s", err.Error()), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, pageSize, lg)
		users, err := l.GetAroundMe(userPublicID)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			FailWith(404, "User not found.", c)
			return
		} else if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"users": serializeUsers(users),
		}, c)
	}
}

// GetTotalMembersHandler is the handler responsible for returning the total number of members in a leaderboard
func GetTotalMembersHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetTotalMembersHandler"),
		)

		leaderboardID := c.Param("leaderboardID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		count, err := l.TotalMembers()

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"count": count,
		}, c)
	}
}

// GetTotalPagesHandler is the handler responsible for returning the total number of pages in a leaderboard given a pageSize
func GetTotalPagesHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetTotalPagesHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		pageSize, err := c.URLParamInt("pageSize")
		if err != nil && err.Error() == noPageSizeProvidedError {
			pageSize = defaultPageSize
		} else if err != nil {
			FailWith(400, fmt.Sprintf("Invalid pageSize provided: %s", err.Error()), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, pageSize, lg)
		count, err := l.TotalPages()

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"count": count,
		}, c)
	}
}

// GetTopUsersHandler retrieves onePage of user score and rank
func GetTopUsersHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetTopUsersHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		pageNumber, err := c.ParamInt("pageNumber")
		if err != nil {
			FailWith(400, fmt.Sprintf("Invalid pageNumber provided: %s", err.Error()), c)
			return
		}
		pageSize, err := c.URLParamInt("pageSize")
		if err != nil && err.Error() == noPageSizeProvidedError {
			pageSize = defaultPageSize
		} else if err != nil {
			FailWith(400, fmt.Sprintf("Invalid pageSize provided: %s", err.Error()), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, pageSize, lg)
		users, err := l.GetLeaders(pageNumber)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			FailWith(404, "User not found.", c)
			return
		} else if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"users": serializeUsers(users),
		}, c)
	}
}

// GetTopPercentageHandler retrieves top x % players in the leaderboard
func GetTopPercentageHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetTopPercentageHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		percentage, err := c.ParamInt("percentage")
		if err != nil {
			FailWith(400, fmt.Sprintf("Invalid percentage provided: %s", err.Error()), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, defaultPageSize, lg)
		users, err := l.GetTopPercentage(percentage, app.Config.GetInt("api.maxReturnedMembers"))

		if err != nil {
			if err.Error() == "Percentage must be a valid integer between 1 and 100." {
				FailWith(400, err.Error(), c)
				return
			}

			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"users": serializeUsers(users),
		}, c)
	}
}
