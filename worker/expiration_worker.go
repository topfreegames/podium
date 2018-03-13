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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	extredis "github.com/topfreegames/extensions/redis"
	"go.uber.org/zap"
)

// ExpirationResult is the struct that represents the result of an expiration job
type ExpirationResult struct {
	DeletedMembers int
	DeletedSet     bool
	Set            string
}

// ExpirationWorker is the struct that represents the scores expirer worker
type ExpirationWorker struct {
	RedisClient             *extredis.Client
	Logger                  zap.Logger
	Config                  *viper.Viper
	ConfigPath              string
	ExpirationCheckInterval time.Duration
	ExpirationLimitPerRun   int
	running                 bool
	shouldRun               bool
}

// GetExpirationWorker returns a new scores expirer worker
func GetExpirationWorker(configPath string, logger zap.Logger) (*ExpirationWorker, error) {
	worker := &ExpirationWorker{
		ConfigPath: configPath,
		Logger:     logger,
		Config:     viper.New(),
	}
	err := worker.configure()
	if err != nil {
		return nil, err
	}
	return worker, nil
}

func (w *ExpirationWorker) configure() error {
	err := w.loadConfiguration()
	if err != nil {
		return err
	}
	w.setConfigurationDefaults()
	w.ExpirationCheckInterval = w.Config.GetDuration("worker.expirationCheckInterval")
	w.ExpirationLimitPerRun = w.Config.GetInt("worker.expirationLimitPerRun")

	l := w.Logger.With(
		zap.String("operation", "configureWorker"),
	)

	redisHost := w.Config.GetString("redis.host")
	redisPort := w.Config.GetInt("redis.port")
	redisPass := w.Config.GetString("redis.password")
	redisDB := w.Config.GetInt("redis.db")
	redisConnectionTimeout := w.Config.GetString("redis.connectionTimeout")

	redisURLObject := url.URL{
		Scheme: "redis",
		User:   url.UserPassword("", redisPass),
		Host:   fmt.Sprintf("%s:%d", redisHost, redisPort),
		Path:   fmt.Sprint(redisDB),
	}
	redisURL := redisURLObject.String()
	w.Config.Set("redis.url", redisURL)
	rl := l.With(
		zap.String("url", fmt.Sprintf("redis://:<REDACTED>@%s:%v/%v", redisHost, redisPort, redisDB)),
		zap.String("connectionTimeout", redisConnectionTimeout),
	)
	rl.Debug("Connecting to redis...")
	cli, err := extredis.NewClient("redis", w.Config)
	if err != nil {
		rl.Error("Failed to connect to redis")
		return err
	}
	w.RedisClient = cli
	rl.Info("Connected to redis successfully.")
	return nil
}

func (w *ExpirationWorker) loadConfiguration() error {
	w.Config.SetConfigFile(w.ConfigPath)
	w.Config.SetEnvPrefix("podium")
	w.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	w.Config.AutomaticEnv()

	if err := w.Config.ReadInConfig(); err == nil {
		w.Logger.Info("Loaded config file.", zap.String("configFile", w.Config.ConfigFileUsed()))
	} else {
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
	l := w.Logger.With(
		zap.String("operation", "expireScores"),
	)
	expirationSets, err := w.RedisClient.Client.SMembers("expiration-sets").Result()
	if err != nil {
		return nil, err
	}
	l.Debug("expiring scores", zap.Object("sets", expirationSets))
	for _, set := range expirationSets {
		setSplitted := strings.Split(set, ":")
		expirationTTLStr := setSplitted[len(setSplitted)-1]
		// remove :ttl:<expirationTTL> from the key
		leaderboardName := strings.Join(setSplitted[:len(setSplitted)-2], ":")
		expirationTTL, err := strconv.Atoi(expirationTTLStr)
		if err != nil {
			l.Error("error expiring scores", zap.Error(err))
			continue
		}
		l.Debug("expiring scores from set", zap.String("set", set), zap.Int("TTL", expirationTTL))
		maxTimeLastActivity := strconv.Itoa(int(time.Now().Unix()) - expirationTTL)

		expirationScript := w.getExpireScoresScript()
		result, err := expirationScript.Run(w.RedisClient.Client, []string{leaderboardName, set}, maxTimeLastActivity).Result()
		if err != nil {
			return nil, err
		}

		results := result.([]interface{})
		deletedMembersCount := results[0].(int64)
		deletedSet := false
		if results[1].(int64) > 0 {
			deletedSet = true
		}

		l.Info("expired members from set",
			zap.String("leaderboardName", leaderboardName),
			zap.String("set", set),
			zap.Int("TTL", expirationTTL),
			zap.Int64("deletedMembersCount", deletedMembersCount),
			zap.Bool("deletedSet", deletedSet),
		)

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
func (w *ExpirationWorker) Run() {
	w.shouldRun = true
	stopChan := make(chan struct{})
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	w.running = false
	l := w.Logger.With(
		zap.String("operation", "Run"),
		zap.Duration("ExpirationCheckInterval", w.ExpirationCheckInterval),
	)
	l.Info("Running scores expirer worker")
	ticker := time.NewTicker(w.ExpirationCheckInterval)
	go func() {
		for range ticker.C {
			if w.shouldRun == false {
				close(stopChan)
				break
			}
			if w.running == false {
				w.running = true
				result, err := w.expireScores()
				if err != nil {
					l.Error("error expiring scores", zap.Error(err))
				}
				l.Debug("expiration results", zap.Object("result", result))
			}
		}
	}()
	for w.shouldRun == true {
		select {
		case <-sigChan:
			w.Logger.Warn("Scores expirer worker exiting...")
			w.shouldRun = false
			<-stopChan
		case <-stopChan:
			break
		}
	}
}
