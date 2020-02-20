. fd
echo 'tail -f out0.txt | grep ^[^2]'
g factomd -prefix="y" -count=1 -db=Map -peers="127.0.0.1:34340" -network=LOCAL -blktime=30 -net=tree -startdelay=1 > out0.txt


