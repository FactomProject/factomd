#!/bin/bash +x

echo "Configuring all of the machines"

while IFS='' read -r line || [[ -n "$line" ]]; do
  echo "Configuring: $line"
  config_remote_test_box.sh $line
done < "./leaders.conf"

while IFS='' read -r line || [[ -n "$line" ]]; do
  echo "Configuring: $line"
  config_remote_test_box.sh $line
done < "./followers.conf"
