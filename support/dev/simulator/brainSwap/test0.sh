#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
SIM_TEST=${DIR}/../../../../peerTest/BrainSwapNetwork_test.go
cd ${DIR}/v0 && go test -v $SIM_TEST #> out1.txt
