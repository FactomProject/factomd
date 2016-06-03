#!/bin/bash

	echo "Transactions: "

for i in `seq 1 10`;
do
	factom-cli newtransaction t
	factom-cli addinput t b .000005
	factom-cli addoutput t b1 .000001
	factom-cli addoutput t b2 .000001
	factom-cli addecoutput t e1 .000003
	factom-cli addfee t b
	factom-cli sign t
	factom-cli transactions
	factom-cli submit t
    sleep 4
    scripts/factomEC.sh &
done
sleep 2
