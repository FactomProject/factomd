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

 { sub(/:/," ");
   fname = substr($1,1,7);
}

/process/ { 
    fnames[fname]++
    l = fnames[fname]
    timestamps[fname][l] =   time2sec($3)
    ts_length[fname]=l
    if(NR%1000==0) {printf("\\r%7d",NR)>"/dev/stderr";}
#    print "input", fname, fnames[fname],	timestamp[fname][l]
 }


#FNode01_dbsig-eom.txt:   4247 09:59:36.023 5-:-0 ProcessEOM          M-4c9cbd|R-924794|H-4c9cbd|0xc422044000                        EOM[ 0]:   EOM-     DBh/VMh/h 5/2/-- minute 0 FF  0 --Leader[1570f8<FNode01>] hash[4c9cbd]  

/ProcessEOM +M-/ && $4!=current {
    start_length[fname]++
    eom_start[fname][start_length[fname]++] =  time2sec($3)
    current=$4
    if(max_start_length < start_length[fname]) {max_start_length = start_length[fname];}
}

/ complete / {
    eom_end[fname][length(eom_end[fname])] =  time2sec($3)
}



END {
  print "" >"/dev/stderr";
  first = 999999999
  for(f in fnames) {
     t_first = (timestamps[f][1]);
     t_last =  (timestamps[f][ts_length[f]]);
     if(first > t_first) {first = t_first;}
     if(last < t_last) {last = t_last;}
#     print "work",f, ts_length[f], t_first, t_last, first, last;
  }

  step = (last-first)/600

#  print "range", first, last, step 
  max_idx = 0
  for(f in fnames) {
#     print "work on", f > "/dev/stderr";
     for(i in timestamps[f]) {
        idx = int((timestamps[f][i]-first)/step)
#        print "work2",f, i, idx, timestamps[f][i]
	results[f][idx]++
        if(idx > max_idx) {max_idx = idx}
     }
  }

# build EOM Sync Table Data
  for(f in start_length) {
     fnames["eom_" f] =  start_length[f] 
  }
  for(i in eom_start[fname]) {
#     print "sync", i
     for(f in fnames) {
#	print "sync2", f
        idx = int((eom_start[f][i]-first)/step)
        edx = int((eom_start[f][i]-first)/step)
#       print idx, edx
        for(j = idx; j<=edx;j++) {
   	   results["eom_" f ][idx] += 100;
#           print "EOM", "eom_" f , idx,  results["eom_" f][idx]
        }
     }
  }

  PROCINFO["sorted_in"] ="@ind_str_asc";
  for(f in fnames) {
     printf("%9s\\t", f);
  }
  print "";
  for(i=1;i<max_idx;i++) {
     for(f in fnames) {
       printf("%9d\\t",results[f][i]);
     }
     print "";
  }
}
EOF
################################
# End of AWK Scripts           #
################################



(grep -HE "done" fnode*_process.txt; grep -HE "ProcessEOM +M-|complete" fnode*_dbsig-eom.txt)  |   awk "$scriptVariable" 
