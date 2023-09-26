module github.com/topfreegames/podium

go 1.20

require (
	github.com/getsentry/raven-go v0.0.0-20170918144728-1452f6376ddb
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.3
	github.com/gosuri/uiprogress v0.0.0-20160202012259-a9f819bfc744
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.1
	github.com/mailru/easyjson v0.0.0-20180320131758-517203d186eb
	github.com/mitchellh/mapstructure v1.5.0
	github.com/newrelic/go-agent v1.11.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.18.1
	github.com/opentracing-contrib/go-grpc v0.0.0-20210225150812-73cb765af46e
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/rcrowley/go-metrics v0.0.0-20180125231941-8732c616f529
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.16.0
	github.com/topfreegames/extensions v8.5.0+incompatible
	github.com/topfreegames/podium/leaderboard/v2 v2.0.0
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/valyala/fasthttp v0.0.0-20161005094451-07f692d02d61
	go.uber.org/zap v1.21.0
	google.golang.org/genproto/googleapis/api v0.0.0-20230724170836-66ad5b6ff146
	google.golang.org/grpc v1.56.2
	google.golang.org/protobuf v1.31.0
)

require (
	cloud.google.com/go/compute v1.20.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/DataDog/datadog-go v4.8.3+incompatible // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/bsm/redis-lock v6.0.0+incompatible // indirect
	github.com/certifi/gocertifi v0.0.0-20200922220541-2c3bb06c6054 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-pg/pg v8.0.7+incompatible // indirect
	github.com/go-redis/redis v6.13.2+incompatible // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gosuri/uilive v0.0.0-20160202011846-efb88ccd0599 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/mod v0.9.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/oauth2 v0.10.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230706204954-ccb25ca9f130 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230724170836-66ad5b6ff146 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mellium.im/sasl v0.3.1 // indirect
)

replace github.com/topfreegames/podium/leaderboard/v2 v2.0.0 => ./leaderboard
