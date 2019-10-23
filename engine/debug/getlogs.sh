#!/usr/bin/env bash
#get logs from the AWS test nodes
scp tdd-d *.txt .
rename 's/fnode0_/fnode03_/' *.txt
