package api

import (
	"testing"
)

func TestRemoveTrailingSlash(t *testing.T) {
	m := removeTrailingSlashMiddleware{}

	tests := []struct {
		input string
		want  string
	}{
		{"/", "/"},
		{"http://localhost", "http://localhost"},
		{"http://localhost:8900/l/invalid-leaderboard/members/", "http://localhost:8900/l/invalid-leaderboard/members"},
		{"http://localhost:8900/l/invalid-leaderboard/members", "http://localhost:8900/l/invalid-leaderboard/members"},
	}

	for _, test := range tests {
		if got := m.removeTrailingSlash(test.input); got != test.want {
			t.Errorf("removeTrailingSlash(%q) = %v, want %v", test.input, got, test.want)
		}
	}
}
