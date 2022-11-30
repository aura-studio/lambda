#!/bin/bash
export dynamic_version="v1.0.3"
export dynamic_warehouse="/tmp/warehouse"
go install github.com/aura-studio/dynamic-cli@${dynamic_version}
dynamic-cli build github.com/aura-studio/testdynamic1@test -w ${dynamic_warehouse}
dynamic-cli build github.com/aura-studio/testdynamic2@test -w ${dynamic_warehouse}
