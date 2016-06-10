#!/bin/bash

	echo "Transactions: "

for i in `seq 1 2`;
do
	factom-cli newtransaction t
	factom-cli addinput t b .000016
	factom-cli addoutput t b1 .000001
	factom-cli addoutput t b2 .000001
	factom-cli addoutput t b3 .000001
	factom-cli addoutput t b4 .000001
	factom-cli addoutput t b5 .000001
	factom-cli addoutput t b6 .000001
	factom-cli addecoutput t e1 .000010
	factom-cli addfee t b
	factom-cli sign t
	factom-cli transactions
	factom-cli submit t
	sleep 2

	scripts/factomEC.sh &
done

