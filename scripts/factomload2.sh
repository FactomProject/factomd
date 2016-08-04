#!/bin/bash

    factom-cli2 importaddress b Fs1fbEjDxWhXasmnp6R7DMFRo1oQBCJaa61hqkQQqhVGE8oJWKFE

	for j in `seq 0 1000000`;
	do
	    echo "Date: $(date)"
		factom-cli2 balances
		if [ $? -eq 0 ]; then
		    export loadrun=$j
    		./scripts/factomTransSubmit2.sh
        else
            echo "balances failed"
            sleep 10
        fi
	done
