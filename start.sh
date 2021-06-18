#!/usr/bin/env bash

set -e

source .env.repository

SERVICE_NAME=$(basename "$PWD")

DOCKER_IMAGE_TAG=gitlab.springup.xyz:5050/autotouch/${SERVICE_NAME}:latest

docker run --rm --network=host -v ${PWD}/config/config.yaml:/app/config.yaml ${DOCKER_IMAGE_TAG}