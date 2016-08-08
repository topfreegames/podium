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

var notFoundError = "Could not find data for member"
var noPageSizeProvidedError = "strconv.ParseInt: parsing \"\": invalid syntax"
var defaultPageSize = 20

type setScorePayload struct {
	Score int
}

type setScoresPayload struct {
	Score        int
	Leaderboards []string
}

func serializeMember(member *leaderboard.Member, position int) map[string]interface{} {
	memberData := map[string]interface{}{
		"publicID": member.PublicID,
		"score":    member.Score,
		"rank":     member.Rank,
	}
	if position >= 0 {
		memberData["position"] = position
	}
	return memberData
}

func serializeMembers(members []*leaderboard.Member, includePosition bool) []map[string]interface{} {
	serializedMembers := make([]map[string]interface{}, len(members))
	for i, member := range members {
		if includePosition {
			serializedMembers[i] = serializeMember(member, i)
		} else {
			serializedMembers[i] = serializeMember(member, -1)
		}
	}
	return serializedMembers
}

// UpsertMemberScoreHandler is the handler responsible for creating or updating the member score
func UpsertMemberScoreHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "UpsertMemberScoreHandler"),
		)
		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")

		var payload setScorePayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			app.AddError()
			FailWith(400, err.Error(), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		member, err := l.SetMemberScore(memberPublicID, payload.Score)

		if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(serializeMember(member, -1), c)
	}
}

func RemoveMemberHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "RemoveMemberHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		err := l.RemoveMember(memberPublicID)

		if err != nil && !strings.HasPrefix(err.Error(), notFoundError) {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}

func RemoveMembersHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "RemoveMembersHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		ids := c.URLParam("ids")

		if ids == "" {
			app.AddError()
			FailWith(400, "Member IDs are required using the 'ids' querystring parameter", c)
			return
		}

		memberIDs := strings.Split(ids, ",")
		idsInter := make([]interface{}, len(memberIDs))
		for i, v := range memberIDs {
			idsInter[i] = v
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		err := l.RemoveMembers(idsInter)

		if err != nil && !strings.HasPrefix(err.Error(), notFoundError) {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}

// GetMemberHandler is the handler responsible for retrieving a member score and rank
func GetMemberHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetMemberHandler"),
		)

		order := c.URLParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		member, err := l.GetMember(memberPublicID, order)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			app.AddError()
			FailWith(404, "Member not found.", c)
			return
		} else if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(serializeMember(member, -1), c)
	}
}

// GetMemberRankHandler is the handler responsible for retrieving a member rank
func GetMemberRankHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetMemberRankHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")
		order := c.URLParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		rank, err := l.GetRank(memberPublicID, order)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			app.AddError()
			FailWith(404, "Member not found.", c)
			return
		} else if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"publicID": memberPublicID,
			"rank":     rank,
		}, c)
	}
}

func GetMemberRankInManyLeaderboardsHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetMemberRankInManyLeaderboardsHandler"),
		)
		memberPublicID := c.Param("memberPublicID")
		ids := c.URLParam("leaderboardIds")
		order := c.URLParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		if ids == "" {
			app.AddError()
			FailWith(400, "Leaderboard IDs are required using the 'leaderboardIds' querystring parameter", c)
			return
		}

		leaderboardIds := strings.Split(ids, ",")
		serializedScores := make([]map[string]interface{}, len(leaderboardIds))

		for i, leaderboardId := range leaderboardIds {
			l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardId, 0, lg)
			member, err := l.GetMember(memberPublicID, order)
			if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
				app.AddError()
				FailWith(404, "Leaderboard not found or member not found in leaderboard.", c)
				return
			} else if err != nil {
				app.AddError()
				FailWith(500, err.Error(), c)
				return
			}
			serializedScores[i] = map[string]interface{}{
				"leaderboardID": leaderboardId,
				"rank":          member.Rank,
				"score":         member.Score,
			}
		}

		SucceedWith(map[string]interface{}{
			"scores": serializedScores,
		}, c)
	}
}

// GetAroundMemberHandler retrieves a list of member score and rank centered in the given member
func GetAroundMemberHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetAroundMemberHandler"),
		)

		order := c.URLParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")
		pageSize, err := c.URLParamInt("pageSize")
		if err != nil && err.Error() == noPageSizeProvidedError {
			pageSize = defaultPageSize
		} else if err != nil {
			app.AddError()
			FailWith(400, fmt.Sprintf("Invalid pageSize provided: %s", err.Error()), c)
			return
		} else if pageSize > app.Config.GetInt("api.maxReturnedMembers") {
			app.AddError()
			FailWith(400, fmt.Sprintf("Max pageSize allowed: %d. pageSize requested: %d", app.Config.GetInt("api.maxReturnedMembers"), pageSize), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, pageSize, lg)
		members, err := l.GetAroundMe(memberPublicID, order)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			app.AddError()
			FailWith(404, "Member not found.", c)
			return
		} else if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"members": serializeMembers(members, false),
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
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"count": count,
		}, c)
	}
}

// GetTopMembersHandler retrieves onePage of member score and rank
func GetTopMembersHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetTopMembersHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		pageNumber, err := c.ParamInt("pageNumber")
		order := c.URLParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		if err != nil {
			app.AddError()
			FailWith(400, fmt.Sprintf("Invalid pageNumber provided: %s", err.Error()), c)
			return
		}
		pageSize, err := c.URLParamInt("pageSize")
		if err != nil && err.Error() == noPageSizeProvidedError {
			pageSize = defaultPageSize
		} else if err != nil {
			app.AddError()
			FailWith(400, fmt.Sprintf("Invalid pageSize provided: %s", err.Error()), c)
			return
		} else if pageSize > app.Config.GetInt("api.maxReturnedMembers") {
			app.AddError()
			FailWith(400, fmt.Sprintf("Max pageSize allowed: %d. pageSize requested: %d", app.Config.GetInt("api.maxReturnedMembers"), pageSize), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, pageSize, lg)
		members, err := l.GetLeaders(pageNumber, order)

		if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"members": serializeMembers(members, false),
		}, c)
	}
}

// GetTopPercentageHandler retrieves top x % members in the leaderboard
func GetTopPercentageHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetTopPercentageHandler"),
		)

		order := c.URLParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		leaderboardID := c.Param("leaderboardID")
		percentage, err := c.ParamInt("percentage")
		if err != nil {
			app.AddError()
			FailWith(400, fmt.Sprintf("Invalid percentage provided: %s", err.Error()), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, defaultPageSize, lg)
		members, err := l.GetTopPercentage(percentage, app.Config.GetInt("api.maxReturnedMembers"), order)

		if err != nil {
			if err.Error() == "Percentage must be a valid integer between 1 and 100." {
				app.AddError()
				FailWith(400, err.Error(), c)
				return
			}

			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"members": serializeMembers(members, false),
		}, c)
	}
}

// GetMembersHandler retrieves several members at once
func GetMembersHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetMembersHandler"),
		)

		order := c.URLParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		leaderboardID := c.Param("leaderboardID")
		ids := c.URLParam("ids")
		if ids == "" {
			app.AddError()
			FailWith(400, "Member IDs are required using the 'ids' querystring parameter", c)
			return
		}

		memberIDs := strings.Split(ids, ",")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, defaultPageSize, lg)
		members, err := l.GetMembers(memberIDs, order)

		if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		notFound := []string{}

		for _, memberID := range memberIDs {
			found := false
			for _, member := range members {
				if member.PublicID == memberID {
					found = true
					break
				}
			}
			if !found {
				notFound = append(notFound, memberID)
			}
		}

		SucceedWith(map[string]interface{}{
			"members":  serializeMembers(members, true),
			"notFound": notFound,
		}, c)
	}
}

// UpsertMemberLeaderboardsScoreHandler sets the member score for all leaderboards
func UpsertMemberLeaderboardsScoreHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "UpsertMemberLeaderboardsScoreHandler"),
		)
		memberPublicID := c.Param("memberPublicID")

		var payload setScoresPayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		serializedScores := make([]map[string]interface{}, len(payload.Leaderboards))

		for i, leaderboardID := range payload.Leaderboards {
			l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
			member, err := l.SetMemberScore(memberPublicID, payload.Score)

			if err != nil {
				app.AddError()
				FailWith(500, err.Error(), c)
				return
			}
			serializedScore := serializeMember(member, -1)
			serializedScore["leaderboardID"] = leaderboardID
			serializedScores[i] = serializedScore
		}

		SucceedWith(map[string]interface{}{
			"scores": serializedScores,
		}, c)
	}
}

// RemoveLeaderboardHandler is the handler responsible for removing a leaderboard
func RemoveLeaderboardHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "RemoveLeaderboardHandler"),
		)
		leaderboardID := c.Param("leaderboardID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		err := l.RemoveLeaderboard()

		if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}
