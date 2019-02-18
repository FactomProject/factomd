#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
SIM_TEST=${DIR}/../../../../../simTest/BrainSwapFollower_test.go

function start_node() {
  FACTOM_HOME=$DIR go test -v $SIM_TEST #> out1.txt
}

function copy() { # and set AcksHeight

  if [[ $2 = 0 ]] ; then
    target="${DIR}/.factom/m2/factomd.conf"
  else
    target="${DIR}/.factom/m2/simConfig/factomd00${2}.conf"
  fi

  cat "../../simConfig/factomd00${1}.conf" \
    | sed 's/ChangeAcksHeight = 0/ChangeAcksHeight = 10/' > $target
}

function config() {
  # copy config files
  mkdir -p $DIR/.factom/m2/simConfig
  copy 9 0
  copy 8 1
}


function clean() {
  rm *.txt  2>/dev/null
  rm -rf .factom
}

function main() {
  cd $DIR
  clean
  config
  start_node
}

main
