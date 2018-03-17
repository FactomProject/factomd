#!/bin/bash
echo "run this from factomproject/factomd eg:"
echo "$ ./p2p/process_cluster_test.sh"
echo
echo "Compiling..."
go install 
if [ $? -eq 0 ]; then
    pkill factomd
    echo "Running..."
    factomd -exclusive=true -count=1 -folder="test1-" -network="TEST" -networkPort=8118 -peers="127.0.0.1:8121" -db=Map & node0=$!
    sleep 6
    factomd -exclusive=true -count=1 -prefix="test2-" -network="TEST" -port=9121 -networkPort=8119 -peers="127.0.0.1:8118" -db=Map & node1=$!
    sleep 6
    factomd -exclusive=true -count=1 -prefix="test3-" -network="TEST" -port=9122 -networkPort=8120 -peers="127.0.0.1:8119" -db=Map & node2=$!
    sleep 6
        echo "Launching 4th node, hopefully you have sim control now:"
    factomd -exclusive=true -count=1 -prefix="test4-" -network="TEST" -port=9123 -networkPort=8121  -peers="127.0.0.1:8120" -db=Map 
    echo
fi
