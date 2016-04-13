#!/bin/bash
cd .. 
echo "Compiling..."
go install -a 
echo "Running..."
if [ $? -eq 0 ]; then
    factomd -count=1 -net=tree -folder="test1-" -port=8089 -serve=40899 -follower=true & node0=$!
    factomd -count=1 -net=tree -folder="test2-" -port=8090 -connect="tcp://217.0.0.1:40899" -leader=true & node1=$!
    sleep 60
    kill $node0 $node1
fi
