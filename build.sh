#!/usr/bin/env bash

set -e

source .env

make build

SERVICE_NAME=`basename "$PWD"`

docker login ${DOCKER_REPOSITORY} --username ${DOCKER_USERNAME} --password ${DOCKER_PASSWORD}
docker build -t ${DOCKER_REPOSITORY}/${DOCKER_IMAGE_NAME}:dev .
docker push ${DOCKER_REPOSITORY}/${DOCKER_IMAGE_NAME}:dev