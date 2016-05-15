#!/bin/bash

	factom-cli deletetransaction t

	factom-cli newtransaction t
	factom-cli addinput t factoid-wallet-address-name01 .40025
	factom-cli addoutput t b .00025
	factom-cli addecoutput t e1 .4
	factom-cli addfee t factoid-wallet-address-name01
	factom-cli sign t
	factom-cli submit t

sleep 5
for i in `seq 1 2`;
do
	sleep 3
	factom-cli deletetransaction t
	factom-cli newtransaction t
	factom-cli addinput t factoid-wallet-address-name01 .0001
	factom-cli addoutput t b1 .0001
	factom-cli addfee t factoid-wallet-address-name01
	factom-cli sign t
	factom-cli submit t
done


