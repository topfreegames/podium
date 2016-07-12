// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva
package cmd

import (
	"io"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/spf13/cobra"
)

var out io.Writer = os.Stdout

func Test(t *testing.T) {
	Describe("Root Cmd", func() {
		It("Should run command", func() {
			var rootCmd = &cobra.Command{
				Use:   "podium",
				Short: "podium handles redis backed leaderboards",
				Long:  `podium handles redis backed leaderboards.`,
			}
			Execute(rootCmd)
		})
	})
}
