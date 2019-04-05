#!/usr/bin/env bash

# this script is specified in .circleci/config.yml
# to run as the 'tests' task
#

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

function runTests() {


  if [[ "${CI}x" ==  "x" ]] ; then
    TESTS=$({ \
      glide nv | grep -v Utilities | grep -v longTest | grep -v peerTest | grep -v simTest; \
      ls simTest/*_test.go; \
      ls peerTest/*_test.go; \
    })
  else
    TESTS=$({ \
      glide nv | grep -v Utilities | grep -v longTest | grep -v peerTest | grep -v simTest; \
      circleci tests glob 'simTest/*_test.go'; \
      circleci tests glob 'peerTest/*A_test.go'; \
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
  FAIL=""

  for TST in ${TESTS[*]} ; do
    if [[ `dirname ${TST}` == "peerTest" ]] ; then
      ATEST_FILE=${TST/$BTEST/$ATEST}
      TST=${TST/$ATEST/$BTEST}
      echo "Concurrent Peer TEST: $ATEST_FILE"
      nohup go test -v -timeout=30m -vet=off $ATEST_FILE &
    fi

    go test -v -timeout=30m -vet=off $TST

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
