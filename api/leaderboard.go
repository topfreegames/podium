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
func (app *App) BulkUpsertScores(ctx context.Context, in *api.BulkUpsertScoresRequest) (*api.MemberListResponse, error) {
	if err := validateBulkUpsertScoresRequest(in); err != nil {
		return nil, err
	}

	lg := app.Logger.With(
		zap.String("handler", "BulkUpsertScores"),
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

	return &api.MemberListResponse{Success: true, Members: responses}, nil
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

// GetRank is the handler responsible for retrieving a member rank
func (app *App) GetRank(ctx context.Context, in *api.GetRankRequest) (*api.GetRankResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetRank"),
		zap.String("leaderboard", in.LeaderboardId),
		zap.String("memberPublicID", in.MemberPublicId),
	)

	order := in.Order
	if order == "" || (order != "asc" && order != "desc") {
		order = "desc"
	}

	var rank int
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting rank.")
		rank, err = app.Leaderboards.GetRank(ctx, in.LeaderboardId, in.MemberPublicId, order)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			lg.Error("Member not found.", zap.Error(err))
			app.AddError()
			return status.New(codes.NotFound, "Member not found.").Err()
		} else if err != nil {
			lg.Error("Getting rank failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Getting rank succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.GetRankResponse{
		Success:  true,
		PublicID: in.MemberPublicId,
		Rank:     int32(rank),
	}, nil
}

//GetRankMultiLeaderboards returns the member rank in several leaderboards at once
func (app *App) GetRankMultiLeaderboards(ctx context.Context, in *api.MultiGetRankRequest) (*api.MultiGetRankResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetRankMultiLeaderboards"),
		zap.String("memberPublicID", in.MemberPublicId),
	)

	order := in.Order
	if order == "" || (order != "asc" && order != "desc") {
		order = "desc"
	}

	if in.LeaderboardIds == "" {
		app.AddError()
		return nil, status.New(codes.InvalidArgument,
			"Leaderboard IDs are required using the 'leaderboardIds' querystring parameter").Err()
	}

	leaderboardIDs := strings.Split(in.LeaderboardIds, ",")
	serializedScores := make([]*api.MultiGetRankResponse_Member, len(leaderboardIDs))

	err := withSegment("Model", ctx, func() error {
		for i, leaderboardID := range leaderboardIDs {
			lg.Debug("Getting member rank on leaderboard.", zap.String("leaderboard", leaderboardID))
			member, err := app.Leaderboards.GetMember(ctx, leaderboardID, in.MemberPublicId, order, in.ScoreTTL)
			if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
				lg.Error("Member not found.", zap.Error(err))
				app.AddError()
				return status.New(codes.NotFound, "Leaderboard not found or member not found in leaderboard.").Err()
			} else if err != nil {
				lg.Error("Getting member rank on leaderboard failed.", zap.Error(err))
				app.AddError()
				return err
			}
			lg.Debug("Getting member rank on leaderboard succeeded.")
			serializedScores[i] = &api.MultiGetRankResponse_Member{
				LeaderboardID: leaderboardID,
				Rank:          int32(member.Rank),
				Score:         float64(member.Score),
				IntScore:      member.Score,
				ExpireAt:      int32(member.ExpireAt),
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.MultiGetRankResponse{
		Success: true,
		Scores:  serializedScores,
	}, nil
}

// GetAroundMember retrieves a list of member score and rank centered in the given member
func (app *App) GetAroundMember(ctx context.Context, in *api.GetAroundMemberRequest) (*api.MemberListResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetAroundMember"),
		zap.String("leaderboard", in.LeaderboardId),
		zap.String("memberPublicID", in.MemberPublicId),
	)

	order := in.Order
	if order == "" || (order != "asc" && order != "desc") {
		order = "desc"
	}

	pageSize := app.getPageSize(int(in.PageSize))
	if pageSize > app.Config.GetInt("api.maxReturnedMembers") {
		msg := fmt.Sprintf(
			"Max pageSize allowed: %d. pageSize requested: %d",
			app.Config.GetInt("api.maxReturnedMembers"),
			pageSize,
		)
		return nil, status.New(codes.InvalidArgument, msg).Err()
	}

	var members leaderboard.Members
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting members around player.")
		members, err = app.Leaderboards.GetAroundMe(ctx, in.LeaderboardId, pageSize, in.MemberPublicId, order,
			in.GetLastIfNotFound)
		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			lg.Error("Member not found.", zap.Error(err))
			app.AddError()
			return status.New(codes.NotFound, "Member not found.").Err()
		} else if err != nil {
			lg.Error("Getting members around player failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Getting members around player succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.MemberListResponse{
		Success: true,
		Members: newMemberResponseList(members),
	}, nil
}

func (app *App) getPageSize(pageSize int) int {
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	return pageSize
}

// GetAroundScore retrieves a list of member scores and ranks centered in a given score
func (app *App) GetAroundScore(ctx context.Context, in *api.GetAroundScoreRequest) (*api.MemberListResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetAroundScoreHandler"),
		zap.String("leaderboard", in.LeaderboardId),
	)

	order := in.Order
	if order == "" || (order != "asc" && order != "desc") {
		order = "desc"
	}

	pageSize := app.getPageSize(int(in.PageSize))
	if pageSize > app.Config.GetInt("api.maxReturnedMembers") {
		msg := fmt.Sprintf(
			"Max pageSize allowed: %d. pageSize requested: %d",
			app.Config.GetInt("api.maxReturnedMembers"),
			pageSize,
		)
		return nil, status.New(codes.InvalidArgument, msg).Err()
	}

	var members leaderboard.Members
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting players around score.", zap.Int64("score", in.Score))
		members, err = app.Leaderboards.GetAroundScore(ctx, in.LeaderboardId, pageSize, in.Score, order)
		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			lg.Error("Member not found.", zap.Error(err))
			app.AddError()
			return status.New(codes.NotFound, "Member not found.").Err()
		} else if err != nil {
			lg.Error("Getting players around score failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Getting players around score succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.MemberListResponse{
		Success: true,
		Members: newMemberResponseList(members),
	}, nil
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

// GetTopMembers retrieves onePage of member score and rank
func (app *App) GetTopMembers(ctx context.Context, in *api.GetTopMembersRequest) (*api.MemberListResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetTopMembers"),
		zap.String("leaderboard", in.LeaderboardId),
	)

	if in.PageNumber == 0 {
		in.PageNumber = 1
	}

	order := in.Order
	if order == "" || (order != "asc" && order != "desc") {
		order = "desc"
	}

	pageSize := app.getPageSize(int(in.PageSize))
	if pageSize > app.Config.GetInt("api.maxReturnedMembers") {
		msg := fmt.Sprintf(
			"Max pageSize allowed: %d. pageSize requested: %d",
			app.Config.GetInt("api.maxReturnedMembers"),
			pageSize,
		)
		return nil, status.New(codes.InvalidArgument, msg).Err()
	}

	var members leaderboard.Members
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting top members.")
		members, err = app.Leaderboards.GetLeaders(ctx, in.LeaderboardId, pageSize, int(in.PageNumber), order)

		if err != nil {
			lg.Error("Getting top members failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Getting top members succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.MemberListResponse{
		Success: true,
		Members: newMemberResponseList(members),
	}, nil
}

// GetTopPercentage retrieves top x % members in the leaderboard
func (app *App) GetTopPercentage(ctx context.Context, in *api.GetTopPercentageRequest) (*api.MemberListResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetTopPercentage"),
		zap.String("leaderboard", in.LeaderboardId),
	)

	if in.Percentage == 0 {
		app.AddError()
		return nil, status.New(codes.InvalidArgument, "Percentage must be a valid integer between 1 and 100.").Err()
	}

	order := in.Order
	if order == "" || (order != "asc" && order != "desc") {
		order = "desc"
	}

	var members leaderboard.Members
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting top percentage.", zap.Int("percentage", int(in.Percentage)))
		members, err = app.Leaderboards.GetTopPercentage(ctx, in.LeaderboardId, defaultPageSize,
			int(in.Percentage), app.Config.GetInt("api.maxReturnedMembers"), order)

		if err != nil {
			lg.Error("Getting top percentage failed.", zap.Error(err))
			if err.Error() == "Percentage must be a valid integer between 1 and 100." {
				app.AddError()
				return status.New(codes.InvalidArgument, err.Error()).Err()
			}

			app.AddError()
			return err
		}
		lg.Debug("Getting top percentage succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.MemberListResponse{
		Success: true,
		Members: newMemberResponseList(members),
	}, nil
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

// UpsertScoreAllLeaderboards sets the member score for all leaderboards
func (app *App) UpsertScoreMultiLeaderboards(ctx context.Context, in *api.MultiUpsertScoreRequest) (*api.MultiUpsertScoreResponse, error) {
	if len(in.ScoreMultiChange.Leaderboards) == 0 {
		return nil, status.New(codes.InvalidArgument, "leaderboards is required").Err()
	}

	lg := app.Logger.With(
		zap.String("handler", "UpsertScoreAllLeaderboards"),
		zap.String("memberPublicID", in.MemberPublicId),
	)

	serializedScores := make([]*api.MultiUpsertScoreResponse_Member, len(in.ScoreMultiChange.Leaderboards))

	err := withSegment("Model", ctx, func() error {
		for i, leaderboardID := range in.ScoreMultiChange.Leaderboards {
			lg.Debug("Updating score.",
				zap.String("leaderboardID", leaderboardID),
				zap.Int64("score", in.ScoreMultiChange.Score))
			member, err := app.Leaderboards.SetMemberScore(ctx, leaderboardID, in.MemberPublicId,
				in.ScoreMultiChange.Score, in.PrevRank, strconv.Itoa(int(in.ScoreTTL)))

			if err != nil {
				lg.Error("Update score failed.", zap.Error(err))
				app.AddError()
				return err
			}
			serializedScore := &api.MultiUpsertScoreResponse_Member{
				PublicID:      member.PublicID,
				Score:         float64(member.Score),
				IntScore:      member.Score,
				Rank:          int32(member.Rank),
				PreviousRank:  int32(member.PreviousRank),
				ExpireAt:      int32(member.ExpireAt),
				LeaderboardID: leaderboardID,
			}
			serializedScores[i] = serializedScore
		}
		lg.Debug("Update score succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.MultiUpsertScoreResponse{
		Success: true,
		Scores:  serializedScores,
	}, nil
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
