module github.com/topfreegames/podium

go 1.15

require (
	github.com/DataDog/datadog-go v4.5.1+incompatible // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20180315120708-ccb8e960c48f // indirect
	github.com/bsm/redis-lock v6.0.0+incompatible // indirect
	github.com/certifi/gocertifi v0.0.0-20200922220541-2c3bb06c6054 // indirect
	github.com/getsentry/raven-go v0.0.0-20170918144728-1452f6376ddb
	github.com/go-pg/pg v8.0.7+incompatible // indirect
	github.com/go-redis/redis v6.13.2+incompatible // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/gorilla/context v0.0.0-20160226214623-1ea25387ff6f // indirect
	github.com/gorilla/mux v1.6.1 // indirect
	github.com/gosuri/uilive v0.0.0-20160202011846-efb88ccd0599 // indirect
	github.com/gosuri/uiprogress v0.0.0-20160202012259-a9f819bfc744
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/grpc-gateway v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/klauspost/compress v0.0.0-20160919184342-d0763f13d86e // indirect
	github.com/klauspost/cpuid v0.0.0-20160302075316-09cded8978dc // indirect
	github.com/klauspost/crc32 v0.0.0-20160219142609-19b0b332c9e4 // indirect
	github.com/mailru/easyjson v0.0.0-20180320131758-517203d186eb
	github.com/newrelic/go-agent v1.11.0
	github.com/onsi/ginkgo v1.16.2
	github.com/onsi/gomega v1.11.0
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20180125231941-8732c616f529
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1
	github.com/topfreegames/extensions v8.0.2+incompatible
	github.com/topfreegames/podium/leaderboard/v2 v2.0.0
	github.com/uber/jaeger-client-go v2.25.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v0.0.0-20161005094451-07f692d02d61
	go.uber.org/zap v1.16.0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/sys v0.0.0-20210521090106-6ca3eb03dfc2 // indirect
	golang.org/x/tools v0.1.1 // indirect
	google.golang.org/genproto v0.0.0-20200311144346-b662892dd51b
	google.golang.org/grpc v1.28.0
	mellium.im/sasl v0.2.1 // indirect
)

replace github.com/topfreegames/podium/leaderboard/v2 v2.0.0 => ./leaderboard
