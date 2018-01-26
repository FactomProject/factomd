#!/usr/bin/env bash
if [[ -z $1 ]]; then
file=out.txt
else
file=$1
fi

reset
tail -f -n 9999999 $file | gawk -f scripts/status.awk
