#!/bin/bash

sleep 1
for i in `seq 1 4`;
do
	sleep 4
	number1=$RANDOM
	number2=$RANDOM
	echo "Make Chain Named " $number
	echo "test" | factom-cli mkchain -e $number1 -e $number2 e1 &

done

