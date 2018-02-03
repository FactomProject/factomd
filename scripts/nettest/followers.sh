. fd
echo 'tail -f out1.txt | grep ^[^2]'
g factomd -prefix="x" -count=1 -networkPort="34341" -port="8092" -ControlPanelPort="9081" -logPort="6062" -db=Map -peers="127.0.0.1:34340" -network=LOCAL -blktime=30 -net=alot+ > out1.txt


