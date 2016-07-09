#!/bin/bash

mychain=$loadrun

sleep 1
chain="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli2 mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
sleep 120
echo xxxxxxxxxxxxxxxxxxxx Made Chain number: $mychain chainid: ${chain}


for i in `seq 1 60`;
do
    sleep 1
    entry="$(echo "                Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli2 put -c ${chain} e1
    echo "Writing:             " ${entry} ${chain}
    echo chain $mychain run $i 1
    entry="$(echo "                Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli2 put -c ${chain} e1
    echo "Writing:             " ${entry} ${chain}
    echo chain $mychain run $i 2
    entry="$(echo "                Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli2 put -c ${chain} e1
    echo "Writing:             " ${entry} ${chain}
    echo chain $mychain run $i 3
    sleep 200
    entry="$(echo "                Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli2 put -c ${chain} e1
    echo "Writing:             " ${entry} ${chain}
    echo chain $mychain run $i 4
    entry="$(echo "                Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli2 put -c ${chain} e1
    echo "Writing:             " ${entry} ${chain}
    echo chain $mychain run $i 5
    entry="$(echo "                Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli2 put -c ${chain} e1
    echo "Writing:             " ${entry} ${chain}
    echo chain $mychain run $i 6
    sleep 200
    entry="$(echo "                Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli2 put -c ${chain} e1
    echo "Writing:             " ${entry} ${chain}
    echo chain $mychain run $i 7
    entry="$(echo "                Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli2 put -c ${chain} e1
    echo "Writing:             " ${entry} ${chain}
    echo chain $mychain run $i 8
    entry="$(echo "                Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli2 put -c ${chain} e1
    echo "Writing:             " ${entry} ${chain}
    echo chain $mychain run $i 9

done
