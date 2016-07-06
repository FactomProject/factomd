#!/bin/bash

while true ;
do
	for j in `seq 1 10`;
	do
	    echo "Date: $(date)"
		factom-cli2 balances
		if [ $? -eq 0 ]; then
    		./scripts/factomTransSubmit.sh
        else
            sleep 10
        fi
	done
done
