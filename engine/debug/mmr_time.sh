#/bin/sh
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
func time2sec(t) {	
  x = split(t,ary,":");
  if(x!=3) {print "bad time",NR, $0; exit;}
  sec = (ary[1]*60+ary[2])*60+ary[3];
#  printf("time2sec(%s) %02d:%02d:%02d= %d\\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
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
    cnt = match($0,/([0-9]+\\/[0-9]\\/[0-9]+, )+/,ary);
    list = substr($0,RSTART,RLENGTH);
    total_request_msgs_a++
 #   print "          1         2         3         4         5         6         7         8         9         0         1         2         3         4         5         6         7         8         9         0"
 #   print "012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
 #   print $0
    ts =  time2sec($2);
    n = split(list,ary,", ");
    for( i in ary ) {
	if(ary[i] ~ /[0-9]+\\/[0-9]\\/[0-9]+/) {
	   total_requests_a++;
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
#    print "MM", $2, $0
    ts =  time2sec($2);
    total_request_msgs_b++
    cnt = match($0,/([0-9]+\\/[0-9]\\/[0-9]+, )+/,ary);
    list = substr($0,RSTART,RLENGTH);
     n = split(list,ary,", ");
    for( i in ary ) {
        v = ary[i]
#	print i, v
	if(v ~ /[0-9]+\\/[0-9]\\/[0-9]+/) {
	   total_requests_b++;
	    asking[v][substr($6,6)]++
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
    sub(/:/," ");
#   print "MMR           1         2         3         4         5         6         7         8         9         0         1         2         3         4         5         6         7         8         9         0"
#   print "MMR 012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
#    print "MMR " $0
    ts =  time2sec($3);
    total_responces++
    cnt = match($0,/([0-9]+\\/[0-9]\\/[0-9]+)+/,ary);
    list = substr($0,RSTART,RLENGTH);
    n = split(list,ary,", ");
    v = ary[1];
#	print list, n, v
    peer = substr($1,1,index($1,"_")-1)
#    print "MMR", v, peer,$1;
	if(v ~ /[0-9]+\\/[0-9]\\/[0-9]+/) {
           if(!(v in firstMR)) {
               firstMR[v] = ts;
               lastMR[v] = ts;
#               print "firstMR["v"] = " firstMR[v], asks[v];
              } else {
               lastMR[v] = ts;
           }
           peers[v][peer]++;
	   countMR[v]++;
        }
    
}


END {
   printf("%10s %10s %10s %10s %10s %10s %10s %10s %10s %10s %10s %-10s\\n", "loc","ask2LAsk","ask2send", "ask2last", "ask2add", "askcount", "ask2p2p", "ask2Lp2p","firstMR","lastMR", "replies","peers");
   PROCINFO["sorted_in"] ="@ind_num_asc";
   for(i in firstsendout) {
     ask = asks[i]
     if(i in lasks) {lask = lasks[i]-ask;}else{lask="NA"}
     fs = firstsendout[i]-ask
     ls = lastsendout[i]-ask
     if(i in adds) {add = adds[i]-ask} else {add = "never"}
     if(i in firstp2p){ fp = firstp2p[i]-ask;} else {fp = "NA";}
     if(i in lastp2p) { lp = lastp2p[i]-ask;}  else {lp = "NA";}
     if(i in firstMR) { fr = firstMR[i]-ask;} else {fr = "NA";}
     if(i in lastMR)  { lr = lastMR[i]-ask;}  else {lr = "NA";}
     delete peerCnt
     replies = countMR[i]+0
     PROCINFO["sorted_in"] ="@ind_str_asc";
     peerStr = ""
     if(i in peers) {
       
        for(j in peers[i]) {
#          print "<"i"><"j">["peers[i][j]"]";
          peerStr = peerStr " "  j "-" peers[i][j];
        }
        peerStr = substr(peerStr,2)
     } else {peerStr = "NA";}

     if (fs > 5 || ls > 5 || add > 5|| fp > 5 )	 {
        printf("%10s %10s %10s %10s %10s %10s %10s %10s %10s %10s %10s %-10s\\n", i, lask, fs, ls, add, sendcnt[i], fp, lp, fr,lr, replies,  peerStr);
     }
   }

	printf("Total requests messages %d/%d, total requests %d/%d, total responces %d\\n", total_request_msgs_a,total_request_msgs_b,total_requests_a,total_requests_b,  total_responces);
}
EOF
################################
# End of AWK Scripts           #
################################

 
(cat $1_missing_messages.txt; grep -h "MissingMsg " $1_NetworkOutputs.txt; grep -HE "Send P2P.* $1 .*Missing Msg Response" FNode*_NetworkOutputs.txt) | awk "$scriptVariable"

