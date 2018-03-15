#!/bin/bash
echo "run this from factomproject/factomd eg:"
echo "$ ./p2p/single_test.sh"
echo
echo "Compiling..."
go install -a 
if [ $? -eq 0 ]; then
    pkill factomd
    echo "Running single node"
    factomd -count=2 -folder="test1-" -network="TEST" -networkPort=8118 -peers="127.0.0.1:8121" -db=Map & node0=$!
    echo
    sleep 480
    echo
    echo
    echo "Killing processes now..."
    echo
    echo "Process number is $node0"
    kill -2 $node0 
fi
