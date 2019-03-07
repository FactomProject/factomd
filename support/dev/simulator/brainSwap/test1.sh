#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
SIM_TEST=${DIR}/../../../../simTest/BrainSwapFollower_test.go
cd ${DIR}/v1
go test -v $SIM_TEST #> out1.txt
