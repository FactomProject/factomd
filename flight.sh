#!/bin/bash


for i in `seq 1 16`;
do
	sleep 8
	factom-cli newtransaction t
	factom-cli addinput t factoid-wallet-address-name01 .013
	factom-cli addecoutput t e1 .013
	factom-cli addfee t factoid-wallet-address-name01
	factom-cli sign t
	factom-cli submit t
done

