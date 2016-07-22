// podium
// https://github.com/topfreegames/podium
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package bench

import (
	"fmt"
	"testing"

	"github.com/satori/go.uuid"
)

var keeper interface{}

func BenchmarkSetUserScore(b *testing.B) {
	l := generateNUsers(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userID := uuid.NewV4().String()
		route := getRoute(fmt.Sprintf("/l/%s/users/%s/score", l.PublicID, userID))
		payload := map[string]interface{}{
			"score": 100,
		}

		status, body, err := putTo(route, payload)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkRemoveUser(b *testing.B) {
	l := generateNUsers(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userID := fmt.Sprintf("bench-user-%d", i)
		route := getRoute(fmt.Sprintf("/l/%s/users/%s", l.PublicID, userID))
		status, body, err := sendTo("DELETE", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetUser(b *testing.B) {
	l := generateNUsers(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userID := fmt.Sprintf("bench-user-%d", i)
		route := getRoute(fmt.Sprintf("/l/%s/users/%s", l.PublicID, userID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetUserRank(b *testing.B) {
	l := generateNUsers(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userID := fmt.Sprintf("bench-user-%d", i)
		route := getRoute(fmt.Sprintf("/l/%s/users/%s/rank", l.PublicID, userID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetAroundUser(b *testing.B) {
	l := generateNUsers(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userID := fmt.Sprintf("bench-user-%d", i)
		route := getRoute(fmt.Sprintf("/l/%s/users/%s/around", l.PublicID, userID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetTotalMembers(b *testing.B) {
	l := generateNUsers(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/l/%s/users-count", l.PublicID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetTotalPages(b *testing.B) {
	l := generateNUsers(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/l/%s/pages", l.PublicID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetTopUsers(b *testing.B) {
	l := generateNUsers(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/l/%s/top/1", l.PublicID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetTopPercentage(b *testing.B) {
	l := generateNUsers(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/l/%s/top-percent/10", l.PublicID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkSetUserScoreForSeveralLeaderboards(b *testing.B) {
	leaderboards := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		l := generateNUsers(b.N)
		leaderboards[i] = l.PublicID
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		userID := uuid.NewV4().String()
		route := getRoute(fmt.Sprintf("/u/%s/scores", userID))
		payload := map[string]interface{}{
			"score":        100,
			"leaderboards": leaderboards,
		}

		status, body, err := putTo(route, payload)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}
