#!/bin/sh
export CGO_ENABLED=1
export GO111MODULE=on
export GOPRIVATE=github.com/aura-studio/testdynamic2
go mod tidy
go build -o /tmp/warehouse/testdynamic2_test/libcgo_testdynamic2_test.so -buildmode=c-shared /tmp/warehouse/testdynamic2_test/libcgo_testdynamic2_test
go build -o /tmp/warehouse/testdynamic2_test/libgo_testdynamic2_test.so -buildmode=plugin -ldflags="-r /tmp/warehouse/testdynamic2_test/" /tmp/warehouse/testdynamic2_test/libgo_testdynamic2_test
