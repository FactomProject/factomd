#!/bin/bash

	echo "Transactions: "
sleep 1
for i in `seq 1 100`;
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

	number1=$RANDOM
	number2=$RANDOM
	echo "Make Chain Named " $number
//	echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $number1 -e $number2 e1 &
done


