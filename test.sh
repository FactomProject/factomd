#!/usr/bin/env bash

# KLUDGE: just test brainswap
cd support/dev/simulator/brainSwap/
nohup ./test0.sh &
./test1.sh
