#!/bin/bash

set -ex
rm -rf ./bin
mkdir bin
go build -o ./bin/fetch-k8s-resource
docker build -t fetch-k8s-resource:$1 . --no-cache
docker tag fetch-k8s-resource:$1 shovan1995/fetch-k8s-resource:$1
docker push shovan1995/fetch-k8s-resource:$1
