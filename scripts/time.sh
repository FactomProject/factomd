#!/usr/bin/env bash

tail -f time.out &

for j in `seq 1 1000000`;
do
		echo "Timer", $j
		pid="$(ps ax | grep "[0-9] factomd" | gawk "{print \$1}")"
		echo "pid=" $pid
		python scripts/log.py $pid > time.out 
		sleep 10
done

