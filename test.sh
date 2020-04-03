#!/usr/bin/env bash

# this script is specified as the 'tests' task in .circleci/config.yml

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)" # get dir containing this script
cd $DIR                                                             # always from from script dir

# set error on return if any part of a pipe command fails
set -o pipefail

# base go test command
GO_TEST="go test -v -timeout=10m -vet=off"

# list modules for CI testing
function listModules() {
	# FIXME: likely a cleaner way to exclude some packages
	#glide nv | grep -v Utilities | grep -v log | grep -v pubsub | grep -v simulation | grep -v modules | grep -v elections | grep -v longTest | grep -v peerTest | grep -v simTest | grep -v activations | grep -v netTest | grep -v factomgenerate | grep "\.\.\."
	ls -l | grep ^d | awk '{ print("./"$9"/...") }' | grep -v Utilities | grep -v log | grep -v pubsub | grep -v simulation | grep -v modules | grep -v elections | grep -v longTest | grep -v peerTest | grep -v simTest | grep -v activations | grep -v netTest | grep -v factomgenerate
}

# formatted list of simTest/<testname>
function listSimTest() {
	go test --list=Test ./simTest/... | awk '/^Test/ { print "simTest/"$1 }'
}

# list of A/B Peer tests that need to be run concurrently
function listPeer() {
	ls peerTest/*A_test.go
}

# read a list of tests from a file called .ci_tests
function hardcodedList() {
	cat .circleci/ci_tests
}

# load a list of tests to execute
function loadTestList() {
	case $1 in
	unittest) # run unit test in batches - manually triggered
		TESTS=$({
			listModules
		})
		;;
	peertest) # run only peer tests - manually triggered
		TESTS=$({
			ls peerTest/*A_test.go
		})
		;;

	simtest) # run only simulation tests - manually triggered

		TESTS=$({
			listSimTest
		})
		;;

	short) # triggered on every commit

		# run on circle
		if [[ "${CI}x" != "x" ]]; then
			TESTS=$({
				listModules
				echo "simTest/TestAnElection"
			} | circleci tests split) # circleci helper spreads tests across containers

		else # run locally
			TESTS=$({
				listModules
				echo "simTest/TestAnElection"
			})

		fi
		;;

	full) # run on circle triggered on develop branch nightly

		if [[ "${CI}x" != "x" ]]; then

			# Just run failing tests if .circleci/ci_tests file is present
			if [[ -f .circleci/ci_tests ]]; then
				TESTS=$({
					hardcodedList # limit the tests to a hardcoded list
				} | circleci tests split)
			else
				TESTS=$({
					listModules
					listSimTest
					listPeer
				} | circleci tests split) # circleci helper spreads tests across containers
			fi

		else # run locally

			TESTS=$({
				listModules
				listSimTest
				listPeer
			})

		fi

		;;

	*)
		echo "Unknown option '$1'"
		echo "usage: ./test.sh [unittest|peertest|simtest|short|full]"
		exit -1
		;;
	esac
}

function testGoFmt() {
	FILES=$(find . -name '*.go' ! -name '*_template.go')

	for FILE in ${FILES[*]}; do
		gofmt -w $FILE

		if [[ $? != 0 ]]; then
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

	for TST in ${TESTS[*]}; do
		case $(dirname $TST) in
		simTest)
			testSim $TST
			;;
		peerTest)
			testPeer $TST
			;;
		*) # package name provided instead
			unitTest $TST
			;;
		esac

		if [[ $? != 0 ]]; then
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
	nohup $GO_TEST $A &>a_testout.txt &

	# run part B in foreground
	$GO_TEST $B &>b_testout.txt
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

function writeTestList() {
    > .circleci/ci_tests
	for TST in ${TESTS[*]}; do
        echo $TST >> .circleci/ci_tests
    done
    echo "Wrote .circleci/ci_tests"
}

function main() {
	FAILURES=()
	FAIL=""

	case $1 in

	mklist)
        # writelist of all tests to a file
        # so full test run can be customized for a given build
	    loadTestList full
        writeTestList
        ;;
	gofmt)
		# check all go files pass gofmt
		testGoFmt
		;;
	*)
		if [[ $CIRCLE_BRANCH =~ _ci$ ]]; then
			# if branch name ends in _ci
			# force 'full' test run
			runTests full
		else
			# otherwise run tests as specified
			runTests $1
		fi
		;;
	esac

	if [[ "${FAIL}x" != "x" ]]; then
		echo "TESTS FAIL"
		echo '---------------'
		for F in ${FAILURES[*]}; do
			echo $F
		done
		exit 1
	else
		echo "ALL TESTS PASS"
		exit 0
	fi
}

main $1
