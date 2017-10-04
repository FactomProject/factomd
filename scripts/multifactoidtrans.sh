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

big=FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q

#transfer funds to each of the entry credit address
for ((i=0; i < 50000; i++)); do
echo $i
for ((j=0; j < 10; j++)); do

factom-cli rmtx  t1 2>/dev/null

factom-cli newtx -q t1
factom-cli addtxinput  -q t1 $big .001
factom-cli addtxoutput -q t1 $fct2 .001
factom-cli addtxfee    -q t1 $big
factom-cli signtx -q t1
factom-cli sendtx -q -f t1
sleep 1

factom-cli newtx -q t1
factom-cli addtxinput  -q t1 $big .001
factom-cli addtxoutput -q t1 $fct3 .001
factom-cli addtxfee    -q t1 $big
factom-cli signtx -q t1
factom-cli sendtx -q -f t1
sleep 1

factom-cli newtx -q t1
factom-cli addtxinput  -q t1 $big .001
factom-cli addtxoutput -q t1 $fct4 .001
factom-cli addtxfee    -q t1 $big
factom-cli signtx -q t1
factom-cli sendtx -q -f t1
sleep 1

factom-cli newtx -q t1
factom-cli addtxinput  -q t1 $big .001
factom-cli addtxoutput -q t1 $fct5 .001
factom-cli addtxfee    -q t1 $big
factom-cli signtx -q t1
factom-cli sendtx -q -f t1
sleep 1

factom-cli newtx -q t1
factom-cli addtxinput  -q t1 $big .001
factom-cli addtxoutput -q t1 $fct6 .001
factom-cli addtxfee    -q t1 $big
factom-cli signtx -q t1
factom-cli sendtx -q -f t1

sleep 1

done

factom-cli listaddresses
done

