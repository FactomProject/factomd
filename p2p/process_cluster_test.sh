#!/bin/bash
echo "run this from factomproject/factomd eg:"
echo "$ ./p2p/process_cluster_test.sh"
echo
echo "Compiling..."
go install -a 
if [ $? -eq 0 ]; then
    echo "Running..."
    factomd -count=2 -folder="test1-" -port=9120 -p2pPort="34340" -peers="127.0.0.1:34343" -netdebug=2 -db=Map & 
    sleep 6
   factomd -count=2 -prefix="test2-" -port=9121 -p2pPort="34341" -peers="127.0.0.1:34340" -netdebug=2 -db=Map &
    sleep 6
    factomd -count=2 -prefix="test3-" -port=9122 -p2pPort="34342" -peers="127.0.0.1:34341" -netdebug=2 -db=Map &
    sleep 6
    factomd -count=2 -prefix="test4-" -port=9123 -p2pPort="34343" -peers="127.0.0.1:34342" -netdebug=2 -db=Map &
    echo
    echo
    echo
    echo
    echo "####################################################################################################################"
    echo "####################################################################################################################"
    echo "####################################################################################################################"
    echo "####################################################################################################################"
    echo "####################################################################################################################"
    echo
    echo
    sleep 240
    echo
    echo
    echo "Killing processes now..."
    echo
    pkill factomd    
fi