#!/bin/bash

TMPDIR="/tmp/factomd-p2p-test-build"
mkdir $TMPDIR
cd ../../
#pwd
echo "Building linux factomd and putting it in $TMPDIR"
CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o "$TMPDIR/factomd"
