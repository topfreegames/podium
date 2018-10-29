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
```
