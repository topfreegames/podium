// podium
// https://github.com/topfreegames/podium
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package bench

import (
	"fmt"
	"strings"
	"testing"

	uuid "github.com/satori/go.uuid"
)

var keeper interface{}

func BenchmarkSetMemberScore(b *testing.B) {
	lbID, _ := generateNMembers(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		memberID := uuid.NewV4().String()
		route := getRoute(fmt.Sprintf("/l/%s/members/%s/score", lbID, memberID))
		payload := map[string]interface{}{
			"score": 100,
		}

		status, body, err := putTo(route, payload)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkSetMembersScore(b *testing.B) {
	members := make([]map[string]interface{}, 50)
	lbID, _ := generateNMembers(b.N)
	for j := 0; j < 50; j++ {
		memberID := uuid.NewV4().String()
		members[j] = map[string]interface{}{"publicID": memberID, "score": int64(100)}
	}
	payload := map[string]interface{}{"members": members}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/l/%s/scores", lbID))

		status, body, err := putTo(route, payload)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkIncrementMemberScore(b *testing.B) {
	lbID, _ := generateNMembers(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		memberID := uuid.NewV4().String()
		route := getRoute(fmt.Sprintf("/l/%s/members/%s/score", lbID, memberID))
		payload := map[string]interface{}{
			"increment": 100,
		}

		status, body, err := patchTo(route, payload)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkRemoveMember(b *testing.B) {
	lbID, _ := generateNMembers(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		memberID := fmt.Sprintf("bench-member-%d", i)
		route := getRoute(fmt.Sprintf("/l/%s/members?ids=%s", lbID, memberID))
		status, body, err := sendTo("DELETE", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetMember(b *testing.B) {
	lbID, _ := generateNMembers(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		memberID := fmt.Sprintf("bench-member-%d", i)
		route := getRoute(fmt.Sprintf("/l/%s/members/%s", lbID, memberID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetMemberRank(b *testing.B) {
	lbID, _ := generateNMembers(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		memberID := fmt.Sprintf("bench-member-%d", i)
		route := getRoute(fmt.Sprintf("/l/%s/members/%s/rank", lbID, memberID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetAroundMember(b *testing.B) {
	lbID, _ := generateNMembers(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		memberID := fmt.Sprintf("bench-member-%d", i)
		route := getRoute(fmt.Sprintf("/l/%s/members/%s/around", lbID, memberID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetTotalMembers(b *testing.B) {
	lbID, _ := generateNMembers(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/l/%s/members-count", lbID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetTopMembers(b *testing.B) {
	lbID, _ := generateNMembers(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/l/%s/top/1", lbID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGetTopPercentage(b *testing.B) {
	lbID, _ := generateNMembers(b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/l/%s/top-percent/10", lbID))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkSetMemberScoreForSeveralLeaderboards(b *testing.B) {
	leaderboards := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		lbID, _ := generateNMembers(b.N)
		leaderboards[i] = lbID
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		memberID := uuid.NewV4().String()
		route := getRoute(fmt.Sprintf("/m/%s/scores", memberID))
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

func BenchmarkGetMembers(b *testing.B) {
	lbID, _ := generateNMembers(b.N)
	memberIDs := []string{}
	for i := 0; i <= 500; i++ {
		memberID := fmt.Sprintf("bench-member-%d", i)
		memberIDs = append(memberIDs, memberID)
	}

	mIDs := strings.Join(memberIDs, ",")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route := getRoute(fmt.Sprintf("/l/%s/members?ids=%s", lbID, mIDs))
		status, body, err := sendTo("GET", route, nil)
		validateResp(status, body, err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}
