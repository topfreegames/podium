# Podium

[![Build Status](https://travis-ci.org/topfreegames/podium.svg?branch=master)](https://travis-ci.org/topfreegames/podium)
[![Coverage Status](https://coveralls.io/repos/github/topfreegames/podium/badge.svg?branch=master)](https://coveralls.io/github/topfreegames/podium?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/topfreegames/podium)](https://goreportcard.com/report/github.com/topfreegames/podium)
[![Docs](https://readthedocs.org/projects/snt/badge/?version=latest
)](http://snt.readthedocs.io/en/latest/)
[![](https://imagelayers.io/badge/tfgco/podium:latest.svg)](https://imagelayers.io/?images=tfgco/podium:latest 'Khan Image Layers')

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

License
-------
© 2016, Top Free Games. Released under the [MIT License](LICENSE).

Forked from:
[© 2013, Maxwell Dayvson da Silva.](https://github.com/dayvson/go-leaderboard)
