// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/topfreegames/podium/config"
	"github.com/topfreegames/podium/leaderboard/v2/database"
)

// ExpirationResult is the struct that represents the result of an expiration job
type ExpirationResult struct {
	DeletedMembers int
	DeletedSet     bool
	Set            string
}

func (r *ExpirationResult) String() string {
	return fmt.Sprintf("(DeletedMembers: %d, DeletedSet: %t, Set: %s)", r.DeletedMembers, r.DeletedSet, r.Set)
}

// ExpirationWorker is the struct that represents the scores expirer worker
type ExpirationWorker struct {
	Config                  *viper.Viper
	Database                database.Expiration
	ConfigPath              string
	ExpirationCheckInterval time.Duration
	ExpirationLimitPerRun   int
	stop                    chan bool
}

// GetExpirationWorker returns a new scores expirer worker
func GetExpirationWorker(configPath string) (*ExpirationWorker, error) {
	worker := &ExpirationWorker{
		ConfigPath: configPath,
	}

	err := worker.loadConfiguration()
	if err != nil {
		return nil, err
	}

	err = worker.configure()
	if err != nil {
		return nil, err
	}

	return worker, nil
}

// NewExpirationWorker returns a new scores expirer worker with already loaded configuration.
func NewExpirationWorker(host string, port int, password string, db int,
	expirationCheckInterval time.Duration, expirationLimitPerRun int) (*ExpirationWorker, error) {

	worker := &ExpirationWorker{
		ConfigPath: "../config/default.yaml",
	}

	err := worker.loadConfiguration()
	if err != nil {
		return nil, err
	}

	err = worker.configure()
	if err != nil {
		return nil, err
	}

	return worker, nil
}

func (w *ExpirationWorker) loadConfiguration() error {
	config, err := config.GetDefaultConfig(w.ConfigPath)
	if err != nil {
		return err
	}
	w.Config = config
	return nil

}

func (w *ExpirationWorker) configure() error {
	w.setConfigurationDefaults()
	w.ExpirationCheckInterval = w.Config.GetDuration("worker.expirationCheckInterval")
	w.ExpirationLimitPerRun = w.Config.GetInt("worker.expirationLimitPerRun")
	w.stop = make(chan bool, 1)

	database := database.NewRedisDatabase(database.RedisOptions{
		ClusterEnabled: w.Config.GetBool("redis.cluster.enabled"),
		Addrs:          w.Config.GetStringSlice("redis.addrs"),
		Host:           w.Config.GetString("redis.host"),
		Port:           w.Config.GetInt("redis.port"),
		Password:       w.Config.GetString("redis.password"),
		DB:             w.Config.GetInt("redis.db"),
	})
	w.Database = database
	return nil
}

func (w *ExpirationWorker) setConfigurationDefaults() {
	w.Config.SetDefault("redis.clusterEnabled", "false")
	w.Config.SetDefault("redis.addrs", "")
	w.Config.SetDefault("redis.host", "localhost")
	w.Config.SetDefault("redis.port", "6379")
	w.Config.SetDefault("redis.password", "")
	w.Config.SetDefault("redis.db", 0)
	w.Config.SetDefault("redis.maxPoolSize", 20)
	w.Config.SetDefault("worker.expirationCheckInterval", "60s")
	w.Config.SetDefault("worker.expirationLimitPerRun", "1000")
}

// Stop finish expiration worker execution
func (w *ExpirationWorker) Stop() {
	w.stop <- true
}

// Run execute a new worker
func (w *ExpirationWorker) Run(resultsChan chan<- []*ExpirationResult, errChan chan<- error) {
	shouldEnd := make(chan bool, 1)
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go w.runWorker(shouldEnd, resultsChan, errChan)

	select {
	case <-sigChan:
		shouldEnd <- true
	case <-w.stop:
		shouldEnd <- true
	}

	close(sigChan)
	close(shouldEnd)
	close(w.stop)
}

func (w *ExpirationWorker) runWorker(shouldEnd chan bool, resultsChan chan<- []*ExpirationResult, errChan chan<- error) {
	ticker := time.NewTicker(w.ExpirationCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-shouldEnd:
			return
		case <-ticker.C:
			w.expireMembers(resultsChan, errChan)
		}
	}

}

func (w *ExpirationWorker) expireMembers(resultsChan chan<- []*ExpirationResult, errChan chan<- error) {
	leaderboardExpirations, err := w.Database.GetExpirationLeaderboards(context.Background())
	if err != nil {
		errChan <- err
	}

	result := []*ExpirationResult{}
	for _, leaderboard := range leaderboardExpirations {
		expirationResult, err := w.expireMembersFromLeaderboard(leaderboard)
		if err != nil {
			errChan <- err
			return
		}

		result = append(result, expirationResult)
	}
	resultsChan <- result
}

func (w *ExpirationWorker) expireMembersFromLeaderboard(leaderboard string) (*ExpirationResult, error) {
	members, err := w.Database.GetMembersToExpire(context.Background(), leaderboard, w.ExpirationLimitPerRun, time.Now().UTC())
	if err != nil {
		if _, ok := err.(*database.LeaderboardWithoutMemberToExpireError); ok {
			err = w.Database.RemoveLeaderboardFromExpireList(context.Background(), leaderboard)
			if err != nil {
				return nil, err
			}

			return &ExpirationResult{
				DeletedMembers: 0,
				DeletedSet:     true,
				Set:            leaderboard,
			}, nil
		}
		return nil, err
	}

	if len(members) == 0 {
		return &ExpirationResult{
			DeletedMembers: 0,
			DeletedSet:     false,
			Set:            leaderboard,
		}, nil
	}

	err = w.Database.ExpireMembers(context.Background(), leaderboard, members)
	if err != nil {
		return nil, err
	}

	return &ExpirationResult{
		DeletedMembers: len(members),
		DeletedSet:     false,
		Set:            leaderboard,
	}, nil
}
