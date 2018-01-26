#!/usr/bin/env bash
if [[ -z $1 ]]; then
file=out.txt
else
file=$1
fi

reset
tail -n 9999999 -f $1 | gawk -f scripts/ProcessList.awk