#!/bin/bash

export dynamic_version="v1.0.3"
docker build -t lambda-dynamic-cli:${dynamic_version} -f ./docker/dynamic-cli.Dockerfile --build-arg VERSION=${dynamic_version} .
docker run -it --rm -v `pwd`/warehouse:/lambda/warehouse lambda-dynamic-cli:${dynamic_version} build github.com/aura-studio/testdynamic1@test -w /lambda/warehouse
docker run -it --rm -v `pwd`/warehouse:/lambda/warehouse lambda-dynamic-cli:${dynamic_version} build github.com/aura-studio/testdynamic2@test -w /lambda/warehouse

docker build -t lambda-http-server:s3 -f ./docker/http-server-s3.Dockerfile --build-arg S3_REGION=us-west-1 --build-arg S3_BUCKET=mirroring-lambda --build-arg S3_ACCESS_KEY=AKIAY6OYEB6NCFBK4ZML --build-arg S3_SECRET_KEY=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 .
docker stop lambda-http-server || true
docker rm lambda-http-server || true
docker run -it --rm --name lambda-http-server -p 8080:8000 lambda-http-server:s3