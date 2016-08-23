#!/usr/bin/env bash

factom-cli2 deletetransaction t

for i in `seq 1 100`;
do

    factom-cli2 newtransaction t
    factom-cli2 addinput t b 1
    factom-cli2 addoutput t b1 1
    factom-cli2 addfee t b
    factom-cli2 sign t
    factom-cli2 submit t

    factom-cli2 newtransaction t
    factom-cli2 addinput t b 1
    factom-cli2 addoutput t b2 1
    factom-cli2 addfee t b
    factom-cli2 sign t
    factom-cli2 submit t

    factom-cli2 newtransaction t
    factom-cli2 addinput t b 1
    factom-cli2 addoutput t b3 1
    factom-cli2 addfee t b
    factom-cli2 sign t
    factom-cli2 submit t

    factom-cli2 newtransaction t
    factom-cli2 addinput t b 1
    factom-cli2 addoutput t b4 1
    factom-cli2 addfee t b
    factom-cli2 sign t
    factom-cli2 submit t

done
sleep 3
echo "Balance of b1 should be 65"
factom-cli2 balances