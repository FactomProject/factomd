#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR


function copy() { # and set AcksHeight

  if [[ $2 = 0 ]] ; then
    target="${DIR}/.factom/m2/factomd.conf"
  else
    target="${DIR}/.factom/m2/simConfig/factomd00${2}.conf"
  fi

  cat "../../simConfig/factomd00${1}.conf" \
    | sed 's/ChangeAcksHeight = 0/ChangeAcksHeight = 10/' > $target
}

function main() {
  copy 1 0
}

main
