#!/usr/bin/env bash

set -e

source .env

docker run --rm --net=host -v ${PWD}/config.yaml:/etc/backup-db-to-dropbox/config.yaml ${DOCKER_REPOSITORY}/${DOCKER_IMAGE_NAME}:dev