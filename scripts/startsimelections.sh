#!/usr/bin/env bash
if [[ -z $1 ]]; then
file=out.txt
else
file=$1
fi

reset
tail -f $file | gawk -f scripts/simelections.awk
