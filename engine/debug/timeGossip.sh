#/bin/sh
pattern="$1"
shift
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

 
  #{print;}
  
/Send broadcast/ {
   
   hash = substr($0,index($0,"R-"),8);
   s = $3
   if(!(hash in sends) || sends[hash] > s){
    sends[hash] = s;
    sender[hash] = substr($1,1,7);
    #print "send", sender[hash], sends[hash];
   }
}
   
/enqueue/ {
   hash = substr($0,index($0,"R-"),8)
   reciever = substr($1,1,7)
   
  if(recvs[hash][reciever] ==  "") {
     recvs[hash][reciever] =  $3;
  }
  
#  print "recv", hash, reciever, recvs[hash][reciever];
  
}

END {

  PROCINFO["sorted_in"] ="@val_num_asc";
  for(hash in sends) {
     printf("%s %2d %s ", hash, length(recvs[hash]), sends[hash])
     min = 99999999;
     max = 0;
     list = ""
     # line 42
     #print "h", hash
     if (hash in recvs) {
       for(reciever in recvs[hash]) {
         if(max < recvs[hash][reciever])  {
            max = recvs[hash][reciever]; 
         }
         #print "r", reciever, recvs[hash][reciever], max
         list = list  sprintf("%s:%s ", reciever, recvs[hash][reciever]);
       }
    }
    printf(" [%s,%s]",  max,sends[hash]);
    printf("%7.3g %s", time2sec(max)-time2sec(sends[hash]), list);
    printf("\\n");
  }

}


EOF
################################
# End of AWK Scripts           #
################################



 (grep -HE "Send broadcast" fnode*_networkoutputs.txt; grep -HE "enqueue" fnode*_networkinputs.txt;) | awk "$scriptVariable"  | less -R

