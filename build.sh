#!/usr/bin/env bash

set -e

source .env.repository

SERVICE_NAME=$(basename "$PWD")

function show_usage() {
    echo "Usage: $(basename $0) [options]"
    echo "Options:"
    echo "    -c make clean"
    echo "    -l build for Linux"
    echo "    -d build for Darwin"
    echo "    -b docker build"
    echo "    -p docker push"
    echo "    -h help"
    exit 0
}

[ $# -eq 0 ] && show_usage

while getopts 'cldbph' opt; do
    case "$opt" in
    c) MAKE_CLEAN=TRUE ;;
    l) BUILD_FOR_LINUX=TRUE ;;
    d) BUILD_FOR_DARWIN=TRUE ;;
    b) DOCKER_BUILD=TRUE ;;
    p) DOCKER_PUSH=TRUE ;;
    h | ? | *) show_usage ;;
    esac
done

if [ "$MAKE_CLEAN" == TRUE ]; then
    echo "===> Cleaning..."
    go clean
fi

if [ "$BUILD_FOR_LINUX" == TRUE ]; then
    echo "===> Building for Linux..."
    env GOOS=linux GOARCH=amd64 go build -o target/linux/$SERVICE_NAME ./*.go
fi

if [ "$BUILD_FOR_DARWIN" == TRUE ]; then
    echo "===> Building for Darwin..."
    env GOOS=darwin GOARCH=amd64 go build -o target/darwin/$SERVICE_NAME ./*.go
fi

DOCKER_IMAGE_TAG=gitlab.springup.xyz:5050/autotouch/${SERVICE_NAME}:latest

if [ "$DOCKER_BUILD" == TRUE ]; then
    echo "===> Deleting old docker images..."
    docker images --format '{{.Repository}} {{.ID}}' | grep "${SERVICE_NAME}" | cut -d' ' -f2 | xargs docker rmi
    
    echo "===> Docker login..."
    echo ${DOCKER_PASSWORD} | docker login gitlab.springup.xyz:5050 --username ${DOCKER_USERNAME} --password-stdin

    echo "===> Docker build..."
    docker build -t $DOCKER_IMAGE_TAG .
fi

if [ "$DOCKER_PUSH" == TRUE ]; then
    echo "===> Docker pushing..."
    docker push $DOCKER_IMAGE_TAG
fi