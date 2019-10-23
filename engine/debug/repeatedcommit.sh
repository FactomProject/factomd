#!/usr/bin/env bash

#set -x
# repeatedly attempt repeating identical commits

#killall factom-walletd
#factom-walletd &

ECADDRESS="EC2DKSYyRcNWf7RS963VFYgMExoHRYLHVeCfQ9PGPmNzwrcmgm2r"
factom-cli buyec  FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q $ECADDRESS 200000
sleep 1

for i in $(seq 10000); do
   # make chain
   echo  111
   echo run $i

   # create random external id
   RAN=$(date +%N)
   echo external id $RAN

   # compose chain
   compose_result=$(echo $RAN | factom-cli -w localhost:8089 -s localhost:8088 composechain -n $RAN $ECADDRESS)
   echo compose_result $compose_result
   echo 222

   # extract and run commit command
   commit_command=$(echo 'curl'$(echo $compose_result | awk -F'curl' '{ print $2 }') | bash)
   echo commit_command $commit_command
   echo
   list=$(echo $commit_command | sed -e 's/[{}]/''/g' | awk -v RS=',' -F: '{print $1 $2}')
   echo list $list
   echo 333

   # grab returned txid
   TXID=$(echo $(echo $list | awk -F'"txid""' '{print $2}') | awk -F'"' '{print $1}')
   echo TXID $TXID
   echo 444

   # wait for transack
   STATUS=''
   until [ "$STATUS" = 'TransactionACK' -o "$STATUS" = "DBlockConfirmed" ]; do
      STATUS=$(echo $(factom-cli -w localhost:8089 -s localhost:8088 status $TXID) | awk '{print $4}')
      echo STATUS $STATUS
   done
   echo 555

   # wait for set time after transack
   post_ack_wait=0
   echo sleeping $post_ack_wait
   sleep $post_ack_wait
   echo 666

   # compose same chain again with different timestamp
   result=$(echo $RAN | factom-cli -w localhost:8089 -s localhost:8088 composechain -n $RAN $ECADDRESS)
   echo result $result
   echo 777 

   # extract and run commit command
   commit_command=$(echo 'curl'$(echo $result | awk -F'curl' '{ print $2 }') | bash)
   echo commit_command $commit_command
   echo 888 
   
   if [[ $commit_command != *"error"* ]]; then echo "break"; break; fi
   echo 999
   echo
done
