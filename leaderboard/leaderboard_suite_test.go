package leaderboard_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLeaderboard(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Leaderboard Suite")
}
