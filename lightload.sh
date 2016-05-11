#!/bin/bash

while true; do
	factom-cli balances
	sleep 10
	echo "Factoid Transaction"

	for i in `seq 1 3`;
	do
		./flight.sh
		./eclight.sh
		factom-cli balances	
	 done
done
