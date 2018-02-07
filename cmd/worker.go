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
	"github.com/topfreegames/podium/worker"
	"go.uber.org/zap"
)

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "starts the podium scores expirer worker",
	Long: `starts the podium worker that expires scores with the specified arguments. you can use environment variables to override
	configuration keys`,
	Run: func(cmd *cobra.Command, args []string) {
		ll := zap.InfoLevel
		if debug {
			ll = zap.DebugLevel
		}
		if quiet {
			ll = zap.WarnLevel
		}
		logger := zap.New(
			zap.NewJSONEncoder(),
			ll,
		).With(
			zap.String("source", "worker"),
		)
		logger.Info("Starting podium score expirer worker...")

		w, err := worker.GetExpirationWorker(ConfigFile, logger)

		if err != nil {
			logger.Fatal("Could not get podium worker.", zap.Error(err))
		}

		w.Run()
	},
}

func init() {
	workerCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode (log=debug)")
	workerCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (log=warn)")
	RootCmd.AddCommand(workerCmd)
}
