#!/bin/bash

export dynamic_version="v1.0.3"
docker build -t lambda-dynamic-cli:${dynamic_version} -f ./docker/dynamic-cli.Dockerfile --build-arg VERSION=${dynamic_version} .
docker run -it --rm -v `pwd`/warehouse:/lambda/warehouse lambda-dynamic-cli:${dynamic_version} build github.com/aura-studio/testdynamic1@test -w /lambda/warehouse
docker run -it --rm -v `pwd`/warehouse:/lambda/warehouse lambda-dynamic-cli:${dynamic_version} build github.com/aura-studio/testdynamic2@test -w /lambda/warehouse

export ecr_repo="615170445210.dkr.ecr.us-west-1.amazonaws.com/slots-nano:test"
export function_name="slots-nano-test"
docker build -t ${ecr_repo} -f ./docker/http-server.Dockerfile .
docker push ${ecr_repo}
aws lambda update-function-code --function-name ${function_name} --image-uri ${ecr_repo}