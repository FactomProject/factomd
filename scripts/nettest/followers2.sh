. fd
echo 'tail -f out2.txt | grep ^[^2]'
g factomd -prefix="_" -count=1 -networkPort="34342" -port="8093" -logPort="6063" -db=Map -peers="127.0.0.1:34340" -network=LOCAL -blktime=60 -net=alot+  > out2.txt

