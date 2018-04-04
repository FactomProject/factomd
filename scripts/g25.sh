. fd

# General testing with specified networks.
# g factomd -count=21 -fnet="scripts/networks/dogbone21.txt" -blktime=20 -enablenet=false -network=LOCAL -startdelay=5 > out.txt


# General testing
g factomd  -count=5 -net=alot+  -blktime=60 -faulttimeout=20 -enablenet=false -network=LOCAL -startdelay=2 $@ > out.txt 2> err.txt



