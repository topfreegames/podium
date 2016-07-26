Podium's Benchmarks
===================

You can see podium's benchmarks in our [CI server](https://travis-ci.org/topfreegames/podium/) as they get run with every build.

## Running Benchmarks

If you want to run your own benchmarks, just download the project, and run:

```
$ make bench-redis bench-podium-app bench-run
```

## Generating test data

If you want to run your perf tests against a database with more volume of data, just run this command, instead:

```
$ make bench-redis bench-seed bench-podium-app bench-run
```
**Warning**: This will take a long time running.

## Results

The results should be similar to these:

```
BenchmarkSetMemberScore-4                      	   20000	    285962 ns/op	   0.32 MB/s	    5219 B/op	      71 allocs/op
BenchmarkRemoveMember-4                        	   50000	    220081 ns/op	   0.07 MB/s	    3823 B/op	      53 allocs/op
BenchmarkGetMember-4                           	   30000	    266313 ns/op	   0.27 MB/s	    4143 B/op	      56 allocs/op
BenchmarkGetMemberRank-4                       	   30000	    231241 ns/op	   0.25 MB/s	    4319 B/op	      57 allocs/op
BenchmarkGetAroundMember-4                     	   10000	    519063 ns/op	   2.38 MB/s	    8314 B/op	      58 allocs/op
BenchmarkGetTotalMembers-4                     	   30000	    196277 ns/op	   0.15 MB/s	    3936 B/op	      52 allocs/op
BenchmarkGetTopMembers-4                       	   20000	    455470 ns/op	   2.59 MB/s	    7973 B/op	      54 allocs/op
BenchmarkGetTopPercentage-4                    	     500	  14354336 ns/op	   8.28 MB/s	  509746 B/op	      65 allocs/op
BenchmarkSetMemberScoreForSeveralLeaderboards-4	    1000	  70326444 ns/op	   1.55 MB/s	  534548 B/op	      96 allocs/op
```
