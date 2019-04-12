#/bin/sh
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'

func time2sec(t) {	
  x = split(t,ary,":");
  if(x!=3) {printf("time2sec(%s) bad split got %d fields. %d", t , x ,NR); print $0; exit;}
  sec = (ary[1]*60+ary[2])*60+ary[3];
#  printf("time2sec(%s) %02d:%02d:%02d= %d\\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
  return sec;
}

 {
   if(NR%1000==0) {printf("\\r%7d",NR)>"/dev/stderr";}
}

#    432 15:35:13.018                             19140-:-0 Send P2P P2P Network           M-9ca69f|R-9ca69f|H-9ca69f|0xc42150a180            DBState Missing[21]:DBStateMissing: 19140-19345 RandomPeer
#    439 15:35:17.574                             19140-:-0 peer-0, enqueue                M-7d1345|R-3b2fe0|H-7d1345|0xc427c16d80                    DBState[20]:DBState: dbht:19140 [size:     618,556] dblock 88d5b0 admin 637d19 fb 5759bf ec f174eb hash 7d1345 FNode00
#    443 15:35:19.874                             19141-:-0 peer-0, enqueue                M-119c63|R-47de82|H-119c63|0xc42a828d80                    DBState[20]:DBState: dbht:19141 [size:     615,788] dblock d4bb61 admin b5e72f fb 962cc9 ec c872f0 hash 119c63 FNode00
#    521 15:35:26.030                             19142-:-0 Send P2P P2P Network           M-ece8de|R-ece8de|H-ece8de|0xc4227580c0            DBState Missing[21]:DBStateMissing: 19142-19347 RandomPeer
 
/Send/ {
   if(prev==0) {prev=$2; first=$2;}  
   loss = time2sec($2)-time2sec(prev);
}

/enqueue/ {prev = $2;}

END {
    printf("\\r%7d\\n",NR)>"/dev/stderr"
    total = time2sec(prev)-time2sec(first)
    print "Time lost = ", loss, "out of", total, loss/total;

}

 
EOF
################################
# End of AWK Scripts           #
################################

 


(grep "enqueue.*DBState\[20\]:" fnode0_networkinputs.txt;  grep "Send.* DBState Missing" fnode0_networkoutputs.txt) | sort -n | gawk  "$scriptVariable" 
