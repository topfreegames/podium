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
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/topfreegames/podium/api"
	"github.com/topfreegames/podium/log"
	"go.uber.org/zap"
)

var host string
var httpPort int
var grpcPort int
var debug, quiet bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the podium API server",
	Long: `Starts podium server with the specified arguments. You can use
	environment variables to override configuration keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		ll := zap.InfoLevel
		if debug {
			ll = zap.DebugLevel
		}
		if quiet {
			ll = zap.WarnLevel
		}

		logger := log.CreateLoggerWithLevel(ll, log.LoggerOptions{WriteSyncer: os.Stdout})
		logger = logger.With(
			zap.String("source", "app"),
		)

		defer logger.Sync()

		app, err := api.New(
			host,
			httpPort,
			grpcPort,
			ConfigFile,
			debug,
			logger,
		)

		if err != nil {
			logger.Fatal("Could not get podium application.", zap.Error(err))
		}

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		err = app.Start(ctx)
		if err != nil {
			logger.Fatal("Could not start podium application.", zap.Error(err))
		}
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind podium to")
	startCmd.Flags().IntVarP(&httpPort, "http_port", "p", 8880, "HTTP Port to bind podium to")
	startCmd.Flags().IntVarP(&grpcPort, "grpc_port", "g", 8881, "GRPC Port to bind podium to")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode (log=debug)")
	startCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (log=warn)")
}
