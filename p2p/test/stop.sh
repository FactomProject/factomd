#!/bin/bash +x

pkill -2 factomd

file="runlog.txt"

echo >> $file
echo "Stopped by stop.sh (Eg: factomd didn't crash.)" >> $file
echo >> $file
echo >> $file

