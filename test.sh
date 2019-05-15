#!/usr/bin/env bash

# this script is specified in .circleci/config.yml
# to run as the 'tests' task

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )" # get dir containing this script
cd $DIR # always from from script dir

function runTests() {
  if [[ "${CI}x" ==  "x" ]] ; then
    # run locally
    TESTS=$({ \
      glide nv | grep -v Utilities | grep -v longTest | grep -v peerTest | grep -v simTest | grep -v elections | grep -v activations | grep -v netTest | grep "..." ; \
      cat */ci_whitelist; \
    })
  else
    # run on circle
    TESTS=$({ \
      glide nv | grep -v Utilities | grep -v longTest | grep -v peerTest | grep -v simTest | grep -v elections | grep -v activations | grep -v netTest | grep "..." ; \
      cat */ci_whitelist; \
    } | circleci tests split --split-by=timings)
  fi

	if [[ "${TESTS}x" ==  "x" ]] ; then
    echo "No Tests"
    exit 0
  else
    echo '---------------'
    echo "${TESTS}"
    echo '---------------'
  fi

  # NOTE: peer tests are expected to be named 
  # in Follower/Network pairs
  # Example:
  #   BrainSwapA_test.go # 'A' runs first in background
  #   BrainSwapB_test.go # 'B' test runs in foreground
  BTEST="B_"
  ATEST="A_"
  FAILURES=()
  FAIL=""

  # exit code should fail if any part of the command fails
  set -o pipefail

  for TST in ${TESTS[*]} ; do
    # start 'A' part of A/B test in background
    if [[ `dirname ${TST}` == "peerTest" ]] ; then
      ATEST_FILE=${TST/$BTEST/$ATEST}
      TST=${TST/$ATEST/$BTEST}
      echo "Concurrent Peer TEST: $ATEST_FILE"
      nohup go test -v -timeout=10m -vet=off $ATEST_FILE &> testout.txt &
    fi

    # run individual sim tests that have been whitelisted 
    if [[ `dirname ${TST}` == "engine" && ${TST/engine\//} != '...' ]] ; then
      TST="./engine/... -run ${TST/engine\//}"
      echo "Testing: $TST"
    fi

    echo "START: ${TST}"
    echo '---------------'
    go test -v -timeout=10m -vet=off $TST | tee -a testout.txt | egrep 'PASS|FAIL|RUN' 
    if [[ $? != 0 ]] ;  then
      FAIL=1
      FAILURES+=($TST)
    fi
    echo "END: ${TST}"
    echo '---------------'

  done

  if [[ "${FAIL}x" != "x" ]] ; then
    echo "TESTS FAIL"
    echo '---------------'
    for F in ${FAILURES[*]} ; do
      echo $F
    done
    exit 1
  else
    echo "ALL TESTS PASS"
    exit 0
  fi
}

runTests
