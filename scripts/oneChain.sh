#!/bin/bash

nchains=1            # number of chains to create
nentries=15000000    # number of entries to add to each chain

#factomd=10.41.0.16:8088
factomd=localhost:8088

# This address is for a LOCAL network
fa1=$(factom-cli -s=$factomd importaddress Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK)

# This address is for a network with a production Genesis block
#fa1=FA3RrKWJLQeDuzC9YzxcSwenU1qDzzwjR1uHMpp1SQbs8wH9Qbbr

maxsleep=15

ec1=$(factom-cli -s=$factomd importaddress Es3LB2YW9bpdWmMnNQYb31kyPzqnecsNqmg5W4K7FKp4UP6omRTa)


buyECs=$(expr $nentries \* $nchains \* 11 )
echo "Buying" $buyECs $fa1 $ec1
factom-cli -s=$factomd buyec $fa1 $ec1 $buyECs
sleep 5s
	

addentries() {
    # create a random datafile
	datalen=$(shuf -i 100-1900 -n 1)
	datafile=$(mktemp)
	base64 /dev/urandom | head -c $datalen > $datafile

	sleep $(( ( RANDOM % $maxsleep )  + 1 ))

	echo "Entry Length " $datalen " bytes, file name: " $datafile

	for ((i=0; i<nentries; i++)); do
    		cat $datafile | factom-cli -s=$factomd addentry -f -c $1 -e test -e $i -e $RANDOM -e $RANDOM -e $RANDOM $ec1
		echo "write entry Chain:"  $1 $2 $i
		sleep .5
	done
  
  # get rid of the random datafile
	rm $datafile
}

echo "Start"

for ((i=0; i<nchains; i++)); do
	echo "create chain" $i
	#chainid=$(echo test $i $RANDOM | factom-cli -s=$factomd addchain -f  -n test -n $i -n $RANDOM $ec1 | awk '/ChainID/{print $2}')
	addentries 658c1e459b9aefca46fa10da43c3938e6b31698dcbdd2f4ad0cf35f7bf9aa6fa $i &
	sleep $(( ( RANDOM % $maxsleep )  + 1 ))
done


echo SLEEP "90 seconds before doing another set of chains."
sleep $(( ( RANDOM % ($maxsleep*2) )  + 1 ))
echo About ready ...
sleep $maxsleep
sleep 10h

