awk ' /put/ {foo[$5]=$0;} /delete/ {delete foo[$5];} /cleanup/ {delete foo[$5];} END {for (i in foo) {print foo[i];}}' $* | sort -n
