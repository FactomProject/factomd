##import addresses 
#factom-cli  importaddress Fs2DNirmGDtnAZGXqca3XHkukTNMxoMGFFQxJA3bAjJnKzzsZBMH
#factom-cli  importaddress Es3LB2YW9bpdWmMnNQYb31kyPzqnecsNqmg5W4K7FKp4UP6omRTa

 
#assign public keys to variables
#ec=EC1nje9iEd4k3hzHad4Qty7fAxKdji9Ep4ZjnRiTAcSrEDL1drU4


#transfer funds to each of the entry credit address
#factom-cli newtx t1
#factom-cli addtxinput t1 FA3EPZYqodgyEGXNMbiZKE5TS2x2J9wF8J9MvPZb52iGR78xMgCb 1
#factom-cli addtxecoutput t1 ${ec} 1
#factom-cli addtxfee t1 FA3EPZYqodgyEGXNMbiZKE5TS2x2J9wF8J9MvPZb52iGR78xMgCb
#factom-cli signtx t1
#factom-cli sendtx t1
#factom-cli listaddresses
#sleep 1s

#create chain with each of the entry credit address
#echo ${ec}
#echo "hello world" | factom-cli addchain -h 01020304050607 -C EC1nje9iEd4k3hzHad4Qty7fAxKdji9Ep4ZjnRiTAcSrEDL1drU4

#sleep 1s
#echo ${ChainId}

#exit
#factom-cli addentry to each of the chains
date
for j in $(seq 10); do
echo "Making entries into chain $j" 
echo "---------------------------------"
echo `date` $j | factom-cli addentry -c a669a4141f56b33d635a76c538dabafb72d136de59cc6373260376376dcba307 EC1nje9iEd4k3hzHad4Qty7fAxKdji9Ep4ZjnRiTAcSrEDL1drU4 -s localhost:8088
echo "---------------------------------"
done
date


