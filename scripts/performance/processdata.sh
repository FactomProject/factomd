if [ -z "$1" ]; then
    	echo "Usage:"
    	echo "    ./processdata.sh blktime=120 <test name>"
else
	gawk -f process.awk $1 $2.out > $2.process.out
fi
