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
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	extredis "github.com/topfreegames/extensions/redis"
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
	RedisClient             *extredis.Client
	Config                  *viper.Viper
	ConfigPath              string
	ExpirationCheckInterval time.Duration
	ExpirationLimitPerRun   int
	running                 bool
	shouldRun               bool
}

// GetExpirationWorker returns a new scores expirer worker
func GetExpirationWorker(configPath string) (*ExpirationWorker, error) {
	worker := &ExpirationWorker{
		ConfigPath: configPath,
		Config:     viper.New(),
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
func NewExpirationWorker(host string, port int, password string, db int, connectionTimeout int,
	expirationCheckInterval time.Duration, expirationLimitPerRun int) (*ExpirationWorker, error) {

	config := viper.New()

	config.Set("redis.host", host)
	config.Set("redis.port", port)
	config.Set("redis.password", password)
	config.Set("redis.db", db)
	config.Set("redis.connectionTimeout", connectionTimeout)
	config.Set("worker.expirationCheckInterval", expirationCheckInterval)
	config.Set("worker.expirationLimitPerRun", expirationLimitPerRun)

	worker := &ExpirationWorker{
		Config: config,
	}

	err := worker.configure()
	if err != nil {
		return nil, err
	}

	return worker, nil
}

func (w *ExpirationWorker) configure() error {
	w.setConfigurationDefaults()
	w.ExpirationCheckInterval = w.Config.GetDuration("worker.expirationCheckInterval")
	w.ExpirationLimitPerRun = w.Config.GetInt("worker.expirationLimitPerRun")

	redisHost := w.Config.GetString("redis.host")
	redisPort := w.Config.GetInt("redis.port")
	redisPass := w.Config.GetString("redis.password")
	redisDB := w.Config.GetInt("redis.db")

	redisURLObject := url.URL{
		Scheme: "redis",
		User:   url.UserPassword("", redisPass),
		Host:   fmt.Sprintf("%s:%d", redisHost, redisPort),
		Path:   fmt.Sprint(redisDB),
	}
	redisURL := redisURLObject.String()
	w.Config.Set("redis.url", redisURL)
	cli, err := extredis.NewClient("redis", w.Config)
	if err != nil {
		return fmt.Errorf("Failed to connect to redis: %v", err)
	}
	w.RedisClient = cli
	return nil
}

func (w *ExpirationWorker) loadConfiguration() error {
	w.Config.SetConfigFile(w.ConfigPath)
	w.Config.SetEnvPrefix("podium")
	w.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	w.Config.AutomaticEnv()

	if err := w.Config.ReadInConfig(); err != nil {
		return fmt.Errorf("Could not load configuration file from: %s", w.ConfigPath)
	}

	return nil

}

func (w *ExpirationWorker) setConfigurationDefaults() {
	w.Config.SetDefault("redis.host", "localhost")
	w.Config.SetDefault("redis.port", 1212)
	w.Config.SetDefault("redis.password", "")
	w.Config.SetDefault("redis.db", 0)
	w.Config.SetDefault("redis.maxPoolSize", 20)
	w.Config.SetDefault("redis.connectionTimeout", 200)
	w.Config.SetDefault("worker.expirationCheckInterval", "60s")
	w.Config.SetDefault("worker.expirationLimitPerRun", "1000")
}

func (w *ExpirationWorker) getExpireScoresScript() *redis.Script {
	return redis.NewScript(fmt.Sprintf(`
		-- Script params:
		-- KEYS[1] is the name of the leaderboard
		-- KEYS[2] is the name of the expiration set
		-- ARGV[1] is the deadline of activity

		local deleted_set = 0
		local deleted_members = 0
		local leaderboard_name = KEYS[1]
		local expiration_set = KEYS[2]
		local activity_deadline = ARGV[1]
		local exists = redis.call("EXISTS", expiration_set)
		if exists == 0 then
			deleted_set = 1
			redis.call("SREM", "expiration-sets", expiration_set)
		end
		if deleted_set == 0 then
			local members_to_remove = redis.call("ZRANGEBYSCORE", expiration_set, "-inf", activity_deadline, "LIMIT", 0, %d)
			if #members_to_remove > 0 then
				-- unpack is limited to the stack size, if we pass a lot of members here, this function will fail,
				-- that being said, it is better to use low expiration check interval
				redis.call("ZREM", expiration_set, unpack(members_to_remove))
				deleted_members = redis.call("ZREM", leaderboard_name, unpack(members_to_remove))
			end
		end
		return {deleted_members, deleted_set}
	`, w.ExpirationLimitPerRun))
}

func (w *ExpirationWorker) expireScores() ([]*ExpirationResult, error) {
	defer func() {
		w.running = false
	}()
	res := make([]*ExpirationResult, 0)
	expirationSets, err := w.RedisClient.Client.SMembers("expiration-sets").Result()
	if err != nil {
		return nil, err
	}
	for _, set := range expirationSets {
		ttlSuffix := ":ttl"
		if !(strings.HasSuffix(set, ttlSuffix)) {
			continue
		}
		leaderboardName := strings.TrimSuffix(set, ttlSuffix)

		expirationScript := w.getExpireScoresScript()
		result, err := expirationScript.Run(w.RedisClient.Client, []string{leaderboardName, set}, time.Now().Unix()).Result()
		if err != nil {
			return nil, err
		}

		results := result.([]interface{})
		deletedMembersCount := results[0].(int64)
		deletedSet := false
		if results[1].(int64) > 0 {
			deletedSet = true
		}

		res = append(res, &ExpirationResult{
			Set:            set,
			DeletedMembers: int(deletedMembersCount),
			DeletedSet:     deletedSet,
		})
	}
	return res, nil
}

// Stop stops
func (w *ExpirationWorker) Stop() {
	w.shouldRun = false
}

// Run starts the worker -- this method blocks
func (w *ExpirationWorker) Run(resultsChan chan<- []*ExpirationResult, errChan chan<- error) {
	w.shouldRun = true
	stopChan := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	w.running = false
	ticker := time.NewTicker(w.ExpirationCheckInterval)
	go func() {
		for range ticker.C {
			if !w.shouldRun {
				close(stopChan)
				break
			}
			if !w.running {
				w.running = true
				results, err := w.expireScores()
				if err != nil {
					errChan <- fmt.Errorf("error expiring scores: %v", err)
					continue
				}
				resultsChan <- results
			}
		}
	}()
	for w.shouldRun {
		select {
		case <-sigChan:
			w.shouldRun = false
			<-stopChan
		case <-stopChan:
			break
		}
	}
}
