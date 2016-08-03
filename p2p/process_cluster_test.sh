#!/bin/bash
echo "run this from factomproject/factomd eg:"
echo "$ ./p2p/process_cluster_test.sh"
echo
echo "Compiling..."
go install -ldflags "-X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD`"
if [ $? -eq 0 ]; then
    pkill factomd
    echo "Running..."
    factomd -count=1 -folder="test1-" -network="TEST" -networkPort=8118 -peers="127.0.0.1:8121" -netdebug=1 -db=Map & node0=$!
    sleep 6
    factomd -count=1 -prefix="test2-" -network="TEST" -port=9121 -networkPort=8119 -peers="127.0.0.1:8118" -netdebug=1 -db=Map & node1=$!
    # sleep 6
    # factomd -count=1 -prefix="test3-" -network="TEST" -port=9122 -networkPort=8120 -peers="127.0.0.1:8119" -netdebug=1 -db=Map & node2=$!
    # sleep 6
    # factomd -count=1 -prefix="test4-" -network="TEST" -port=9123 -networkPort=8121  -peers="127.0.0.1:8120" -netdebug=1 -db=Map & node3=$!
    echo
    echo
    sleep 480
    echo
    echo
    echo "Killing processes now..."
    echo
    # kill -2 $node0 $node1 $node2 $node3
    kill -2 $node1 # Kill this first to see how node0 handles it.
    sleep 25
    kill -2 $node0 $node2 $node3
fi