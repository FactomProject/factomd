#!/bin/bash


// Allocate a boatload of Entry Credits to e2
factom-cli newtransaction t
	factom-cli addinput t b 1
	factom-cli addecoutput t e2 1
	factom-cli addfee t b
	factom-cli sign t
	factom-cli transactions
	factom-cli submit t
export chain1="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain2="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain3="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain4="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain5="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain6="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain7="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain8="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain0="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain10="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain11="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain12="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain13="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain14="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"
echo "Made Chain: " ${chain1}
export chain15="$(echo "one two three four, this is a test of making an entry.  More to test, Test, Test, test" $RANDOM $RANDOM $RANDOM | factom-cli mkchain -e $RANDOM -e $RANDOM e1 | gawk "{print \$3}")"


sleep 60

while true ;
do
	for j in `seq 1 10`;
	do
	    echo "Date: $(date)"
		factom-cli balances
		if [ $? -eq 0 ]; then
    		./scripts/factomEC2.sh
        else
            sleep 10
        fi
	done
done
