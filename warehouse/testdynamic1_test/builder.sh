#!/bin/sh
export CGO_ENABLED=1
export GO111MODULE=on
export GOPRIVATE=github.com/aura-studio/testdynamic1
go mod tidy
go build -o /tmp/warehouse/testdynamic1_test/libcgo_testdynamic1_test.so -buildmode=c-shared /tmp/warehouse/testdynamic1_test/libcgo_testdynamic1_test
go build -o /tmp/warehouse/testdynamic1_test/libgo_testdynamic1_test.so -buildmode=plugin -ldflags="-r /tmp/warehouse/testdynamic1_test/" /tmp/warehouse/testdynamic1_test/libgo_testdynamic1_test
