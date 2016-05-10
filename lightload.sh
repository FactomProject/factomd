#!/bin/bash

while true; do
	factom-cli balances
	sleep 3
	echo "Factoid Transaction"
	./flight.sh
	factom-cli balances
	sleep 3
	./eclight.sh
	sleep 3
done
