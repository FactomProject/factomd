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
#  printf("time2sec(%s) %02d:%02d:%06.3g= %d\\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
  return sec;
}

 { 
   if(NR%1000==0) {printf("\\r%7d",NR)>"/dev/stderr";}
}

NR == 1 {prev = time2sec($2);}

 { now = time2sec($2);
   delay = now-prev;
   gap[delay]++;
   gapsrc[delay] = $0;

   printf("%7.3f %s\\n", delay, $0);
   prev = now;
 }

END {
  printf("\\r%7d\\n",NR)>"/dev/stderr"
  PROCINFO["sorted_in"] ="@ind_num_desc";
  print "Gaps in log"
   for(i in gap) {
       printf("%7.3f %4d %s\\n", i, gap[i], gapsrc[i]);
       if(j++>10) {break;}
   }
 

}

EOF
################################
# End of AWK Scripts           #
################################

 
 grep -Eh "$pattern" $@ | awk "$scriptVariable"

