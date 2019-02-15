#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

function start_node() {
  echo "NOTE: you must manually create authoriy set using simulator"
  echo "EX: g7 <enter> 1 <enter> L <enter> L <enter> L <enter> o enter> o <enter> o <wait> z"

  ./factomd \
      --factomhome=$DIR \
			--network=LOCAL \
			--db=Map \
			--blktime=30 \
			--net=alot+ \
			--enablenet=true \
			--count=7 \
			--startdelay=1 \
			--stdoutlog=out.txt \
			--stderrlog=out.txt \
			--checkheads=false \
			--controlpanelsetting=readwrite \
			--logPort=38000 \
			--port=38001 \
			--controlpanelport=38002 \
			--networkport=38003 \
			--peers=127.0.0.1:37003 \
			--debuglog=".*" > out1.txt
}

function copy() { # and set AcksHeight
  cat "../../simConfig/factomd00${1}.conf" | sed 's/ChangeAcksHeight = 0/ChangeAcksHeight = 1/' > "${DIR}/.factom/m2/simConfig/factomd00${1}.conf"
}

# set identity for all nodes
function config() {
  mkdir -p $DIR/.factom/m2/simConfig

  copy 1
  copy 2
  copy 3
  copy 4
  copy 5
  copy 6
}

function clean() {
  rm *.txt 
  rm -rf .factom
}

function main() {
  clean
  config
  start_node
}

main
