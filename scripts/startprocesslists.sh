#!/usr/bin/env bash
if [[ -z $1 ]]; then
file=out.txt
else
file=$1
fi

reset
tail -f out.txt | gawk -f scripts/ProcessList.awk
