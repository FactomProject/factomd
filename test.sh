#!/usr/bin/env bash

# this script is specified as the 'tests' task in .circleci/config.yml

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )" # get dir containing this script
cd $DIR # always from from script dir

# set error on return if any part of a pipe command fails
set -o pipefail

# base go test command
GO_TEST="go test -v -timeout=10m -vet=off"

# list modules for CI testing
function listModules() {
  glide nv | grep -v Utilities | grep -v elections | grep -v longTest | grep -v peerTest | grep -v simTest | grep -v activations | grep -v netTest | grep "\.\.\."
}

# formatted list of simTest/<testname>
function listSimTest() {
  go test --list=Test ./simTest/... | awk '/^Test/ { print "simTest/"$1 }'
}

# list of A/B Peer tests
function listPeer() {
  ls peerTest/*A_test.go
}

# load a list of tests to execute
function loadTestList() {
  case $1 in
    unittest ) # run unit test in batches
      TESTS=$({ \
        listModules ; \
      })
      ;;
    peertest ) # run only peer tests
      TESTS=$({ \
        ls peerTest/*A_test.go; \
      })
      ;;

    simtest ) # run only simulation tests
      TESTS=$({ \
          listSimTest ; \
      })
      ;;

    "" ) # no param: run everything

      # running locally - default to all tests
      if [[ "${CI}x" ==  "x" ]] ; then

          TESTS=$({ \
            listModules ; \
            listSimTest ; \
            listPeer ; \
          })

      else # running on circle

        # run all tests 
        if [[ "${CIRCLE_TAG}${CIRCLE_PULL_REQUEST}x" !=  "x" ]] ; then

          TESTS=$({ \
            listModules ; \
            listSimTest ; \
            listPeer ; \
          } | circleci tests split ) # circleci helper spreads tests across containers

        else # run single sim + all unit tests on every commit

          TESTS=$({ \
            listModules ; \
            echo "simTest/TestAnElection" ; \
          } | circleci tests split ) # circleci helper spreads tests across containers

        fi
      fi
      ;;

    * )
      echo "Unknown option" $1
      exit -1
      ;;
  esac
}

function testGoFmt() {
  FILES=$( find . -name '*.go')

  for FILE in ${FILES[*]} ; do
    gofmt -w $FILE

    if [[ $? != 0 ]] ;  then
      FAIL=1
      FAILURES+=($FILE)
    fi
  done

}

function runTests() {
  loadTestList $1

  echo '---------------'
  echo "${TESTS}"
  echo '---------------'

  for TST in ${TESTS[*]} ; do
    case `dirname $TST` in
      simTest )
        testSim $TST
        ;;
      peerTest )
        testPeer $TST
        ;;
      * ) # package name provided instead
        unitTest $TST
        ;;
    esac

    if [[ $? != 0 ]] ;  then
      FAIL=1
      FAILURES+=($TST)
    fi
  done
}

# run A/B peer coodinated tests
# $1 should be a path to a test file
function testPeer() {
  A=${1/B_/A_}
  B=${1/A_/B_}

  # run part A in background
  nohup $GO_TEST $A &> a_testout.txt &

  # run part B in foreground
  $GO_TEST $B &> b_testout.txt
}

# run unit tests per module this ignores all simtests
function unitTest() {
  $GO_TEST $1 | egrep "PASS|FAIL|panic|bind|Timeout"
}

# run a simtest
# $1 matches simTest/<TestSomeTestName>
function testSim() {
  $GO_TEST -run=${1/simTest\//} ./simTest/... | egrep "PASS|FAIL|panic|bind|Timeout"
}

function main() {
  FAILURES=()
  FAIL=""

  if [[ "${1}" == "gofmt" ]] ; then
    # check all go files pass gofmt
    testGoFmt
  else
    # run tests
    runTests $1
  fi

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

main $1
