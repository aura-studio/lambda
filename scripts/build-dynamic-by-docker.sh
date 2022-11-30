#!/bin/bash
export dynamic_version="v1.0.3"
export dynamic_warehouse="/tmp/warehouse"
docker build -t lambda-dynamic-cli:${dynamic_version} -f ./docker/dynamic-cli.Dockerfile --build-arg VERSION=${dynamic_version} .
docker run -it --rm -v `pwd`/warehouse:${dynamic_warehouse} lambda-dynamic-cli:${dynamic_version} build github.com/aura-studio/testdynamic1@test -w ${dynamic_warehouse}
docker run -it --rm -v `pwd`/warehouse:${dynamic_warehouse} lambda-dynamic-cli:${dynamic_version} build github.com/aura-studio/testdynamic2@test -w ${dynamic_warehouse}
