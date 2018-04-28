#!/usr/bin/env bash
reset
if [[ -z $1 ]]; then
file=out.txt
else
file=$1
fi
tail -f $file | gawk -f scripts/ProcessList.awk
