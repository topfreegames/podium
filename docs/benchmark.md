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
BenchmarkSetMemberScore-4                      	   30000	    280816 ns/op	   0.32 MB/s	    5218 B/op	      71 allocs/op
BenchmarkRemoveMember-4                        	   30000	    294856 ns/op	   0.05 MB/s	    3823 B/op	      53 allocs/op
BenchmarkGetMember-4                           	   30000	    241180 ns/op	   0.29 MB/s	    4143 B/op	      56 allocs/op
BenchmarkGetMemberRank-4                       	   50000	    190689 ns/op	   0.30 MB/s	    4319 B/op	      57 allocs/op
BenchmarkGetAroundMember-4                     	   20000	    451855 ns/op	   2.76 MB/s	    8317 B/op	      58 allocs/op
BenchmarkGetTotalMembers-4                     	   50000	    177635 ns/op	   0.18 MB/s	    3936 B/op	      52 allocs/op
BenchmarkGetTotalPages-4                       	   50000	    179556 ns/op	   0.17 MB/s	    3904 B/op	      52 allocs/op
BenchmarkGetTopMembers-4                       	   20000	    347018 ns/op	   3.40 MB/s	    7973 B/op	      54 allocs/op
BenchmarkGetTopPercentage-4                    	     500	  14287286 ns/op	   8.32 MB/s	  509741 B/op	      65 allocs/op
BenchmarkSetMemberScoreForSeveralLeaderboards-4	    1000	  73939762 ns/op	   1.47 MB/s	  534492 B/op	      96 allocs/op
```
