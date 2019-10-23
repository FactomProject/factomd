#!/usr/bin/env bash
# show the details of how a node processed a given minute
#showminute.sh <dbht> <minute> <node(s)>
block="$1"
shift
minute="$1"
shift
(
grep -E ".*DBh/VMh/h $block/./-- minute $minute" $1_networkinputs.txt | grep -v Drop; 
grep -E ".*DBh/VMh/h $block/./-- minute $minute" $1_executemsg.txt; 
grep -E "(Add|done).*DBh/VMh/h $block/./-- minute $minute" $1_processlist.txt; 
grep -E "$block-:-.*ProcessEOM complete for $minute" $1_dbsig-eom.txt; 
grep -E "Send.*DBh/VMh/h $block/./-- minute $minute" $1_networkoutputs.txt
) | sort -n
