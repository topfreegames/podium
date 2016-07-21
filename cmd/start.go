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
	"github.com/spf13/cobra"
	"github.com/topfreegames/podium/api"
	"github.com/uber-go/zap"
)

var host string
var port int
var debug bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the podium API server",
	Long: `Starts podium server with the specified arguments. You can use
	environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := zap.NewJSON(zap.InfoLevel).With(
			zap.String("source", "app"),
		)

		app := api.GetApp(
			host,
			port,
			ConfigFile,
			debug,
			logger,
		)

		app.Start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind podium to")
	startCmd.Flags().IntVarP(&port, "port", "p", 8890, "Port to bind podium to")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode")
}
