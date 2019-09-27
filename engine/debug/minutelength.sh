#!/usr/bin/env bash
#grep "done.*EOM.*./0/" fnode0_processlist.txt  | ./minutelength.sh


column=$1
shift
format=$1
shift


echo "| minutelength.sh" ${column:-2} ${format:-sec}


################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
func time2sec(t) {    
    x = split(t,ary,":");
    if(x!=3) {
        printf("time2sec(%s) bad split got %d fields.\\n",t, x)
        printf("Line:  %s:%d\\n", FILENAME,FNR); 
        print "<"$0">"; 
        exit;
    }
    sec = (ary[1]*60+ary[2])*60+ary[3];
    #printf("time2sec(%s) %02d:%02d:%02d= %d\\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
    return sec;
}

func timeDiff(t1,t2){

    tDiff = time2sec(t1)-time2sec(t2);
    if(tDiff < 0) {
        tDiff = tDiff+24*60*60;
    }
    if(format == "hms" || format == "HMS") {
       return sprintf("%2d:%02d:%02d.%03d", (tDiff/(60*60)),(tDiff/60)%60,tDiff%60,(tDiff - int(tDiff))*1000 );
    } else {
       return sprintf("%9.3f", tDiff);
    }
}

 BEGIN { print "using column:", column, "format:",format;}

 { sub(/\\(standard input\\):/,"");
   
   #sub(/:/," ");
   fname = $1;
   if(NR%1000==0) {printf("\\r%7d",NR)>"/dev/stderr";}
}

NR == 1 {prev = $(column);}

 { now = $(column);
   delay = timeDiff(now,prev);
   gap[delay]++;
   gapsrc[delay] = $0;
   printf("%s %s\\n", delay, $0);
   prev = now;
 }

END {
  if(format == "hms" || format == "HMS") {
     PROCINFO["sorted_in"] ="@ind_num_desc";
  } else {
     PROCINFO["sorted_in"] ="@ind_str_desc";
  }
  print "Gaps in log"
   for(i in gap) {
     if(format == "hms" || format == "HMS") {
       printf("%s %4d %s\\n", i, gap[i], gapsrc[i]);
    } else {
       printf("%9.2f %4d %s\\n", i, gap[i], gapsrc[i]);
    }
       if(j++>10) {break;}
   }
 

}

EOF
################################
# End of AWK Scripts           #
################################

 awk -v column=${column:-2} -v format=${format:-sec} "$scriptVariable"

