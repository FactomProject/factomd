#!/bin/bash

	echo "Transactions: "
for i in `seq 1 10`;
do
    sleep 1
	factom-cli deletetransaction t
	factom-cli newtransaction t
	factom-cli addinput t factoid-wallet-address-name01 .0003
	factom-cli addoutput t b .0001
	factom-cli addoutput t b2 .0001
	factom-cli addecoutput t e1 .0001
	factom-cli addfee t factoid-wallet-address-name01
	factom-cli sign t
	factom-cli transactions
	factom-cli submit t

	number1=$RANDOM
	number2=$RANDOM
	echo "Make Chain Named " $number
	echo "test" | factom-cli mkchain -e $number1 -e $number2 e1 &

done


