#!/bin/bash

nchains=4      # number of chains to create
nentries=2000  # number of entries to add to each chain

fa1=$(factom-cli importaddress Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK)

ec1=$(factom-cli importaddress Es3LB2YW9bpdWmMnNQYb31kyPzqnecsNqmg5W4K7FKp4UP6omRTa)

buyECs=$(expr $nentries \* $nchains \* 2 )
echo "Buying" $buyECs $fa1 $ec1
factom-cli buyec $fa1 $ec1 $buyECs
sleep 1s
	
addentries() {
	sleep 70s
	for ((i=0; i<nentries; i++)); do
    	echo test $i $RANDOM | factom-cli addentry -c $1 -e test -e $i -e $RANDOM $ec1
		echo "write entry" $i
		sleep .01s
	done
}

echo "Start"

for ((i=0; i<nchains; i++)); do
	echo "create chain" $i
	chainid=$(echo test $i $RANDOM | factom-cli addchain -e test -e $i -e $RANDOM $ec1 | awk '/ChainID/{print $2}')
	addentries $chainid &
    sleep 10s
done

sleep 100s
