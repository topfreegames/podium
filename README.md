Leaderboard
===========

A leaderboard written in [Go](http://golang.org/) using [Redis](http://redis.io/) database.

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

    go get github.com/topfreegames/go-leaderboard

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
