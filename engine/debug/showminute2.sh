#!/usr/bin/env bash

# showminute.sh <height> <minute> <node>

if [ "$#" -gt 3 || "$#" -lt 2  ]; then
    echo "showminute.sh <height> <minute> <node>"
fi

height=$1
shift
minute=$1
shift
fnode=$1
shift

pattern=`grep -E "done.*DBh/VMh/h $height/./-- minute $minute" ${fnode:-fnode0}_processlist.txt | grep -Eo "H-[0-9a-f]{6}" | awk ' NR==1 {printf("%s",$1);next;} {printf("|%s",$1);}'`
echo "EOMs " $pattern

grep -HE $pattern ${fnode:-fnode0}_networkinputs.txt ${fnode:-fnode0}_executemsg.txt ${fnode:-fnode0}_process.txt ${fnode:-fnode0}_processlist.txt | grep -v Drop | grep -iE "enq|xecu|retry|process.txt.*done|add" | ./merge.sh | less
