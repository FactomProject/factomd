#!/usr/bin/env bash

# this script is specified in .circleci/config.yml
# to run as the 'tests' task
#

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

function runTests() {

  if [[ "${CI}x" ==  "x" ]] ; then
    TESTS=$(find . -name '*_test.go' \
      | grep -v vendor | grep -v Utilities | grep -v longTest | grep -v peerTest | grep -v simTest)
  else
    # NOTE: this command causes Circle.ci to run tests in parallel across several containers
    TESTS=$(find . -name '*_test.go' \
      | grep -v vendor | grep -v Utilities | grep -v longTest | grep -v peerTest | grep -v simTest \
      | circleci tests split --split-by=timings)
  fi

	if [[ "${TESTS}x" ==  "x" ]] ; then
    echo "No Tests"
    exit 0
  else
    echo '---------------'
    echo "${TESTS}"
    echo '---------------'
  fi

  FAIL=""
  for TST in ${TESTS[*]} ; do
    go test -v -vet=off $TST
    if [[ $? != 0 ]] ;  then
      FAIL=1
    fi
  done

  if [[ "${FAIL}x" != "x" ]] ; then
    echo "TESTS FAIL"
    exit 1
  else
    echo "ALL TESTS PASS"
    exit 0
  fi
}

runTests
