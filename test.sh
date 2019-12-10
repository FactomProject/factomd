#!/usr/bin/env bash

# this script is specified as the 'tests' task in .circleci/config.yml

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)" # get dir containing this script
cd $DIR                                                             # always from from script dir

# set error on return if any part of a pipe command fails
set -o pipefail

# base go test command
GO_TEST="go test -v -timeout=10m -vet=off"

$GO_TEST ./longTest/... -run=TestLeaderModule
