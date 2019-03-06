#!/bin/bash

VERSION=$(cat ./api/version.go | grep "var VERSION" | awk ' { print $4 } ' | sed s/\"//g)

cp ./config/default.yaml ./dev

docker build -t podium .
if [ $? -ne 0 ]; then
    exit 1
fi
docker build -t podium-dev -f ./DevDockerfile .
if [ $? -ne 0 ]; then
    exit 1
fi

docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
if [ $? -ne 0 ]; then
    exit 1
fi

docker tag podium:latest tfgco/podium:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/podium:$VERSION.$TRAVIS_BUILD_NUMBER
if [ $? -ne 0 ]; then
    exit 1
fi
docker tag podium:latest tfgco/podium:latest
docker push tfgco/podium
if [ $? -ne 0 ]; then
    exit 1
fi

docker tag podium-dev:latest tfgco/podium-dev:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/podium-dev:$VERSION.$TRAVIS_BUILD_NUMBER
if [ $? -ne 0 ]; then
    exit 1
fi
docker tag podium-dev:latest tfgco/podium-dev:latest
docker push tfgco/podium-dev
if [ $? -ne 0 ]; then
    exit 1
fi

DOCKERHUB_LATEST=$(python ./scripts/get_latest_tag.py)

if [ "$DOCKERHUB_LATEST" != "$VERSION.$TRAVIS_BUILD_NUMBER" ]; then
    echo "Last version is not in docker hub!"
    echo "docker hub: $DOCKERHUB_LATEST, expected: $VERSION.$TRAVIS_BUILD_NUMBER"
    exit 1
fi
