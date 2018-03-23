#!/bin/bash

nchains=21    # number of chains to create
nchains2=2    # number of chains to create
nentries=10   # number of entries to add to each chain

#factomd=10.41.2.5:8088
 factomd=localhost:8088

# This address is for a LOCAL network
fa1=$(factom-cli -s=$factomd importaddress Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK)

# This address is for a network with a production Genesis block
#fa1=FA3RrKWJLQeDuzC9YzxcSwenU1qDzzwjR1uHMpp1SQbs8wH9Qbbr

minsleep=1
randsleep=2
entrysize=2048

ec1=$(factom-cli -s=$factomd importaddress Es3LB2YW9bpdWmMnNQYb31kyPzqnecsNqmg5W4K7FKp4UP6omRTa)


buyECs=10000000
echo "Buying" $buyECs $fa1 $ec1
factom-cli -s=$factomd buyec $fa1 $ec1 $buyECs
	
factom-cli -s=$factomd listaddresses

addentries() {
    # create a random datafile
	datalen=$(shuf -i 10-$entrysize -n 1)
	datafile=$(mktemp)
	base64 /dev/urandom | head -c $datalen > $datafile

	sleep $(( ( RANDOM % $randsleep )/4  + minsleep ))

	echo "Entry Length " $datalen " bytes, file name: " $datafile

	for ((i=0; i<nentries; i++)); do
    		cat $datafile | factom-cli -s=$factomd addentry -f -c $1 -e test -e $i -e $RANDOM -e $RANDOM -e $RANDOM $ec1
		echo "write entry Chain:"  $2 $i
		sleep $((  1 ))
	done
	for ((i=0; i<nentries; i++)); do
    		cat $datafile | factom-cli -s=$factomd addentry -f -c $1 -e test -e $i -e $RANDOM -e $RANDOM -e $RANDOM $ec1
		echo "write entry Chain:"  $2 $i
		sleep $(( RANDOM % 20 ))
	done
  
  # get rid of the random datafile
	rm $datafile
}

echo "Start"

for ((ii=0; ii<nchains2; ii++)); do
	for ((i=0; i<nchains; i++)); do
		echo "create chain" $i
		chainid=$(echo test $i $RANDOM | factom-cli -s=$factomd addchain -f  -n test -n $i -n $RANDOM $ec1 | awk '/ChainID/{print $2}')
		addentries $chainid $i &
	done
	sleep $minsleep
done

echo SLEEP "a little pause before we go again!"
sleep $(( $minsleep * 10 ))

echo About ready ...
sleep $(( $minsleep * 2 ))

