#!/usr/bin/env bash
pid="$(ps ax | grep "[0-9] factomd" | gawk "{print \$1}")"
echo "pid=" $pid
python scripts/log.py $pid > time.out &
tail -f time.out
