#!/bin/bash
docker build -t 615170445210.dkr.ecr.us-west-1.amazonaws.com/slots-nano:test .
docker push 615170445210.dkr.ecr.us-west-1.amazonaws.com/slots-nano:test
sudo aws lambda update-function-code --function-name slots-nano-test --image-uri 615170445210.dkr.ecr.us-west-1.amazonaws.com/slots-nano:test

sam build 