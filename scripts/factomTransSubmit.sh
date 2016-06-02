#!/bin/bash

	echo "Transactions: "

for i in `seq 1 1`;
do
	factom-cli newtransaction t
	factom-cli addinput t factoid-wallet-address-name01 .000005
	factom-cli addoutput t b .000001
	factom-cli addoutput t b2 .000001
	factom-cli addecoutput t e1 .000003
	factom-cli addfee t factoid-wallet-address-name01
	factom-cli sign t
	factom-cli transactions
	factom-cli submit t

    scripts/factomEC.sh &
done

sleep 5
