#!/usr/bin/env bash

# this script is specified in .circleci/config.yml
# to run as the 'tests' task
#

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR
export FACTOM_HOME=/dev/shm
export FACTOM_DEBUG_LOG_PATH=/dev/shm/logs
mkdir $FACTOM_DEBUG_LOG_PATH

function runTests() {


  if [[ "${CI}x" ==  "x" ]] ; then
    TESTS=$({ \
      glide nv | grep -v Utilities | grep -v longTest | grep -v peerTest | grep -v simTest; \
      #ls simTest/*_test.go;\
      #ls peerTest/*_test.go;\
    })
  else
    TESTS=$({ \
      glide nv | grep -v Utilities | grep -v longTest | grep -v peerTest | grep -v simTest; \
      circleci tests glob 'simTest/*_test.go'; \
      circleci tests glob 'peerTest/*Follower_test.go'; \
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
  #   BrainSwapFollower_test.go
  #   BrainSwapNetwork_test.go
  FOLLOWER="Follower"
  NETWORK="Network"
  FAIL=""

  for TST in ${TESTS[*]} ; do
    if [[ `dirname ${TST}` == "peerTest" ]] ; then
      NETWORK_TEST=${TST/$FOLLOWER/$NETWORK}
      TST=${TST/$NETWORK/$FOLLOWER}
      echo "Concurrent Peer TEST: $NETWORK_TEST"
      nohup go test -v -vet=off $NETWORK_TEST
    fi

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
