#!/usr/bin/env bash

#!/bin/bash

nchains=20   # number of chains to create
nentries=0  # number of entries to add to each chain

fa1=$(factom-cli importaddress Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK)

ec1=$(factom-cli importaddress Es3LB2YW9bpdWmMnNQYb31kyPzqnecsNqmg5W4K7FKp4UP6omRTa)

echo "Buying" 1000 $fa1 $ec1
factom-cli buyec $fa1 $ec1 100
sleep 5s

addentries() {
    # create a random datafile
	datalen=$(shuf -i 100-9900 -n 1)
	datafile=$(mktemp)
	base64 /dev/urandom | head -c $datalen > $datafile

	echo "Entry Length " $datalen " bytes, file name: " $datafile

	let y=$(shuf -i 30-120 -n 1)
	echo "sleep"  $y  " seconds before writing entries"
	sleep $y
	for ((i=0; i<nentries; i++)); do
    		cat $datafile | factom-cli addentry -f -c $1 -e test -e $i -e $RANDOM -e $RANDOM -e $RANDOM $ec1
		echo "write entry Chain:"  $2 $i
		sleep .8s
	done

    # get rid of the random datafile
	rm $datafile
}

echo "Start"

for ((i=0; i<nchains; i++)); do
	echo "create chain" $i
	chainid=$(echo test $i $RANDOM | factom-cli addchain -f -n test -n $i -n $RANDOM $ec1 | awk '/ChainID/{print $2}')
	addentries $chainid $i &
done

sleep .5

