if [ -z "$1" ]; then
    	echo "Usage:"
    	echo "    ./pulldata.sh <test name>"
else
	cp ../../time.out ./$1.out
	tail -c 1000000 ../../out.txt > ./$1.out.txt
fi
