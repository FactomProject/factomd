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
   if(NR%1000==0) {printf("\\r%7d",NR)>"/dev/stderr";}
 }

/enqueue/ { 
    fnames[fname]++
    l = fnames[fname]
    timestamps[fname][l] =   time2sec($3)
    ts_length[fname]=l
    if(NR%1000==0) {printf("\\r%7d",NR)>"/dev/stderr";}
#    print "input", fname, fnames[fname],	timestamp[fname][l]
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



grep -HE "enqueue" FNode*NetworkInputs.txt  |   awk "$scriptVariable" 
