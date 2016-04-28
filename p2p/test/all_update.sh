#!/bin/bash +x
echo "Building the current Factomd as a linux binary"
./build_factomd.sh
if [ $? -eq 0 ]; then
    echo "was binary updated? Current:`date`"
    ls -G -lh "/tmp/factomd-p2p-test-build/factomd"
    echo "Updating all the servers...."

    while IFS='' read -r line || [[ -n "$line" ]]; do
      echo "Updating: $line"
      copy_files_to_test_box.sh $line
    done < "./leaders.conf"

    while IFS='' read -r line || [[ -n "$line" ]]; do
      echo "Updating: $line"
      copy_files_to_test_box.sh $line
    done < "./followers.conf"
fi
