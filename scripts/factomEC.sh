#!/bin/bash

mychain=$loadrun

sleep 1
chain="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli2 mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo xxxxxxxxxxxxxxxxxxxx Made Chain number: $mychain chainid: ${chain}
let x=$(shuf -i 10-30 -n 1)
sleep $x

sleep 600

for i in `seq 1 12`;
do
    let y=$(shuf -i 3-10 -n 1)

    sleep $y
    entry="$(echo "                Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli2 put -c ${chain} e1
    echo "Writing:             " ${entry} ${chain}
    echo chain $mychain run $i 9

    sleep $y
    entry="$(echo "                Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli2 put -c ${chain} e1
    echo "Writing:             " ${entry} ${chain}
      echo chain $mychain run $i 9

    sleep $y
    entry="$(echo "                Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli2 put -c ${chain} e1
    echo "Writing:             " ${entry} ${chain}
    echo chain $mychain run $i 9

done
