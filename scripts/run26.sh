#!/usr/bin/env bash
# Run a 26 node simulation

g factomd -enablenet=false -count=26 -fnet=scripts/networks/fourSegments26.txt -blktime=60 > out.txt
