#!/usr/bin/env bash
factom-cli importaddress Fs2DNirmGDtnAZGXqca3XHkukTNMxoMGFFQxJA3bAjJnKzzsZBMH
factom-cli importaddress Fs1XY1RcH3b9ezVqCA3WDxAQhnVFnWs5AgcKUU9YqJeradPsimcA
factom-cli importaddress Fs2ddd3Wh37sWYNFuGgNjdCPxBMpLYZjLKGXE859SyTfiGGPX9JU
factom-cli importaddress Fs26rt5wF8AE8YnfsRbvDG9YXdX2XnG77x99fMFLeTsxehnPNZTK
factom-cli importaddress Fs2nDerk8yp2GdAjsrukKWHatuCqH36rupxKaYnKZFXigKbW9VEm
factom-cli importaddress Fs26rt5wF8AE8YnfsRbvDG9YXdX2XnG77x99fMFLeTsxehnPNZTK
factom-cli importaddress Fs1aXUnnPGu1Q2ddP2kz5CwD4fCVpnrDJDzMtuDDwcgfrMN1yoeS

fct1=FA3EPZYqodgyEGXNMbiZKE5TS2x2J9wF8J9MvPZb52iGR78xMgCb
fct2=FA3hhUyUoqn8W19DsNtvrCiyvbVSD5rv6ncNoVPek7FmwFX8nUrY
fct3=FA35Bm8D59XSNXZW5JpDSdZBHnrpVS25Tzy7zk99FM4YZu565Uq5
fct4=FA2D3KjuK8GcGBt3JP3tgSA1RjyYd6C57JmpeAyVwM4vBQy17JiM
fct5=FA35dnDS4ic1mZbud8fN9zmKpNv5L7FC16DRdXnCvkTYTg3T7i2W
fct6=FA3pVAwckTzAhCA4U5fg2ygdkQYr5UggaapkNwLhAksMBFiZrZmx



#transfer funds to each of the entry credit address
for ((i=0; i < 50000; i++)); do
echo $i
factom-cli newtx t1
factom-cli addtxinput  t1 $fct1 .05
factom-cli addtxoutput t1 $fct2 .01
factom-cli addtxoutput t1 $fct3 .01
factom-cli addtxoutput t1 $fct4 .01
factom-cli addtxoutput t1 $fct5 .01
factom-cli addtxoutput t1 $fct6 .01
factom-cli addtxfee t1 $fct1
factom-cli signtx t1
factom-cli sendtx t1
sleep 1s
factom-cli listaddresses
done

