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
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/topfreegames/podium/api"
	"github.com/uber-go/zap"
)

var host string
var port int
var debug bool
var quiet bool
var profile bool
var profilePath string

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
		logger := zap.NewJSON(ll).With(
			zap.String("source", "app"),
		)

		app, err := api.GetApp(
			host,
			port,
			ConfigFile,
			debug,
			profile,
			logger,
		)

		if err != nil {
			logger.Fatal("Could not start podium application.", zap.Error(err))
		}

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func(l zap.Logger) {
			sig := <-sigs
			l.Info("Received terminate signal. Cleaning up...", zap.String("signal", fmt.Sprintf("%v", sig)))
			cleanUpAndExit(l)
		}(logger)

		app.Start()
	},
}

func cleanUpAndExit(logger zap.Logger) {
	logger.Info("Cleaned up successfully. Exiting...", zap.Int("exitCode", 0))
	os.Exit(0)
}

func writeCPUProfile(logger zap.Logger) {
	logger.Debug("Stopping cpu profiler...")
	pprof.StopCPUProfile()
	logger.Info("CPU profiling finished successfully")
}

func writeHeapProfile(logger zap.Logger) {
	logger.Debug("Running memory profiler...")
	f, err := os.Create(path.Join(profilePath, "podium.mem.p"))
	if err != nil {
		logger.Fatal("Could not create memory profile.", zap.Error(err))
	}
	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("Could not write memory profile.", zap.Error(err))
	}
	f.Close()
	logger.Info("Memory profiling finished successfully")
}

func init() {
	RootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&host, "bind", "b", "0.0.0.0", "Host to bind podium to")
	startCmd.Flags().IntVarP(&port, "port", "p", 8880, "Port to bind podium to")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Debug mode (log=debug)")
	startCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (log=warn)")
	startCmd.Flags().BoolVarP(&profile, "profile", "s", false, "Profile mode")
	startCmd.Flags().StringVarP(&profilePath, "profile-path", "f", "/tmp/podium-profile", "Profile path to store profile files")
}
