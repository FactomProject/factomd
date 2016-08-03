#!/usr/bin/env bash

fct deletetransaction t
fct newtransaction t
fct addinput t factoid-wallet-address-name01 .2
fct addoutput t b1 .2
fct addfee t factoid-wallet-address-name01
fct sign t
fct submit t

