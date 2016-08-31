#!/bin/bash

p=.01
for i in `seq 1 3`;
do
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain1} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain2} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain3} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain4} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain5} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain6} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain7} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain8} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain9} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain10} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain11} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain12} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain13} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain14} e1
    echo ${entry}
    sleep $p
    entry="$(echo "Here are some random numbers: " $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM $RANDOM)"
    echo ${entry} | factom-cli put -c ${chain15} e1
    echo ${entry}
done
