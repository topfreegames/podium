// podium
// https://github.com/topfreegames/podium
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	"github.com/topfreegames/podium/config"
	"github.com/topfreegames/podium/leaderboard/v2/database/redis"
)

var currentStage int
var stages map[int]string

var leaderboardCount = flag.Int("leaderboards", 3, "number of leaderboards to create")
var membersPerLeaderboard = flag.Int("mpl", 5000000, "number of members per leaderboard")

func main() {
	flag.Parse()

	start := time.Now().Unix()

	totalOps := *leaderboardCount * *membersPerLeaderboard

	config, err := config.GetDefaultConfig("../../default.yaml")
	if err != nil {
		panic(err)
	}

	uiprogress.Start()                     // start rendering
	bar := uiprogress.AddBar(totalOps - 1) // Add a new bar
	bar.AppendCompleted()
	bar.PrependElapsed()

	// prepend the deploy step to the bar
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		ellapsed := time.Now().Unix() - start
		itemsPerSec := float64(b.Current()+1) / float64(ellapsed) / 1000
		timeToComplete := float64(totalOps) / itemsPerSec / 60.0 / 60.0
		text := fmt.Sprintf("%d/%d (%.2fhs to complete)", b.Current()+1, totalOps, timeToComplete)
		return strutil.Resize(text, uint(len(text)))
	})

	client := redis.NewStandaloneClient(redis.StandaloneOptions{
		Host:     config.GetString("redis.host"),
		Port:     config.GetInt("redis.port"),
		Password: config.GetString("redis.password"),
		DB:       config.GetInt("redis.db"),
	})

	createTestData(client, *leaderboardCount, *membersPerLeaderboard, bar.Incr)
}

func createTestData(cli redis.Redis, leaderboardCount, membersPerLeaderboard int, progress func() bool) error {
	for i := 0; i < leaderboardCount; i++ {
		for j := 0; j < membersPerLeaderboard; j++ {
			setScore(cli, fmt.Sprintf("leaderboard-%d", i), fmt.Sprintf("member-%d", j), i*j)
			progress()
		}
	}

	return nil
}

func setScore(cli redis.Redis, leaderboard, member string, score int) {
	err := cli.ZAdd(context.Background(), leaderboard, &redis.Member{Score: float64(score), Member: member})
	if err != nil {
		panic(err)
	}
}
