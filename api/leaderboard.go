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
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"github.com/buger/jsonparser"
	"github.com/labstack/echo"
	"github.com/topfreegames/podium/leaderboard"
	"go.uber.org/zap"

	api "github.com/topfreegames/podium/proto/podium/api/v1"
)

var notFoundError = "Could not find data for member"
var noPageSizeProvidedError = "strconv.ParseInt: parsing \"\": invalid syntax"
var defaultPageSize = 20

func serializeMember(member *leaderboard.Member, position int, includeTTL bool) map[string]interface{} {
	memberData := map[string]interface{}{
		"publicID": member.PublicID,
		"score":    member.Score,
		"rank":     member.Rank,
	}
	if member.PreviousRank != 0 {
		memberData["previousRank"] = member.PreviousRank
	}
	if position >= 0 {
		memberData["position"] = position
	}
	if includeTTL {
		memberData["expireAt"] = member.ExpireAt
	}
	return memberData
}

func serializeMembers(members leaderboard.Members, includePosition bool, includeTTL bool) []map[string]interface{} {
	serializedMembers := make([]map[string]interface{}, len(members))
	for i, member := range members {
		if includePosition {
			serializedMembers[i] = serializeMember(member, i, includeTTL)
		} else {
			serializedMembers[i] = serializeMember(member, -1, includeTTL)
		}
	}
	return serializedMembers
}

func validateBulkUpsertScoresRequest(in *api.BulkUpsertScoresRequest) error {
	for _, m := range in.ScoreUpserts.Members {
		if m.PublicID == "" {
			return status.New(codes.InvalidArgument, "publicID is required").Err()
		}
	}
	return nil
}

func newDefaultMemberResponse(member *leaderboard.Member) *api.DefaultMemberResponse {
	return &api.DefaultMemberResponse{
		Success:      true,
		PublicID:     member.PublicID,
		Score:        float64(member.Score),
		IntScore:     member.Score,
		Rank:         int32(member.Rank),
		PreviousRank: int32(member.PreviousRank),
		ExpireAt:     int32(member.ExpireAt),
	}
}

// BulkUpsertMembersScoreHandler is the handler responsible for creating or updating members score
func (app *App) BulkUpsertScores(ctx context.Context, in *api.BulkUpsertScoresRequest) (*api.BulkUpsertScoresResponse, error) {
	if err := validateBulkUpsertScoresRequest(in); err != nil {
		return nil, err
	}

	lg := app.Logger.With(
		zap.String("handler", "BulkUpsertMembersScoreHandler"),
		zap.String("leaderboard", in.LeaderboardId),
	)

	members := make(leaderboard.Members, len(in.ScoreUpserts.Members))

	err := withSegment("Model", ctx, func() error {
		lg.Debug("Setting member scores.")
		for i, ms := range in.ScoreUpserts.Members {
			members[i] = &leaderboard.Member{Score: ms.Score, PublicID: ms.PublicID}
		}
		err := app.Leaderboards.SetMembersScore(ctx, in.LeaderboardId, members, in.PrevRank, in.ScoreTTL)

		if err != nil {
			lg.Error("Setting member scores failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Setting member scores succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	responses := make([]*api.MemberResponse, len(in.ScoreUpserts.Members))

	for i, m := range members {
		responses[i] = &api.MemberResponse{
			PublicID:     m.PublicID,
			Score:        float64(m.Score),
			IntScore:     m.Score,
			Rank:         int32(m.Rank),
			PreviousRank: int32(m.PreviousRank),
			ExpireAt:     int32(m.ExpireAt),
		}
	}

	return &api.BulkUpsertScoresResponse{Success: true, Members: responses}, nil
}

// UpsertScore is the handler responsible for creating or updating the member score
func (app *App) UpsertScore(ctx context.Context, in *api.UpsertScoreRequest) (*api.DefaultMemberResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "UpsertScore"),
		zap.String("leaderboard", in.LeaderboardId),
		zap.String("memberPublicID", in.MemberPublicId),
	)

	var member *leaderboard.Member
	err := withSegment("Model", ctx, func() error {
		lg.Debug("Setting member score.", zap.Int64("score", in.ScoreChange.Score))

		var err error
		member, err = app.Leaderboards.SetMemberScore(ctx, in.LeaderboardId, in.MemberPublicId, in.ScoreChange.Score,
			in.PrevRank, strconv.Itoa(int(in.ScoreTTL)))

		if err != nil {
			lg.Error("Setting member score failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Setting member score succeeded.")
		return nil
	})

	if err != nil {
		return nil, err
	}

	return newDefaultMemberResponse(member), nil
}

// IncrementMemberScore is the handler responsible for incrementing the member score
func (app *App) IncrementScore(ctx context.Context, in *api.IncrementScoreRequest) (*api.DefaultMemberResponse, error) {
	if in.Body.Increment == 0 {
		return nil, status.New(codes.InvalidArgument, "increment is required").Err()
	}

	lg := app.Logger.With(
		zap.String("handler", "IncrementScore"),
		zap.String("leaderboard", in.LeaderboardId),
		zap.String("memberPublicID", in.MemberPublicId),
	)

	var member *leaderboard.Member
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Incrementing member score.", zap.Int64("increment", in.Body.Increment))
		member, err = app.Leaderboards.IncrementMemberScore(context.Background(), in.LeaderboardId, in.MemberPublicId,
			int(in.Body.Increment), strconv.Itoa(int(in.ScoreTTL)))

		if err != nil {
			lg.Error("Member score increment failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Member score increment succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return newDefaultMemberResponse(member), nil
}

//RemoveMemberHandler removes a member from a leaderboard
func (app *App) RemoveMember(ctx context.Context, in *api.RemoveMemberRequest) (*api.BasicResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "RemoveMember"),
		zap.String("leaderboard", in.LeaderboardId),
		zap.String("memberPublicID", in.MemberPublicId),
	)

	err := withSegment("Model", ctx, func() error {
		lg.Debug("Removing member.")
		err := app.Leaderboards.RemoveMember(ctx, in.LeaderboardId, in.MemberPublicId)

		if err != nil && !strings.HasPrefix(err.Error(), notFoundError) {
			lg.Error("Member removal failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Member removal succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.BasicResponse{Success: true}, nil

}

//RemoveMembers removes several members from a leaderboard
func (app *App) RemoveMembers(ctx context.Context, in *api.RemoveMembersRequest) (*api.BasicResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "RemoveMembers"),
		zap.String("leaderboard", in.LeaderboardId),
	)

	if in.Ids == "" {
		app.AddError()
		return nil, status.New(codes.InvalidArgument,
			"Member IDs are required using the 'ids' querystring parameter").Err()
	}

	memberIDs := strings.Split(in.Ids, ",")
	idsInter := make([]interface{}, len(memberIDs))
	for i, v := range memberIDs {
		idsInter[i] = v
	}

	err := withSegment("Model", ctx, func() error {
		lg.Debug("Removing members.", zap.String("ids", in.Ids))
		err := app.Leaderboards.RemoveMembers(ctx, in.LeaderboardId, idsInter)

		if err != nil && !strings.HasPrefix(err.Error(), notFoundError) {
			lg.Error("Members removal failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Members removal succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.BasicResponse{Success: true}, nil
}

// GetMember is the handler responsible for retrieving a member score and rank
func (app *App) GetMember(ctx context.Context, in *api.GetMemberRequest) (*api.DefaultMemberResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetMember"),
		zap.String("leaderboard", in.LeaderboardId),
		zap.String("memberPublicID", in.MemberPublicId),
	)

	order := in.Order
	if order == "" || (order != "asc" && order != "desc") {
		order = "desc"
	}

	var member *leaderboard.Member
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting member.")
		member, err = app.Leaderboards.GetMember(ctx, in.LeaderboardId, in.MemberPublicId, order, in.ScoreTTL)
		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			lg.Error("Member not found.", zap.Error(err))
			app.AddError()
			return status.New(codes.NotFound, "Member not found.").Err()
		} else if err != nil {
			lg.Error("Get member failed.")
			app.AddError()
			return err
		}
		lg.Debug("Getting member succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return newDefaultMemberResponse(member), nil
}

// GetMemberRankHandler is the handler responsible for retrieving a member rank
func GetMemberRankHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")
		lg := app.Logger.With(
			zap.String("handler", "GetMemberRankHandler"),
			zap.String("leaderboard", leaderboardID),
			zap.String("memberPublicID", memberPublicID),
		)

		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		status := 404
		rank := 0
		err := WithSegment("Model", c, func() error {
			var err error
			lg.Debug("Getting rank.")
			rank, err = app.Leaderboards.GetRank(c.StdContext(), leaderboardID, memberPublicID, order)

			if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
				lg.Error("Member not found.", zap.Error(err))
				app.AddError()
				status = 404
				return fmt.Errorf("Member not found.")
			} else if err != nil {
				lg.Error("Getting rank failed.", zap.Error(err))
				app.AddError()
				status = 500
				return err
			}
			lg.Debug("Getting rank succeeded.")
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"publicID": memberPublicID,
			"rank":     rank,
		}, c)
	}
}

//GetMemberRankInManyLeaderboardsHandler returns the member rank in several leaderboards at once
func GetMemberRankInManyLeaderboardsHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		memberPublicID := c.Param("memberPublicID")
		lg := app.Logger.With(
			zap.String("handler", "GetMemberRankInManyLeaderboardsHandler"),
			zap.String("memberPublicID", memberPublicID),
		)

		ids := c.QueryParam("leaderboardIds")
		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}
		scoreTTL := c.QueryParam("scoreTTL") == "true"

		if ids == "" {
			app.AddError()
			return FailWith(400, "Leaderboard IDs are required using the 'leaderboardIds' querystring parameter", c)
		}

		leaderboardIDs := strings.Split(ids, ",")
		serializedScores := make([]map[string]interface{}, len(leaderboardIDs))

		status := 404
		err := WithSegment("Model", c, func() error {
			for i, leaderboardID := range leaderboardIDs {
				lg.Debug("Getting member rank on leaderboard.", zap.String("leaderboard", leaderboardID))
				member, err := app.Leaderboards.GetMember(c.StdContext(), leaderboardID, memberPublicID, order, scoreTTL)
				if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
					lg.Error("Member not found.", zap.Error(err))
					app.AddError()
					status = 404
					return fmt.Errorf("Leaderboard not found or member not found in leaderboard.")
				} else if err != nil {
					lg.Error("Getting member rank on leaderboard failed.", zap.Error(err))
					app.AddError()
					status = 500
					return err
				}
				lg.Debug("Getting member rank on leaderboard succeeded.")
				serializedScores[i] = map[string]interface{}{
					"leaderboardID": leaderboardID,
					"rank":          member.Rank,
					"score":         member.Score,
				}
				if scoreTTL {
					serializedScores[i]["expireAt"] = member.ExpireAt
				}
			}
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"scores": serializedScores,
		}, c)
	}
}

// GetAroundMemberHandler retrieves a list of member score and rank centered in the given member
func GetAroundMemberHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")
		lg := app.Logger.With(
			zap.String("handler", "GetAroundMemberHandler"),
			zap.String("leaderboard", leaderboardID),
			zap.String("memberPublicID", memberPublicID),
		)

		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}
		getLastIfNotFound := false
		getLastIfNotFoundStr := c.QueryParam("getLastIfNotFound")
		if getLastIfNotFoundStr == "true" {
			getLastIfNotFound = true
		}

		pageSize, err := GetPageSize(app, c, defaultPageSize)
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		var members leaderboard.Members
		status := 404
		err = WithSegment("Model", c, func() error {
			lg.Debug("Getting members around player.")
			members, err = app.Leaderboards.GetAroundMe(c.StdContext(), leaderboardID, pageSize, memberPublicID, order,
				getLastIfNotFound)
			if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
				lg.Error("Member not found.", zap.Error(err))
				app.AddError()
				status = 404
				return fmt.Errorf("Member not found.")
			} else if err != nil {
				lg.Error("Getting members around player failed.", zap.Error(err))
				app.AddError()
				status = 500
				return err
			}
			lg.Debug("Getting members around player succeeded.")
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"members": serializeMembers(members, false, false),
		}, c)
	}
}

// GetAroundScoreHandler retrieves a list of member scores and ranks centered in a given score
func GetAroundScoreHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		leaderboardID := c.Param("leaderboardID")
		lg := app.Logger.With(
			zap.String("handler", "GetAroundScoreHandler"),
			zap.String("leaderboard", leaderboardID),
		)

		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		score, err := strconv.ParseInt(c.Param("score"), 10, 64)
		if err != nil {
			return FailWith(400, "Score not sent or wrongly formatted", c)
		}

		pageSize, err := GetPageSize(app, c, defaultPageSize)
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		var members leaderboard.Members
		status := 404
		err = WithSegment("Model", c, func() error {
			lg.Debug("Getting players around score.", zap.Int64("score", score))
			members, err = app.Leaderboards.GetAroundScore(c.StdContext(), leaderboardID, pageSize, score, order)
			if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
				lg.Error("Member not found.", zap.Error(err))
				app.AddError()
				status = 404
				return fmt.Errorf("Member not found.")
			} else if err != nil {
				lg.Error("Getting players around score failed.", zap.Error(err))
				app.AddError()
				status = 500
				return err
			}
			lg.Debug("Getting players around score succeeded.")
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"members": serializeMembers(members, false, false),
		}, c)
	}
}

// TotalMembers is the handler responsible for returning the total number of members in a leaderboard
func (app *App) TotalMembers(ctx context.Context, in *api.TotalMembersRequest) (*api.TotalMembersResponse, error) {
	leaderboardID := in.LeaderboardId
	lg := app.Logger.With(
		zap.String("handler", "TotalMembers"),
		zap.String("leaderboard", leaderboardID),
	)

	var count int
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting total members.")
		count, err = app.Leaderboards.TotalMembers(ctx, leaderboardID)

		if err != nil {
			lg.Error("Getting total members failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Getting total members succeeded.")
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &api.TotalMembersResponse{Success: true, Count: int32(count)}, nil
}

// GetTopMembersHandler retrieves onePage of member score and rank
func GetTopMembersHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		leaderboardID := c.Param("leaderboardID")
		lg := app.Logger.With(
			zap.String("handler", "GetTopMembersHandler"),
			zap.String("leaderboard", leaderboardID),
		)

		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		pageNumber, err := GetIntRouteParam(app, c, "pageNumber", 1)
		if err != nil {
			app.AddError()
			return FailWith(400, err.Error(), c)
		}

		pageSize, err := GetPageSize(app, c, defaultPageSize)
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		var members leaderboard.Members
		err = WithSegment("Model", c, func() error {
			lg.Debug("Getting top members.")
			members, err = app.Leaderboards.GetLeaders(c.StdContext(), leaderboardID, pageSize, pageNumber, order)

			if err != nil {
				lg.Error("Getting top members failed.", zap.Error(err))
				app.AddError()
				return err
			}
			lg.Debug("Getting top members succeeded.")
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"members": serializeMembers(members, false, false),
		}, c)
	}
}

// GetTopPercentageHandler retrieves top x % members in the leaderboard
func GetTopPercentageHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		leaderboardID := c.Param("leaderboardID")
		lg := app.Logger.With(
			zap.String("handler", "GetTopPercentageHandler"),
			zap.String("leaderboard", leaderboardID),
		)

		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		percentageStr := c.Param("percentage")
		percentage, err := strconv.ParseInt(percentageStr, 10, 32)
		if err != nil {
			app.AddError()
			return FailWith(400, fmt.Sprintf("Invalid percentage provided: %s", err.Error()), c)
		}
		if percentage == 0 {
			app.AddError()
			return FailWith(400, "Percentage must be a valid integer between 1 and 100.", c)
		}

		var members leaderboard.Members
		status := 400
		err = WithSegment("Model", c, func() error {
			lg.Debug("Getting top percentage.", zap.Int64("percentage", percentage))
			members, err = app.Leaderboards.GetTopPercentage(c.StdContext(), leaderboardID, defaultPageSize,
				int(percentage), app.Config.GetInt("api.maxReturnedMembers"), order)

			if err != nil {
				lg.Error("Getting top percentage failed.", zap.Error(err))
				if err.Error() == "Percentage must be a valid integer between 1 and 100." {
					app.AddError()
					status = 400
					return err
				}

				app.AddError()
				status = 500
				return err
			}
			lg.Debug("Getting top percentage succeeded.")
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"members": serializeMembers(members, false, false),
		}, c)
	}
}

// GetMembers retrieves several members at once
func (app *App) GetMembers(ctx context.Context, in *api.GetMembersRequest) (*api.GetMembersResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetMembers"),
		zap.String("leaderboard", in.LeaderboardId),
	)

	order := in.Order
	if order == "" || (order != "asc" && order != "desc") {
		order = "desc"
	}

	if in.Ids == "" {
		app.AddError()
		return nil, status.New(codes.InvalidArgument,
			"Member IDs are required using the 'ids' querystring parameter").Err()
	}

	memberIDs := strings.Split(in.Ids, ",")

	var members leaderboard.Members
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting members.", zap.String("ids", in.Ids))
		members, err = app.Leaderboards.GetMembers(ctx, in.LeaderboardId, memberIDs, order, in.ScoreTTL)

		if err != nil {
			lg.Error("Getting members failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Getting members succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	var notFound []string

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

	return &api.GetMembersResponse{
		Success:  true,
		Members:  newMemberResponseList(members),
		NotFound: notFound,
	}, nil
}

func newMemberResponseList(members leaderboard.Members) []*api.MemberResponse {
	list := make([]*api.MemberResponse, len(members))
	for i, m := range members {
		list[i] = &api.MemberResponse{
			PublicID:     m.PublicID,
			Score:        float64(m.Score),
			IntScore:     m.Score,
			Rank:         int32(m.Rank),
			PreviousRank: int32(m.PreviousRank),
			ExpireAt:     int32(m.ExpireAt),
			Position:     int32(i),
		}
	}
	return list
}

// UpsertMemberLeaderboardsScoreHandler sets the member score for all leaderboards
func UpsertMemberLeaderboardsScoreHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		memberPublicID := c.Param("memberPublicID")
		lg := app.Logger.With(
			zap.String("handler", "UpsertMemberLeaderboardsScoreHandler"),
			zap.String("memberPublicID", memberPublicID),
		)

		scoreTTL := c.QueryParam("scoreTTL")

		var payload setScoresPayload

		prevRank := false
		prevRankStr := c.QueryParam("prevRank")
		if prevRankStr != "" && prevRankStr == "true" {
			prevRank = true
		}

		err := WithSegment("Payload", c, func() error {
			b, err := GetRequestBody(c)
			if err != nil {
				app.AddError()
				return err
			}
			if _, err := jsonparser.GetInt(b, "score"); err != nil {
				app.AddError()
				return fmt.Errorf("score is required")
			}
			if err := LoadJSONPayload(&payload, c, lg); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		serializedScores := make([]map[string]interface{}, len(payload.Leaderboards))

		err = WithSegment("Model", c, func() error {
			for i, leaderboardID := range payload.Leaderboards {
				lg.Debug("Updating score.",
					zap.String("leaderboardID", leaderboardID),
					zap.Int64("score", payload.Score))
				member, err := app.Leaderboards.SetMemberScore(c.StdContext(), leaderboardID, memberPublicID,
					payload.Score, prevRank, scoreTTL)

				if err != nil {
					lg.Error("Update score failed.", zap.Error(err))
					app.AddError()
					return err
				}
				serializedScore := serializeMember(member, -1, scoreTTL != "")
				serializedScore["leaderboardID"] = leaderboardID
				serializedScores[i] = serializedScore
			}
			lg.Debug("Update score succeeded.")
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"scores": serializedScores,
		}, c)
	}
}

// RemoveLeaderboard is the handler responsible for removing a leaderboard
func (app *App) RemoveLeaderboard(ctx context.Context, in *api.RemoveLeaderboardRequest) (*api.BasicResponse, error) {
	leaderboardID := in.LeaderboardId
	lg := app.Logger.With(
		zap.String("handler", "RemoveLeaderboard"),
		zap.String("leaderboard", leaderboardID),
	)

	err := withSegment("Model", ctx, func() error {
		lg.Debug("Removing leaderboard.")
		err := app.Leaderboards.RemoveLeaderboard(ctx, leaderboardID)

		if err != nil {
			lg.Error("Remove leaderboard failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Remove leaderboard succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.BasicResponse{Success: true}, nil
}
