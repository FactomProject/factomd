#!/usr/bin/env bash
# fund ecaddress
FADDRESS=$(factom-cli importaddress 'Fs2DNirmGDtnAZGXqca3XHkukTNMxoMGFFQxJA3bAjJnKzzsZBMH')
echo $FADDRESS
ECADDRESS=$(factom-cli newecaddress)
echo $ECADDRESS
TEXT=$(factom-cli buyec $FADDRESS $ECADDRESS 10000)
echo $TEXT
TXID=$(echo $TEXT | awk '{print $2}')
echo $TXID
STATUS=''
until [ "$STATUS" = 'TransactionACK' -o "$STATUS" = "DBlockConfirmed" ]; do
   STATUS=$(echo $(factom-cli status $TXID) | awk '{print $4}')
   echo $STATUS
done
BALANCE=$(factom-cli balance $ECADDRESS)
echo $BALANCE

for i in $(seq 100); do
   # make chain
   echo
   echo run $i
   RAN=$(date +%N)
   echo external id $RAN
   TXID=$(echo $RAN | factom-cli addchain -n $RAN -T $ECADDRESS)
   echo TXID $TXID
   STATUS=''
   until [ "$STATUS" = 'TransactionACK' -o "$STATUS" = "DBlockConfirmed" ]; do
      STATUS=$(echo $(factom-cli status $TXID) | awk '{print $4}')
      echo STATUS $STATUS
   done

   # get raw
   RAW=$(factom-cli  get raw $TXID)
   echo RAW $RAW
   if [ -z "$RAW" ] ; then break; fi
done
