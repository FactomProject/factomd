#!/bin/bash

# do the grep
# $1=file $2=id $3=stage
function counts {
    COUNT=`grep -i ", $3, $2, Message," $1 | wc -l`
    echo "$2 $3 $COUNT"
}

# Iterate the stages
# $1=file $2=id $3=Name
function stages {
    echo "Status for $3"
    counts $1 $2 "a"
    counts $1 $2 "b"
    counts $1 $2 "c"
    counts $1 $2 "d"
    counts $1 $2 "e"
    counts $1 $2 "f"
    counts $1 $2 "G"
    counts $1 $2 "H"
    counts $1 $2 "I"
    counts $1 $2 "J"
    counts $1 $2 "K"
    counts $1 $2 "L"
    counts $1 $2 "M"
    counts $1 $2 "N"
}

# Iterate the messages
# $1 = file
function messages {
    stages $1 "0" "EOM_MSG"
    stages $1 "1" "ACK_MSG"
    stages $1 "8" "EOM_TIMEOUT_MSG"
    stages $1 "10" "HEARTBEAT_MSG"
    stages $1 "14" "REQUEST_BLOCK_MSG"
    stages $1 "16" "MISSING_MSG"
    stages $1 "17" "MISSING_DATA"
    stages $1 "18" "DATA_RESPONSE"
    stages $1 "19" "MISSING_MSG_RESPONSE"
    stages $1 "20" "DBSTATE_MSG"
    stages $1 "21" "DBSTATE_MISSING_MSG"
}


echo "Filtering and sorting..."

cat node1.out node2.out | grep '^ParcelTrace, ' | sort -f > message_traces.out

echo "analyzing..."

messages message_traces.out