#!/bin/bash

files=""

if [ "$2" != "" ]; then
    files="$2"
fi

grep $1 $files*.txt | awk -f msgOrder.awk | sort -n
