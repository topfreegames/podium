module github.com/topfreegames/podium/leaderboard/v2

go 1.15

replace github.com/codahale/hdrhistogram => github.com/HdrHistogram/hdrhistogram-go v0.0.0-20200919145931-8dac23c8dac1

require (
	github.com/go-redis/redis v6.13.2+incompatible
	github.com/onsi/ginkgo v1.16.1
	github.com/onsi/gomega v1.11.0
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/viper v1.7.1
	github.com/topfreegames/extensions/v9 v9.0.0
)
