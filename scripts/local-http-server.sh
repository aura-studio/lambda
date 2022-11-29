#!/bin/bash

export dynamic_version="v1.0.3"
go install github.com/aura-studio/dynamic-cli@${dynamic_version}
dynamic-cli build github.com/aura-studio/testdynamic1@test -w /lambda/warehouse
dynamic-cli build github.com/aura-studio/testdynamic2@test -w /lambda/warehouse

go run ./http-server