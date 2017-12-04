#!/usr/bin/env bash

fct1=$(factom-cli -s=$factomd importaddress Fs2DNirmGDtnAZGXqca3XHkukTNMxoMGFFQxJA3bAjJnKzzsZBMH)
fct2=$(factom-cli -s=$factomd importaddress Fs1XY1RcH3b9ezVqCA3WDxAQhnVFnWs5AgcKUU9YqJeradPsimcA)
fct3=$(factom-cli -s=$factomd importaddress Fs2ddd3Wh37sWYNFuGgNjdCPxBMpLYZjLKGXE859SyTfiGGPX9JU)
fct4=$(factom-cli -s=$factomd importaddress Fs26rt5wF8AE8YnfsRbvDG9YXdX2XnG77x99fMFLeTsxehnPNZTK)
fct5=$(factom-cli -s=$factomd importaddress Fs2nDerk8yp2GdAjsrukKWHatuCqH36rupxKaYnKZFXigKbW9VEm)
fct6=$(factom-cli -s=$factomd importaddress Fs1aXUnnPGu1Q2ddP2kz5CwD4fCVpnrDJDzMtuDDwcgfrMN1yoeS)

nfct=$(factom-cli newfctaddress)

echo 1 $fct1
echo 2 $fct2
echo 3 $fct3
echo 4 $fct4
echo 5 $fct5
echo 6 $fct6

tx=$(cat /dev/urandom | tr -cd 'a-f' | head -c 10)
paws=.8

big=FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q

#transfer funds to each of the entry credit address
for ((i=0; i < 5000000; i+=50)); do
echo Submitted $i transactions
for ((j=0; j < 10; j++)); do

factom-cli rmtx  $tx 2>/dev/null

factom-cli newtx -q $tx
factom-cli addtxinput  -q $tx $big .001
factom-cli addtxoutput -q $tx $fct2 .001
factom-cli addtxfee    -q $tx $big
factom-cli signtx -q $tx
factom-cli sendtx -q -f $tx
sleep $paws

factom-cli newtx -q $tx
factom-cli addtxinput  -q $tx $big .001
factom-cli addtxoutput -q $tx $fct3 .001
factom-cli addtxfee    -q $tx $big
factom-cli signtx -q $tx
factom-cli sendtx -q -f $tx
sleep $paws

factom-cli newtx -q $tx
factom-cli addtxinput  -q $tx $big .001
factom-cli addtxoutput -q $tx $fct4 .001
factom-cli addtxfee    -q $tx $big
factom-cli signtx -q $tx
factom-cli sendtx -q -f $tx
sleep $paws

factom-cli newtx -q $tx
factom-cli addtxinput  -q $tx $big .001
factom-cli addtxoutput -q $tx $fct5 .001
factom-cli addtxfee    -q $tx $big
factom-cli signtx -q $tx
factom-cli sendtx -q -f $tx
sleep $paws

factom-cli newtx -q $tx
factom-cli addtxinput  -q $tx $big .001
factom-cli addtxoutput -q $tx $fct6 .001
factom-cli addtxfee    -q $tx $big
factom-cli signtx -q $tx
factom-cli sendtx -q -f $tx

sleep $paws

done

done

