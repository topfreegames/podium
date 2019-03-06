# podium - webhook dispatching service
# https://github.com/topfreegames/podium
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2019 Top Free Games <backend@tfgco.com>

FROM alpine:3.9

MAINTAINER TFG Co <backend@tfgco.com>

EXPOSE 80

RUN apk add bash
ADD bin/podium-linux-x86_64 /go/bin/podium
RUN chmod +x /go/bin/podium

RUN mkdir -p /home/podium/
ADD ./config/default.yaml /home/podium/default.yaml

ENV PODIUM_REDIS_HOST localhost
ENV PODIUM_REDIS_PORT 6379
ENV PODIUM_REDIS_PASSWORD ""
ENV PODIUM_REDIS_DB 0
ENV PODIUM_SENTRY_URL ""
ENV PODIUM_BASICAUTH_USERNAME ""
ENV PODIUM_BASICAUTH_PASSWORD ""

CMD /go/bin/podium start -c /home/podium/default.yaml -p 80
