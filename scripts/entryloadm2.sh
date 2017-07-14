#!/bin/bash

nchains=10   # number of chains to create
nentries=15    # number of entries to add to each chain

#factomd=10.41.0.16:8088
factomd=localhost:8088

# This address is for a LOCAL network
fa1=$(factom-cli -s=$factomd importaddress Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK)

# This address is for a network with a production Genesis block
#fa1=FA3RrKWJLQeDuzC9YzxcSwenU1qDzzwjR1uHMpp1SQbs8wH9Qbbr


ec1=$(factom-cli -s=$factomd importaddress Es3LB2YW9bpdWmMnNQYb31kyPzqnecsNqmg5W4K7FKp4UP6omRTa)

factom-cli -s=$factomd listaddresses

buyECs=$(expr $nentries \* $nchains \* 11 )
echo "Buying" $buyECs $fa1 $ec1
factom-cli -s=$factomd buyec $fa1 $ec1 $buyECs
sleep 5s
	
factom-cli -s=$factomd listaddresses

addentries() {
    # create a random datafile
	datalen=$(shuf -i 100-1900 -n 1)
	datafile=$(mktemp)
	base64 /dev/urandom | head -c $datalen > $datafile

	echo "Entry Length " $datalen " bytes, file name: " $datafile

	for ((i=0; i<nentries; i++)); do
    		cat $datafile | factom-cli -s=$factomd addentry  -c $1 -e test -e $i -e $RANDOM -e $RANDOM -e $RANDOM $ec1
		echo "write entry Chain:"  $2 $i
		sleep 5.2s
	done
  
  # get rid of the random datafile
	rm $datafile
}

echo "Start"

for ((i=0; i<nchains; i++)); do
	echo "create chain" $i
	chainid=$(echo test $i $RANDOM | factom-cli -s=$factomd addchain  -n test -n $i -n $RANDOM $ec1 | awk '/ChainID/{print $2}')
	addentries $chainid $i &
	sleep 10
done


echo SLEEP "90 seconds before doing another set of chains."
sleep 20
echo About ready ...
echo 10
sleep 1
echo  9
sleep 1
echo  8
sleep 1
echo  7
sleep 1
echo  6
sleep 1
echo  5
sleep 1
echo  4
sleep 1
echo  3
sleep 1
echo  2
sleep 1
echo  1
sleep 1

