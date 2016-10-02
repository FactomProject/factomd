#!/usr/bin/env bash
# Add some funds to our wallet.
factom-cli2 importaddress b Fs1fbEjDxWhXasmnp6R7DMFRo1oQBCJaa61hqkQQqhVGE8oJWKFE

# Make us some Factom Addresses
factom-cli2 generateaddress fct b1
factom-cli2 generateaddress fct b2
factom-cli2 generateaddress fct b3
factom-cli2 generateaddress fct b4

# make us an EC address
factom-cli2 generateaddress ec e1

factom-cli2 newtransaction t
factom-cli2 addinput t b 1
factom-cli2 addecoutput t e1 1
factom-cli2 addfee t b
factom-cli2 sign t
factom-cli2 submit t

for i in `seq 1 100`;
do

    chain1="$(echo "Chain" rand1: $RANDOM rand2: $RANDOM rand3: $RANDOM | factom-cli2 mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
    echo $chain1
    chain2="$(echo "Chain" rand1: $RANDOM rand2: $RANDOM rand3: $RANDOM | factom-cli2 mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
    echo $chain2
    chain3="$(echo "Chain" rand1: $RANDOM rand2: $RANDOM rand3: $RANDOM | factom-cli2 mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
    echo $chain3

    sleep 60

    for i in `seq 1 300`;
    do
     echo $RANDOM $RANDOM "entry" | factom-cli2 put -c ${chain1} e1 &
     echo $RANDOM $RANDOM "entry" | factom-cli2 put -c ${chain2} e1 &
     echo $RANDOM $RANDOM "entry" | factom-cli2 put -c ${chain3} e1 &
     sleep 0.3s
    done

    factom-cli2 balances
done