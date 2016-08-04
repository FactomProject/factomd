#!/usr/bin/env bash
#import addresses

factom-cli rmtx t

b=$(factom-cli  importaddress Fs1fbEjDxWhXasmnp6R7DMFRo1oQBCJaa61hqkQQqhVGE8oJWKFE)

echo "Start loop" $b

for ((i=0; i < 1000000; i++)); do
    d=$(factom-cli newfctaddress)
    factom-cli newtx t
    factom-cli addtxinput t $b 0.000001
    factom-cli addtxoutput t $d 0.000001

    factom-cli addtxfee t $b
    factom-cli signtx t
    tx=$(factom-cli sendtx t)

    sleep .01
    (( v = i % 100 ))
    if [ $v -eq "0" ]
    then
        factom-cli listaddresses
    fi
done

