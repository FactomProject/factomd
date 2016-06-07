if [ -z "$1" ]; then
    	echo "Usage:"
    	echo "    ./processdata.sh <test name>"
else
	gawk -f process.awk $1.out > $1.process.out
fi
