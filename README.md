# Podium

[![Podium](https://github.com/topfreegames/podium/actions/workflows/go.yml/badge.svg)](https://github.com/topfreegames/podium/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/topfreegames/podium/badge.svg?branch=master)](https://coveralls.io/github/topfreegames/podium?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/topfreegames/podium)](https://goreportcard.com/report/github.com/topfreegames/podium)
[![Docs](https://readthedocs.org/projects/podium/badge/?version=latest)](http://podium.readthedocs.io/en/latest/)
[![GoDoc](https://godoc.org/github.com/topfreegames/podium/leaderboard?status.svg)](https://godoc.org/github.com/topfreegames/podium/leaderboard)
[![](https://imagelayers.io/badge/tfgco/podium:latest.svg)](https://imagelayers.io/?images=tfgco/podium:latest 'Podium Image Layers')

A leaderboard system written in [Go](http://golang.org/) using [Redis](http://redis.io/) database. For more info, [read the docs](http://podium.readthedocs.io/en/latest/).

Features
--------

* **Multi-tenant** - Just vary the name of the leaderboard and you can have any number of tenants using leaderboards;
* **Seasonal Leaderboards** - Including suffixes like `year2016week01` or `year2016month06` is all you need to create seasonal leaders. I'm serious! That's all there is to it;
* **No leaderboard configuration** - Just start notifying scores for members of a leaderboard. There's no need to create, configure or maintain leaderboards. Let Podium do that for you;
* **Top Members** - Get the top members of a leaderboard whether you need by absolute value (top 200 members) or percentage (top 3% members);
* **Members around me** - Podium easily returns members around a specific member in the leaderboard. It will even compensate if you ask for the top member or last member to make sure you get a consistent amount of members;
* **Batch score update** - In a single operation, send a member score to many different leaderboards or many members score to the same leaderboard. This allows easy tracking of member rankings in several leaderboards at once (global, regional, clan, etc.);
* **Easy to deploy** - Podium comes with containers already exported to docker hub for every single of our successful builds. Just pick your choice!
* **Leaderboards with expiration** - If a player last update is older than (timeNow - X seconds), delete it from the leaderboard;
* **Use as library** - You can use podium as a library as well, adding leaderboard functionality directly to your application;

Installation
------------

Install Leaderboard using the "go get" command:

    go get github.com/topfreegames/podium

And then run

    make setup
    
Quickstart (for using as library)
--------------------------------

```
import (
	"context"
	"fmt"
	"log"

	"github.com/topfreegames/podium/leaderboard"
)

func main() {
	leaderboards, err := leaderboard.NewClient("localhost", 6379, "", 0, 200)
	if err != nil {
		log.Fatalf("leaderboard.NewClient failed: %v", err)
	}

	const leaderboardID = "myleaderboardID"

	//setting player scores
	players := leaderboard.Members{
		&leaderboard.Member{Score: 10, PublicID: "player1"},
		&leaderboard.Member{Score: 20, PublicID: "player2"},
	}

	err = leaderboards.SetMembersScore(context.Background(), leaderboardID, players, false, "")
	if err != nil {
		log.Fatalf("leaderboards.SetMembersScore failed: %v", err)
	}

	//getting the leaders of the leaderboard
	leaders, err := leaderboards.GetLeaders(context.Background(), leaderboardID, 10, 1, "desc")
	if err != nil {
		log.Fatalf("leaderboards.GetLeaders failed: %v", err)
	}

	for _, player := range leaders {
		fmt.Printf("Player(id: %s, score: %d rank: %d)\n", player.PublicID, player.Score, player.Rank)
	}
}
```

Testing
-------
    make test

Coverage
---------
    make test-coverage test-coverage-html

Benchmarks
----------

Podium benchmarks prove it's blazing fast:

    BenchmarkSetMemberScore-8                           30000        284307 ns/op       0.32 MB/s        5635 B/op         81 allocs/op
    BenchmarkSetMembersScore-8                           5000       1288746 ns/op       3.01 MB/s       51452 B/op        583 allocs/op
    BenchmarkIncrementMemberScore-8                     30000        288306 ns/op       0.32 MB/s        5651 B/op         81 allocs/op
    BenchmarkRemoveMember-8                             50000        202398 ns/op       0.08 MB/s        4648 B/op         68 allocs/op
    BenchmarkGetMember-8                                30000        215802 ns/op       0.33 MB/s        4728 B/op         68 allocs/op
    BenchmarkGetMemberRank-8                            50000        201367 ns/op       0.28 MB/s        4712 B/op         68 allocs/op
    BenchmarkGetAroundMember-8                          20000        397849 ns/op       3.14 MB/s        8703 B/op         69 allocs/op
    BenchmarkGetTotalMembers-8                          50000        192860 ns/op       0.16 MB/s        4536 B/op         64 allocs/op
    BenchmarkGetTopMembers-8                            20000        306186 ns/op       3.85 MB/s        8585 B/op         66 allocs/op
    BenchmarkGetTopPercentage-8                          1000      10011287 ns/op      11.88 MB/s      510300 B/op         77 allocs/op
    BenchmarkSetMemberScoreForSeveralLeaderboards-8      1000     106129629 ns/op       1.03 MB/s      516103 B/op         98 allocs/op
    BenchmarkGetMembers-8                                2000       3931289 ns/op       9.13 MB/s      243755 B/op         76 allocs/op

To run the benchmarks: `make bench-redis bench-podium-app bench-run`.

Our builds also show the difference to the previous build.

License
-------
© 2016, Top Free Games. Released under the [MIT License](LICENSE).

Forked from:
[© 2013, Maxwell Dayvson da Silva.](https://github.com/dayvson/go-leaderboard)

