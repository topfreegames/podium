Leaderboard
===========

A leaderboard written in [Go](http://golang.org/) using [Redis](http://redis.io/) database.

Features
--------

* Create multiple Leaderboards by name 
* You can rank a member with and the leaderboard will be updated automatically
* Remove a member from a specific Leaderboard
* Get total of users in a specific Leaderboard and also how many pages it has.
* Get leaders on any page
* Get an "Around Me" leaderboard for a member
* Get rank and score for an arbitrary list of members (e.g. friends)	

How to use
----------

Create a new leaderboard or attach to an existing leaderboard named 'highscores': 
<pre>
    highScore := NewLeaderboard(RedisSettings{
    	Host: "localhost:6379", 
    	Password: "mypassword",
    }, "highscores", 10)
    //return a Leaderboard: Leaderboard{name:"highscores", pageSize:10}
</pre>  

Adding members to highscores using RankMember(username, score):
<pre>
    highScore.RankMember("dayvson", 9876)
    highScore.RankMember("arthur", 2000123)
    highScore.RankMember("felipe", 100000)
</pre>

You can call RankMember with the same member and the leaderboard will be updated automatically:
<pre>
	highScore.RankMember("dayvson", 7481523)
	//return an user: User{name:"dayvson", score:7481523, rank:1}
</pre>

Getting a total members on highscores:
<pre>
	highScore.TotalMembers()
	//return an int: 3
</pre>

Getting the rank from a member:
<pre>
	highScore.GetRank("dayvson")
	//return an int: 1
</pre>

Getting a member from a rank position:
<pre>
	highScore.GetMemberByRank(2)
	//return an user: User{name:"felipe", score:100000, rank:2}
</pre>

Getting members around you:
<pre>
	highScore.GetAroundMe("felipe")
	//return an array of users around you [pageSize]User:
</pre>

Getting Leaders by page:
<pre>
	highScore.GetLeaders(1)
	//return an array of users with highest score in a first page (you can specify any page): [pageSize]User
</pre>

Installation
------------

Install Leaderboard using the "go get" command:

    go get github.com/dayvson/go-leaderboard


Testing
-------
    make test

Dependencies
------------
* Go language distribution
* redigo (github.com/garyburd/redigo/redis)



Contributing
------------

* Contributions are welcome.
* Take care to maintain the existing coding style.
* Add unit tests for any new or changed functionality.
* Open a pull request :)


License
-------
© 2013, Maxwell Dayvson da Silva. Released under the [MIT License](LICENSE).
