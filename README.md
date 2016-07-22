# Podium

[![Build Status](https://travis-ci.org/topfreegames/podium.svg?branch=master)](https://travis-ci.org/topfreegames/podium)
[![Coverage Status](https://coveralls.io/repos/github/topfreegames/podium/badge.svg?branch=master)](https://coveralls.io/github/topfreegames/podium?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/topfreegames/podium)](https://goreportcard.com/report/github.com/topfreegames/podium)
[![](https://imagelayers.io/badge/tfgco/podium:latest.svg)](https://imagelayers.io/?images=tfgco/podium:latest 'Podium Image Layers')

A leaderboard system written in [Go](http://golang.org/) using [Redis](http://redis.io/) database.

Features
--------

* Create multiple Leaderboards by name
* You can set a member score and the leaderboard will be updated automatically
* Remove a member from a specific Leaderboard
* Get total of users in a specific Leaderboard and also how many pages it has.
* Get leaders on any page
* Get an "Around Me" leaderboard for a member
* Get rank and score for an arbitrary list of members (e.g. friends)

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

To run the benchmarks: `make bench-podium-app bench-run`.

Our builds also show the difference to the previous build.

License
-------
© 2016, Top Free Games. Released under the [MIT License](LICENSE).

Forked from:
[© 2013, Maxwell Dayvson da Silva.](https://github.com/dayvson/go-leaderboard)
