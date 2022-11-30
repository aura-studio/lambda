#!/bin/bash
docker build -t lambda-http-server -f ./docker/http-server.Dockerfile .
docker stop lambda-http-server || true
docker rm lambda-http-server || true
docker run -it --rm --name lambda-http-server -p 8080:8000 lambda-http-server