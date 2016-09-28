#!/bin/bash

	echo "Transactions: "

for i in `seq 1 5`;
do
    export loadrun
	factom-cli2 newtransaction t
	factom-cli2 addinput t b .0061
	factom-cli2 addoutput t b1 .0001
	factom-cli2 addecoutput t e1 .006
	factom-cli2 addfee t b
	factom-cli2 sign t
	factom-cli2 submit t
	scripts/factomEC.sh &
done

let y=$(shuf -i 7-13 -n 1)
echo SLEEP $y  YAWN
sleep $y
