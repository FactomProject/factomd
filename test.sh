#!/usr/bin/env bash

# this script is specified in .circleci/config.yml
# to run as the 'tests' task
#

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

# run simluation tests in parallel
# NOTE: that this also runs the unit test suite
function simTests() {
	# NOTE: this command causes Circle.ci to run tests in parallel across several containers
	TESTS=$(circleci tests glob "simTest/*_test.go" | circleci tests split --split-by=timings)

	if [[ "${TESTS}x" ==  "x" ]] ; then
    # when there is not test argument run default tests
    echo '---------------'
    echo "UnitTests"
    echo '---------------'
    unitTests
	else
    echo '---------------'
    echo "${TESTS}"
    echo '---------------'

		go test -v -vet=off $TESTS
	fi
}

# run "safe" test suite in serial to avoid hitting memory limits on circle.ci containers
function unitTests() {

	PACKAGES=$(glide nv | grep -v Utilities | grep -v longTest | grep -v peerTest | grep -v simTest)
	FAIL=""

	for PKG in ${PACKAGES[*]} ; do
		go test -v -vet=off $PKG
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

function runOnCircle() {
	echo "runOnCircle: $1"

  case "$1" in
    sim)
        simTests
        ;;
    unit)
        unitTests
        ;;
    *)
        echo "Unknown test: '${1}'"
        exit -1
  esac
}

# run this script with no arguments and it will execute the unit test suite
# otherwise the argument selects which tests to run
if [[ "${CI}x" ==  "x" ]] ; then
	unitTests # run locally
else
	runOnCircle $1
fi
