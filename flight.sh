#!/bin/bash

	echo "Transactions: "
	factom-cli deletetransaction t

	factom-cli newtransaction t
	factom-cli addinput t factoid-wallet-address-name01 .10025
	factom-cli addoutput t b .00025
	factom-cli addecoutput t e1 .1
	factom-cli addfee t factoid-wallet-address-name01
	factom-cli sign t
	factom-cli transactions
	factom-cli submit t

sleep 5
for i in `seq 24`;
do
	sleep 3
	factom-cli deletetransaction t
	factom-cli newtransaction t
	factom-cli addinput t factoid-wallet-address-name01 .0001
	factom-cli addoutput t b1 .0001
	factom-cli addfee t factoid-wallet-address-name01
	factom-cli sign t
	factom-cli transactions
	factom-cli submit t
done


