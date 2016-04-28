#!/bin/bash +x
echo
echo "Stopping factomd on $1 before updating."
ssh -n $1 './stop.sh'
echo "Copies the local authorized_keys to ~/.ssh/authorized_keys"
scp ../../../FactomInc/ops/authorized_keys $1:~/.ssh/authorized_keys
echo "Removing previous log file."
ssh -n $1 'echo \"Config Remote Test Box Reset\" > runlog.txt'
echo "Making the .factom directory, and m2 directory and copying factomd.conf"
ssh -n $1 'rm -rf .factom; mkdir .factom; cd .factom; mkdir m2'
scp factomd.conf $1:~/.factom/m2/
scp factomd.conf $1:~/.factom/
