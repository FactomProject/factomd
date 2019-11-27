#!/usr/bin/env bash
# extract minute length and commits per minute into a file for graphing

if [ "$#" -gt 1  ]; then
    echo "bootAnalysis.sh <node> "
fi

fnode=$1
shift

echo "bootAnalysis.sh" ${fnode:-fnode0}


(grep -E "done.*EOM.*./00?/" ${fnode:-fnode0}_processlist.txt  | minutelength.sh | awk '/Gaps/ {exit(0);} {printf("t %s-:-%-2d %s\n",$12+0,$14,$1);}' ; 
 grep -E "done.*Commit" ${fnode:-fnode0}_processlist.txt | awk ' {hits[$3]++;} END{for(i in hits){print "c", i, hits[i];}}' | sort) | 
 awk ' BEGIN {print "MINUTE\tDURATION\tCOMMITS\tRATE";} /t/{t[$2]=$3;m[$2]++;} /c/ {c[$2]=$3;m[$2]++;} END {for(i in m){if(t[i]!=0) {printf("%s\t%9.3f\t%5d\t%7.1f\n", i,t[i]+0.0,c[i]+0,c[i]/(t[i]+.000001));}}}' | sort -n
