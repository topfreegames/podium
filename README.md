# Podium

[![Build Status](https://travis-ci.org/topfreegames/podium.svg?branch=master)](https://travis-ci.org/topfreegames/podium)
[![Coverage Status](https://coveralls.io/repos/github/topfreegames/podium/badge.svg?branch=master)](https://coveralls.io/github/topfreegames/podium?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/topfreegames/podium)](https://goreportcard.com/report/github.com/topfreegames/podium)
[![Docs](https://readthedocs.org/projects/podium/badge/?version=latest
)](http://podium.readthedocs.io/en/latest/) [![](https://imagelayers.io/badge/tfgco/podium:latest.svg)](https://imagelayers.io/?images=tfgco/podium:latest 'Podium Image Layers')

A leaderboard system written in [Go](http://golang.org/) using [Redis](http://redis.io/) database. For more info, [read the docs](http://podium.readthedocs.io/en/latest/).

Features
--------

* **Multi-tenant** - Just vary the name of the leaderboard and you can have any number of tenants using leaderboards;
* **Seasonal Leaderboards** - Including suffixes like `year2016week01` or `year2016month06` is all you need to create seasonal leaders. I'm serious! That's all there is to it;
* **No leaderboard configuration** - Just start notifying scores for members of a leaderboard. There's no need to create, configure or maintain leaderboards. Let Podium do that for you;
* **Top Members** - Get the top members of a leaderboard whether you need by absolute value (top 200 members) or percentage (top 3% members);
* **Members around me** - Podium easily returns members around a specific member in the leaderboard. It will even compensate if you ask for the top member or last member to make sure you get a consistent amount of members;
* **Batch score update** - Send a member score to many different leaderboards in a single operation. This allows easy tracking of member rankings in several leaderboards at once (global, regional, clan, etc.);
* **Easy to deploy** - Podium comes with containers already exported to docker hub for every single of our successful builds. Just pick your choice!
* **Leaderboards with expiration** - If a player last update is older than (timeNow - X seconds), delete it from the leaderboard

Installation
------------

Install Leaderboard using the "go get" command:

    go get github.com/topfreegames/podium

And then run

    make setup

Testing
-------
    make test

Coverage
---------
    make test-coverage test-coverage-html

Benchmarks
----------

Podium benchmarks prove it's blazing fast:

    BenchmarkSetMemberScore-4                      	   20000	    285962 ns/op	   0.32 MB/s	    5219 B/op	      71 allocs/op
    BenchmarkRemoveMember-4                        	   50000	    220081 ns/op	   0.07 MB/s	    3823 B/op	      53 allocs/op
    BenchmarkGetMember-4                           	   30000	    266313 ns/op	   0.27 MB/s	    4143 B/op	      56 allocs/op
    BenchmarkGetMemberRank-4                       	   30000	    231241 ns/op	   0.25 MB/s	    4319 B/op	      57 allocs/op
    BenchmarkGetAroundMember-4                     	   10000	    519063 ns/op	   2.38 MB/s	    8314 B/op	      58 allocs/op
    BenchmarkGetTotalMembers-4                     	   30000	    196277 ns/op	   0.15 MB/s	    3936 B/op	      52 allocs/op
    BenchmarkGetTopMembers-4                       	   20000	    455470 ns/op	   2.59 MB/s	    7973 B/op	      54 allocs/op
    BenchmarkGetTopPercentage-4                    	     500	  14354336 ns/op	   8.28 MB/s	  509746 B/op	      65 allocs/op
    BenchmarkSetMemberScoreForSeveralLeaderboards-4	    1000	  70326444 ns/op	   1.55 MB/s	  534548 B/op	      96 allocs/op

To run the benchmarks: `make bench-redis bench-podium-app bench-run`.

Our builds also show the difference to the previous build.

License
-------
© 2016, Top Free Games. Released under the [MIT License](LICENSE).

Forked from:
[© 2013, Maxwell Dayvson da Silva.](https://github.com/dayvson/go-leaderboard)

