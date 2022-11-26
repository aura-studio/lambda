#!/bin/bash
docker build -t lambda-http-server -f http.Dockerfile .
docker run -d --rm --name lambda-http-server -p 8080:8000 -v `pwd`/warehouse:/opt/go-dynamic-warehouse lambda-http-server
# docker build -t 615170445210.dkr.ecr.us-west-1.amazonaws.com/slots-nano:test .
# docker push 615170445210.dkr.ecr.us-west-1.amazonaws.com/slots-nano:test
# sudo aws lambda update-function-code --function-name slots-nano-test --image-uri 615170445210.dkr.ecr.us-west-1.amazonaws.com/slots-nano:test

# sam build

		# go install github.com/aura-studio/dynamic/dynamic-cli@test
		# dynamic-cli build github.com/aura-studio/testdynamic1@test
		# dynamic-cli build github.com/aura-studio/testdynamic2@test
		# go run ./testdynamic@test.go


