#!/bin/bash

	echo "Transactions: " $loadrun
sleep 1
let s=loadrun*100
let f=$s+99
for i in `seq $s $f`;
do
    adr=b$i
    factom-cli2 generateaddress fct $adr > nul
	factom-cli2 newtransaction t > nul
	factom-cli2 addinput t b .000001 > nul
	factom-cli2 addoutput t $adr .000001 > nul
	factom-cli2 addfee t b > nul
	factom-cli2 sign t > nul
	factom-cli2 submit t > nul
	sleep .01
done

