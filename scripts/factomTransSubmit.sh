#!/bin/bash

	echo "Transactions: "

for i in `seq 1 1`;
do
    export loadrun
	factom-cli2 newtransaction t
	factom-cli2 addinput t b .000031
	factom-cli2 addoutput t b1 .000001
	factom-cli2 addecoutput t e1 .000030
	factom-cli2 addfee t b
	factom-cli2 sign t
	factom-cli2 submit t
	scripts/factomEC.sh &
	sleep 10
done

