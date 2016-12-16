#!/bin/bash

nchains=5      # number of chains to create
nentries=75  # number of entries to add to each chain

fa1=$(factom-cli importaddress Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK)

ec1=$(factom-cli importaddress Es3LB2YW9bpdWmMnNQYb31kyPzqnecsNqmg5W4K7FKp4UP6omRTa)

buyECs=$(expr $nentries \* $nchains \* 2 )
echo "Buying" $buyECs $fa1 $ec1
factom-cli buyec $fa1 $ec1 $buyECs
sleep 5s
	
addentries() {
	let y=$(shuf -i 30-120 -n 1)
	echo "sleep"  $y  " seconds before writing entries"
	sleep $y
	for ((i=0; i<nentries; i++)); do
    		cat scripts/data.txt | factom-cli addentry -f -c $1 -e test -e $i -e $RANDOM -e $RANDOM -e $RANDOM $ec1
		echo "write entry Chain:"  $2 $i
		sleep .4s
	done
}

echo "Start"

for ((i=0; i<nchains; i++)); do
	echo "create chain" $i
	chainid=$(echo test $i $RANDOM | factom-cli addchain -f -n test -n $i -n $RANDOM $ec1 | awk '/ChainID/{print $2}')
	addentries $chainid $i &
	let y=$(shuf -i 10-30 -n 1)
	echo SLEEP $y  YAWN
	sleep $y
done

let y=$(shuf -i 10-50 -n 1)
echo SLEEP $y "seconds before doing another set of chains."
sleep $y
sleep 50s
