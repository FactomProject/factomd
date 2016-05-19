#!/bin/bash

while true; do
	
	echo "Factoid Transaction"

	for i in `seq 1 3`;
	do
		factom-cli balances
		./flight.sh
		 factom-cli balances
		 ./eclight.sh
	 done
done
