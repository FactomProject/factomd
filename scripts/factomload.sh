#!/bin/bash

	for j in `seq 1 1000000`;
	do
	    echo "Date: $(date)"
		factom-cli2 balances
		if [ $? -eq 0 ]; then
		    export loadrun=$j
    		./scripts/factomTransSubmit.sh
        else
            echo "balances failed"
            sleep 10
        fi
	done
