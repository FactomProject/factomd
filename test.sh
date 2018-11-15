#!/usr/bin/env bash

# run same tests as specified in .circleci/config.yml

if [[ "$1" = "-l" ]] ; then
  echo "filtering test output to list of failed tests"
  go test -v -vet=off $(glide nv | grep -v Utilities) | grep ^---\ FAIL
else
  go test -v -vet=off $(glide nv | grep -v Utilities)
fi
