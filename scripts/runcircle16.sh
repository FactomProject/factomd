#!/usr/bin/env bash
# Run a 26 node simulation

g factomd -enablenet=false -count=16 -fnet=scripts/networks/doublecircle16.txt -blktime=20 > out.txt
