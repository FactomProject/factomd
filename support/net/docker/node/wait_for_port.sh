#!/bin/bash

port="$1"
timeout="${2:-0}"
waited=0
delay=2

function try () {
    timeout 1 bash -c "$1"
}

until try "> /dev/tcp/127.0.0.1/$port > /dev/null 2> /dev/null"; do
  if [ "$timeout" -gt 0 ] && [ "$waited" -ge "$timeout" ]; then
    echo "failed"
    exit 1
  fi

  sleep 1
  waited=$(($waited + $delay))
done

echo "ok"
