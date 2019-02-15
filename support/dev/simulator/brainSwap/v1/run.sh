#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

function start_node() {
  echo "starting as follower"
  ./factomd \
      --factomhome=$DIR \
			--network=LOCAL \
			--db=Map \
			--blktime=30 \
			--net=alot+ \
			--enablenet=true \
			--count=1 \
			--startdelay=1 \
			--stdoutlog=out.txt \
			--stderrlog=out.txt \
			--checkheads=false \
			--controlpanelsetting=readwrite \
			--logPort=37000 \
			--port=37001 \
			--controlpanelport=37002 \
			--networkport=37003 \
			--peers=127.0.0.1:38003 \
			--debuglog=".*" > out1.txt
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
  rm *.txt 
  rm -rf .factom
}

function main() {
  clean
  config
  start_node
}

main
