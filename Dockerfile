# podium - webhook dispatching service
# https://github.com/topfreegames/podium
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>

FROM golang:1.6.2-alpine

MAINTAINER TFG Co <backend@tfgco.com>

EXPOSE 80

RUN apk update
RUN apk add bash git make g++ apache2-utils

# http://stackoverflow.com/questions/34729748/installed-go-binary-not-found-in-path-on-alpine-linux-docker
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

ADD bin/podium-linux-x86_64 /go/bin/podium
RUN chmod +x /go/bin/podium

RUN mkdir -p /home/podium/
ADD ./docker/default.yaml /home/podium/default.yaml

ENV PODIUM_REDIS_HOST localhost
ENV PODIUM_REDIS_PORT 6379
ENV PODIUM_REDIS_PASSWORD ""
ENV PODIUM_REDIS_DB 0
ENV PODIUM_SENTRY_URL ""
ENV PODIUM_BASICAUTH_USERNAME ""
ENV PODIUM_BASICAUTH_PASSWORD ""

CMD /go/bin/podium start -c /home/podium/default.yaml -p 80
