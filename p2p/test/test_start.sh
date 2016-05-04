#!/bin/bash +x

echo "Starting the network.  Starting the leaders..."

while IFS='' read -r line || [[ -n "$line" ]]; do
  echo "Starting Leader: $line"
  ssh -n -T $line './leader.sh'
done < "./leaders.conf"

while IFS='' read -r line || [[ -n "$line" ]]; do
  echo "Starting Follower: $line"
  ssh -n -T $line './follower.sh'
done < "./followers.conf"
