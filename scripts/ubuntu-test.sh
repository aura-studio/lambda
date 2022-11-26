#!/bin/bash
GOOS=linux CGO_ENABLED=1 go install github.com/aura-studio/dynamic/dynamic-cli@v1.0.2
dynamic-cli build github.com/aura-studio/testdynamic1@test
dynamic-cli build github.com/aura-studio/testdynamic2@test
GO_DYNAMIC_WAREHOUSE=/opt/go-dynamic-warehouse go run ./httpserver