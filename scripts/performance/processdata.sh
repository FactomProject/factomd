if [ -z "$1" ]; then
    	echo "Usage:"
    	echo "    ./processdata.sh <test name>"
else
        grep "FNode0\[" $1.out.txt | tail -n 1 | gawk "BEGIN{FS=\"[ [/]+\"}{print \$10 \"\\t\" \$26 \"\\t\" \$27 \"\\t\" \$28}" > $1.process.out
	echo "Real Time (days)	Time (Days)	%CPU	%MEM	% of One Core	Memory" >> $1.process.out
	gawk -f process.awk $1.out >> $1.process.out

fi
