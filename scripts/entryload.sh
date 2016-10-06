#!/usr/bin/env bash
# Add some funds to our wallet.
factom-cli2 importaddress b Fs1fbEjDxWhXasmnp6R7DMFRo1oQBCJaa61hqkQQqhVGE8oJWKFE

# make us an EC address
factom-cli2 generateaddress ec e1

factom-cli2 newtransaction t
factom-cli2 addinput       t b 1000
factom-cli2 addecoutput    t e1 1000
factom-cli2 addfee         t b
factom-cli2 sign t
factom-cli2 submit t

for i in `seq 1 1000000`;
do
    factom-cli2 newtransaction t
    factom-cli2 addinput t b 1
    factom-cli2 addecoutput t e1 1
    factom-cli2 addfee t b
    factom-cli2 sign t
    factom-cli2 submit t
    export loopi=$i

    chain1="$(echo "Chain" $i 1 | factom-cli2 mkchain -e tchain -e $RANDOM -e $i -e 1 e1 | gawk "{print \$3}")"
    echo $chain1
    export chain1=$chain1
    chain2="$(echo "Chain" $i 2 | factom-cli2 mkchain -e tchain -e $RANDOM -e $i -e 2 e1 | gawk "{print \$3}")"
    echo $chain2
    export chain2=$chain2
    chain3="$(echo "Chain" $i 3 | factom-cli2 mkchain -e tchain -e $RANDOM -e $i -e 3 e1 | gawk "{print \$3}")"
    echo $chain3
    export chain3=$chain3

    sleep 65

    factom-cli2 balances

    sleep 5

    scripts/entryload2.sh &

done