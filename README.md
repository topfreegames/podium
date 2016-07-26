# Podium

[![Build Status](https://travis-ci.org/topfreegames/podium.svg?branch=master)](https://travis-ci.org/topfreegames/podium)
[![Coverage Status](https://coveralls.io/repos/github/topfreegames/podium/badge.svg?branch=master)](https://coveralls.io/github/topfreegames/podium?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/topfreegames/podium)](https://goreportcard.com/report/github.com/topfreegames/podium)
[![Docs](https://readthedocs.org/projects/podium/badge/?version=latest
)](http://podium.readthedocs.io/en/latest/)
[![](https://imagelayers.io/badge/tfgco/podium:latest.svg)](https://imagelayers.io/?images=tfgco/podium:latest 'Podium Image Layers')

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

    BenchmarkSetUserScore-8    	   50000	    121245 ns/op	   0.75 MB/s	    5281 B/op	      71 allocs/op
    BenchmarkRemoveUser-8      	  100000	    119257 ns/op	   0.13 MB/s	    3840 B/op	      53 allocs/op
    BenchmarkGetUser-8         	  100000	    102257 ns/op	   0.63 MB/s	    4367 B/op	      57 allocs/op
    BenchmarkGetUserRank-8     	  100000	     82795 ns/op	   0.59 MB/s	    4367 B/op	      57 allocs/op
    BenchmarkGetAroundUser-8   	   10000	    512711 ns/op	   1.94 MB/s	    8028 B/op	      58 allocs/op
    BenchmarkGetTotalMembers-8 	  100000	     79465 ns/op	   0.39 MB/s	    4192 B/op	      53 allocs/op
    BenchmarkGetTotalPages-8   	  100000	     80390 ns/op	   0.36 MB/s	    4160 B/op	      53 allocs/op
    BenchmarkGetTopUsers-8     	   20000	    498763 ns/op	   2.04 MB/s	    7717 B/op	      55 allocs/op
    BenchmarkGetTopPercentage-8	   10000	   5265023 ns/op	   9.31 MB/s	  231009 B/op	      60 allocs/op

To run the benchmarks: `make bench-redis bench-podium-app bench-run`.

Our builds also show the difference to the previous build.

License
-------
© 2016, Top Free Games. Released under the [MIT License](LICENSE).

Forked from:
[© 2013, Maxwell Dayvson da Silva.](https://github.com/dayvson/go-leaderboard)
