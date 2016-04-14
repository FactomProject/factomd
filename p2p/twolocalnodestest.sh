#!/bin/bash
echo "run this from factomproject/factomd eg:"
echo "$ ./p2p/twolocalnodestest.sh"
echo
echo "Compiling..."
go install -a 
echo "Running..."
if [ $? -eq 0 ]; then
    factomd -count=1 -folder="test1-" -port 9123 -address="tcp://127.0.0.1:40891" -follower=true & node0=$!
    factomd -count=1 -folder="test2-" -port 9124 -address="tcp://127.0.0.1:40891" -leader=true & node1=$!
    sleep 60
    kill $node0 $node1
fi