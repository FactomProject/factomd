#!/bin/bash

nchains=20      # number of chains to create
nentries=1500  # number of entries to add to each chain

fa1=$(factom-cli importaddress Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK)

ec1=$(factom-cli importaddress Es3LB2YW9bpdWmMnNQYb31kyPzqnecsNqmg5W4K7FKp4UP6omRTa)

buyECs=$(expr $nentries \* $nchains \* 2 )
echo "Buying" $buyECs $fa1 $ec1
factom-cli buyec $fa1 $ec1 $buyECs
sleep 1s
	
addentries() {
	sleep 60s
	echco ">>>>>>>>>>>>>>>>>>>>>>>>>> "   Writing Entries soon.
	Sleep 10s
	for ((i=0; i<nentries; i++)); do
    	echo test $i $RANDOM | factom-cli addentry -c $1 -e test -e $i -e $RANDOM $ec1
		echo "write entry Chain:"  $2 $i
		sleep .1s
	done
}

echo "Start"

for ((i=0; i<nchains; i++)); do
	echo "create chain" $i
	chainid=$(echo test $i $RANDOM | factom-cli addchain -e test -e $i -e $RANDOM $ec1 | awk '/ChainID/{print $2}')
	addentries $chainid $i &
    sleep 15s
done

sleep 540s
