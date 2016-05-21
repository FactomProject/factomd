#!/bin/bash

while true ;
do
	for j in `seq 1 10`;
	do
		factom-cli balances
		if [ $? -eq 0 ]; then
    		./scripts/flight.sh
        else
            sleep 10
        fi
	done
done
