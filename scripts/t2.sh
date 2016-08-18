#!/usr/bin/env bash

factom-cli2
factom-cli2 deletetransaction t
factom-cli2 newtransaction t
factom-cli2 addinput t b .2
factom-cli2 addoutput t b1 .2
factom-cli2 addfee t b
factom-cli2 sign t
factom-cli2 submit t

