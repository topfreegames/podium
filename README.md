Leaderboard
===========

A leaderboard write in [Go](http://golang.org/) using [Redis](http://redis.io/) database.

Features
--------

* Create multiple Leaderboards by name 
* You can rank a member with and the leaderboard will be updated automatically
* Remove a member from a specific Leaderboard
* Get total of users in a specific Leaderboard and also how many pages it has.
* Get leaders or any page
* Get an "Around Me" leaderboard for a member
* Get rank and score for an arbitrary list of members (e.g. friends)	

How to use
----------

*****TODO:

Installation
------------

Install Leaderboard using the "go get" command:

    go get github.com/dayvson/go-leaderboard


Dependencies
------------
redigo (github.com/garyburd/redigo/redis)


Contributing
------------

Contributions are welcome.
Take care to maintain the existing coding style. 
Add unit tests for any new or changed functionality. 
Open a pull request :)


License
-------
© 2013, Maxwell Dayvson da Silva. Released under the [MIT License](LICENSE).
