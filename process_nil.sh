#!/usr/bin/env bash
#/usr/bin/bash
cd ~/go/src/github.com/FactomProject/factomd
#/usr/bin/gnome-terminal -- factomd -prefix="x" -count=1 -networkPort=34341 -port=8090 -ControlPanelPort=9081 -logPort=6062 -db=Map -peers=127.0.0.1:34340 -network=LOCAL -blktime=30 -net=alot+ -stdoutlog out0.txt -debugconsole localhost:8093
/usr/bin/gnome-terminal -- factomd -count=16 -port="8091" -networkPort="34340" -logPort="6061" -peers="127.0.0.1:34341" -network=LOCAL -blktime=30 -net=alot+ -startdelay=30 -stdoutlog  out0.txt -debugconsole remotehost:8094
/usr/bin/gnome-terminal --geometry=200x60 -- scripts/startstatus.sh out0.txt
/usr/bin/gnome-terminal --geometry=200x40 -- scripts/startstatus.sh out1.txt
/usr/bin/gnome-terminal tail -f  out1.txt | tee pl1.txt | grep --line-buffered -E -A1 "<nil>"  | awk ' {print; if($0~"<nil>") x++; printf("%d\r",x)}'

