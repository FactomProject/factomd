#!/bin/bash

factom-walletd &

trap_with_arg() { # from https://stackoverflow.com/a/2183063/804678
  local func="$1"; shift
  for sig in "$@"; do
    trap "$func $sig" "$sig"
  done
}

stop() {
  trap - SIGINT EXIT
  printf '\n%s\n' "recieved $1, killing children"
  kill -s SIGINT 0
}

trap_with_arg 'stop' EXIT SIGINT SIGTERM SIGHUP
for ((i=0; i<100000; i++)); do
	scripts/entryloadm2.sh
done

