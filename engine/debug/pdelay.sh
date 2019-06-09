#!/usr/bin/env bash

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

 {sub(/:/," ");}

#FNode0_processList.txt     194 09:59:22.472 5-:-0 Added at 5/1/0      M-1656cf|R-446aad|H-1656cf|0xc4243b2000  Directory Block Signature[ 7]: DBSig-VM  1:          DBHt:    5 -- Signer=b1455b7b PrevDBKeyMR[:3]=bbc033 hash=1656cf 


/ Added at / {
	msg[$7] = substr($0,index($0,"M-"));
	added[$7] = $3;
#     print;
#     printf("Added <%s> <%s> <%s>\\n", $7, added[$7], msg[$7]);
} 

# 257433 09:59:55.815 5-:-1 retry 5/1/252       M-567dad|R-9527a0|H-567dad|0xc420241040                        EOM[ 0]:   EOM-     DBh/VMh/h 5/1/-- minute 1 FF  0 --Leader[1570f8<FNode01>] hash[567dad]  
/ retry / {
	retries[$6]++;
#     print;
#     printf("retry <%s> <%s>\\n", $6,retries[$6]);
}
/ done / {
	done[$6] = $3;
#    printf("done <%s> <%s>\\n", $6,done[$6]);
}

END {
   print "computing wait times"
   for(i in done) {
#      printf("%s(%s)(%s)\\n",i,done[i],added[i]);
      waittime[i] = time2sec(done[i]) - time2sec(added[i])
   }

   PROCINFO["sorted_in"] ="@val_num_desc";
   cnt=0
   print "Wait times"
   for(i in waittime) {
     printf("waited %7.3f for %s\\n", waittime[i], msg[i]);
     if(cnt++ > 10) {break;}
   }

   cnt=0;
   print "Retries"
   for(i in retries) {
     printf("retries %5d for %s\\n", retries[i], msg[i]);
     if(cnt++ > 10) {break;}
   }

}
EOF
################################
# End of AWK Scripts           #
################################

 
grep -H -E "Added|done|retry"  $1_process*.txt  | awk "$scriptVariable"

