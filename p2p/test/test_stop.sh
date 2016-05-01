#!/bin/bash +x

echo "Stopping the network."

while IFS='' read -r line || [[ -n "$line" ]]; do
  echo "Stopping Leader: $line"
  ssh -n $line './stop.sh'
done < "./leaders.conf"


while IFS='' read -r line || [[ -n "$line" ]]; do
  echo "Stopping Follower: $line"
  ssh -n $line './stop.sh'
done < "./followers.conf"
