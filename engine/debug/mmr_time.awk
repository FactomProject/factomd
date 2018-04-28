func time2sec(t) {
	
  x = split(t,ary,":");
  if(x!=3) {print "bad time",NR, $0; exit;}
  sec = (ary[1]*60+ary[2])*60+ary[3];
#  printf("time2sec(%s) %02d:%02d:%02d= %d\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
  return sec;
}

#   8285 16:41:39 3-:-0 Enqueue             M-f8a39e|R-f8a39e|H-f8a39e                Missing Msg[16]:MissingMsg --> 455b7b<FNode0> asking for DBh/VMh/h[3/1/1, ] Sys: 0 msgHash[f8a39e]

 
/Ask/         { 
      if (!($5 in asks)) {
          asks[$5] = time2sec($2); 
          lasks[$5] = asks[$5];
      } else {
          lasks[$5] = time2sec($2); 
      }
     #print "Ask", "<"$5">", asks[$5];
}
/Add/         {
     adds[$5] = time2sec($2)
}
/sendout/     {
    cnt = match($0,/([0-9]+\/[0-9]\/[0-9]+, )+/,ary);
    list = substr($0,RSTART,RLENGTH);
 #   print "          1         2         3         4         5         6         7         8         9         0         1         2         3         4         5         6         7         8         9         0"
 #   print "012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
 #   print $0
    ts =  time2sec($2);
    n = split(list,ary,", ");
    for( i in ary ) {
	if(ary[i] ~ /[0-9]+\/[0-9]\/[0-9]+/) {
           v = ary[i];
	   a = asks[v]; 
#	   print v, "ask to sendout", ts-a; 
           if(!(v in firstsendout)) {
               firstsendout[v] = ts;
               lastsendout[v] = ts;
           } else {
               lastsendout[v] = ts;
           }
	    sendcnt[v]++
        }
    }
}

#   8286 16:41:39 3-:-0 Send P2P FNode02    M-f8a39e|R-f8a39e|H-f8a39e                Missing Msg[16]:MissingMsg --> 455b7b<FNode0> asking for DBh/VMh/h[3/1/1, ] Sys: 0 msgHash[f8a39e]
/Send P2P.*MissingMsg / {
    print "MM", $2, $0
    ts =  time2sec($2);

    cnt = match($0,/([0-9]+\/[0-9]\/[0-9]+, )+/,ary);
    list = substr($0,RSTART,RLENGTH);
     n = split(list,ary,", ");
    for( i in ary ) {
        v = ary[i]
	print i, v
	if(v ~ /[0-9]+\/[0-9]\/[0-9]+/) {
	    asking[v] = asking[v] "," substr($6,6)
            if(!(v in firstp2p)) {
               firstp2p[v] = ts;
               lastp2p[v] = ts;
#               print "firstp2p["v"] = " firstp2p[v], asks[v];
              } else {
               lastp2p[v] = ts;
           }
        }
    }

}

#   4159 16:41:26 1-:-4 Send P2P FNode02    M-b86815|R-b86815|H-b86815       Missing Msg Response[19]:MissingMsgResponse <-- DBh/VMh/h[         1/0/45] msgHash[b86815] EmbeddedMsg: REntry-VM  0: Min:   4          -- Leader[455b7b<FNode0>] Entry[c1c4d4] ChainID[888888dc44] hash[c1c4d4] |    ACK-    DBh/VMh/h 1/0/45        -- Leader[455b7b<FNode0>] hash[c1c4d4]
/MissingMsgResponse/{
#    print "MMR", $0
    ts =  time2sec($2);
#  print "          1         2         3         4         5         6         7         8         9         0         1         2         3         4         5         6         7         8         9         0"
#   print "012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
#   print $0

    cnt = match($0,/([0-9]+\/[0-9]\/[0-9]+)+/,ary);
    list = substr($0,RSTART,RLENGTH);
#   peer = substr($6,6)
    n = split(list,ary,", ");
#	print n,ary[1]
    i=1;
        v = ary[1];
	print "MMR", v, peer;
	if(v ~ /[0-9]+\/[0-9]\/[0-9]+/) {
           if(!(v in firstMR)) {
               firstMR[v] = ts;
               lastMR[v] = ts;
#               print "firstMR["v"] = " firstMR[v], asks[v];
              } else {
               lastMR[v] = ts;
           }
		countMR[v]++
        }
    
}


END {
   printf("%10s %10s %10s %10s %10s %10s %10s %10s %10s %10s %10s\n", "loc","ask2LAsk","ask2send", "ask2last", "ask2add", "askcount", "ask2p2p", "ask2Lp2p","peers","firstMR","lastMR");
   for(i in firstsendout) {
     ask = asks[i]
     if(i in lasks) {lask = lasks[i]-ask;}else{lask="NA"}
     fs = firstsendout[i]-ask
     ls = lastsendout[i]-ask
     add = asks[i]-ask   
     if(i in firstp2p){ fp = firstp2p[i]-ask;} else {fp = "NA";}
     if(i in lastp2p) { lp = lastp2p[i]-ask;}  else {lp = "NA";}
     if(i in asking)  { peers = substr(asking[i],2);} else { peers="";}
     if(i in firstMR) { fr = firstMR[i]-ask;} else {fr = "NA";}
     if(i in lastMR)  { lr = lastMR[i]-ask;}  else {lr = "NA";}
     if (fs > 1 || ls > 1 || add > 1|| fp != "NA" || 1)	 {
        printf("%10s %10s %10s %10s %10s %10s %10s %10s %10s %10s %10s\n", i, lask, fs, ls, add, sendcnt[i], fp, lp, peers, fr,lr);
     }
   }
}

