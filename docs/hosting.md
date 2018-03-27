Hosting Podium
==============

There are three ways to host Podium: docker, binaries or from source.

## Docker

Running Podium with docker is rather simple. Our docker container image comes bundled with the API binary. All you need to do is load balance all the containers and you're good to go.

Podium uses Redis to store leaderboard information. The container takes parameters to specify this connection:

* `PODIUM_REDIS_HOST` - Redis host to connect to;
* `PODIUM_REDIS_PORT` - Redis port to connect to;
* `PODIUM_REDIS_PASSWORD` - Password of the Redis Server to connect to;
* `PODIUM_REDIS_DB` - DB Number of the Redis Server to connect to;

Other than that, there are a couple more configurations you can pass using environment variables:

* `PODIUM_NEWRELIC_KEY` - If you have a [New Relic](https://newrelic.com/) account, you can use this variable to specify your API Key to populate data with New Relic API;
* `PODIUM_SENTRY_URL` - If you have a [sentry server](https://docs.getsentry.com/hosted/) you can use this variable to specify your project's URL to send errors to;
* `PODIUM_BASICAUTH_USERNAME` - If you specify this key, Podium will be configured to use basic auth with this user;
* `PODIUM_BASICAUTH_PASSWORD` - If you specify `BASICAUTH_USERNAME`, Podium will be configured to use basic auth with this password.
* `PODIUM_EXTENSIONS_DOGSTATSD_HOST` - If you have a [statsd datadog daemon](https://docs.datadoghq.com/developers/dogstatsd/), Podium will publish metrics to the given host at a certain port. Ex. localhost:8125
]* `PODIUM_EXTENSIONS_DOGSTATSD_RATE` - If you have a [statsd daemon](https://docs.datadoghq.com/developers/dogstatsd/), Podium will export metrics to the deamon at the given rate
* `PODIUM_EXTENSIONS_DOGSTATSD_TAGS_PREFIX` - If you have a [statsd daemon](https://docs.datadoghq.com/developers/dogstatsd/), you may set a prefix to every tag sent to the daemon

## Binaries

Whenever we publish a new version of Podium, we'll always supply binaries for both Linux and Darwin, on i386 and x86_64 architectures. If you'd rather run your own servers instead of containers, just use the binaries that match your platform and architecture.

The API server is the `podium` binary. It takes a configuration yaml file that specifies the connection to Redis and some additional parameters. You can learn more about it at [default.yaml](https://github.com/topfreegames/podium/blob/master/config/default.yaml).

## Source

Left as an exercise to the reader.
