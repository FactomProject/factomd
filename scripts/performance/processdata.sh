if [ -z "$1" ]; then
    	echo "Usage:"
    	echo "    ./processdata.sh <test name>"
else
	gawk -f performance.awk ../../out.txt > $1.performance.out
    gawk -f processdata.awk $1.out.txt > $1.process.out
	echo "Real Time (hours)	Real Time (hours)	%CPU	%MEM	% of One Core	Memory" >> $1.process.out
	gawk -f process.awk $1.out >> $1.process.out
fi
