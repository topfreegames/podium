package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"google.golang.org/grpc/metadata"
)

func TestBasicAuthMiddleware(t *testing.T) {
	app := &App{}
	app.Config = viper.New()

	tests := []struct {
		user          string
		pwd           string
		auth          string
		shouldSucceed bool
	}{
		{"user", "pwd", "Basic dXNlcjpwd2Q=", true},
		{"user", "pwd", "Basic dXNlcjpwd2Q", false},
		{"user", "pwd", "dXNlcjpwd2Q", false},
	}

	for _, test := range tests {
		ctx := context.Background()
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", fmt.Sprintf("%s", test.auth)))

		app.Config.Set("basicauth.username", test.user)
		app.Config.Set("basicauth.password", test.pwd)

		_, err := app.basicAuthMiddleware(ctx)

		if (err == nil && test.shouldSucceed == false) || (err != nil && test.shouldSucceed == true) {
			t.Errorf("basicAuthMiddleware(%s, %s) = %v, should succeed is %v", test.user, test.pwd, err,
				test.shouldSucceed)
		}
	}
}

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
