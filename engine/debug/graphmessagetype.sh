#/bin/sh
pattern="$1"
binsize="$2"
shift
shift
################################
# AWK scripts                  #
################################
  #  27 16:57:39.729 1-:-0 peer-1, enqueue     M-93fe7a|R-93fe7a|H-93fe7a|0xc0032e7000                Missing Msg[16]:MissingMsg --> 1570f8<FNode01> asking for DBh/VMh/h[1/0/0, ] Sys: 0 msgHash[93fe7a] from peer-2  FNode01
  #   28 16:57:39.729 1-:-0 peer-1, enqueue     M-93fe7a|R-93fe7a|H-93fe7a|0xc0032e7000                Missing Msg[16]:MissingMsg --> 1570f8<FNode01> asking for DBh/VMh/h[1/0/0, ] Sys: 0 msgHash[93fe7a] from peer-2  FNode01
  #   30 16:57:39.729 1-:-0 peer-2, enqueue     M-6c6bc1|R-6c6bc1|H-6c6bc1|0xc0032e7100                Missing Msg[16]:MissingMsg --> 8da6ed<FNode02> asking for DBh/VMh/h[1/0/0, ] Sys: 0 msgHash[6c6bc1] from peer-3  FNode02
  #
read -d '' scriptVariable << 'EOF'
func time2sec(t) {	
  x = split(t,ary,":");
  if(x!=3) {printf("time2sec(%s) bad split got %d fields. %d", t , x ,NR); print $0; exit;}
  sec = (ary[1]*60+ary[2])*60+ary[3];
#  printf("time2sec(%s) %02d:%02d:%02d= %d\\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
  return sec;
}

 # seperate filename and print warm fuzzy
 { sub(/:/," ");
   fname = substr($1,1,7);
   if(NR%1000==0) {printf("\\r%7d",NR)>"/dev/stderr";}
  # print FILENAME":"FNR,$0 > "/dev/stderr";

 }

 # remeber start time
 NR==1 {
     starttime = time2sec($3);
 }
 
 /continue:/ {
 #   print "Skip:", FILENAME":"FNR,$0 > "/dev/stderr"
    next;
 }
 
# remove duplicates mesages
 { match($0,/R-/);
   hash = substr($0,RSTART,8);
   
   if(hash in dups) {
 #    print FILENAME":"FNR,"drop", hash > "/dev/stderr";
     next;
   }
   dups[$6]++;
 }
 
/Send|[Ee]nque/ {
   match($0,/\[[ 0-9]?[0-9]\]/)
   type = substr($0,RSTART-27,26);
   gsub(/[\\t ]+/,"",type);
   bins[i] = $4;
   if(type == "Ack") {
     if ($0 ~ /EmbeddedMsg: [0-9a-f]{6}/) {
        getline <= 0;
        embedded = index($0,"\]:")+2; # add length of the ]: tag
        match(substr($0,embedded),/[A-Za-z]+/);
        subtype = substr($0,embedded+RSTART-1,RLENGTH);
        type = "Ack:" subtype;
   #     print type, embedded+RSTART-1, RLENGTH> "/dev/stderr";
   #     print $0> "/dev/stderr";
   #     printf("%*s%*s\n",embedded+RSTART-1,"^",RLENGTH,"^")> "/dev/stderr";
     
     } else {
        embedded = index($0,"EmbeddedMsg:")+12; # add length of the EmbeddeMsg: tag;
        match(substr($0,embedded),/[A-Za-z]+/);
        subtype = substr($0,embedded+RSTART-1,RLENGTH);
        type = "Ack:" subtype;
  #      print type, embedded+RSTART-1, RLENGTH> "/dev/stderr";
  #      print $0> "/dev/stderr";
  #      printf("%*s%*s\n",embedded+RSTART-1,"^",RLENGTH,"^")> "/dev/stderr";
     }
   }
   types[type]++;
   time = time2sec($3);
   #print FILENAME":"FNR,"|"type"|",binsize, time-starttime >"/dev/stderr";
   i = int((time - starttime)/binsize);
   count[i][type]++
 #  print type, i, count[i][type];
 }
 

END {
  print "" >"/dev/stderr";

  PROCINFO["sorted_in"] ="@ind_str_asc";
  printf("%8.8s,", "bin");
  for(f in types) {
     printf("%26.26s,", f);
  }
  printf("\\n");

  PROCINFO["sorted_in"] ="@ind_num_asc";
  for(i in count) {
    printf("%8.8s,", bins[i]);
    for(f in types) {
     #print i, f, "|"count[i][f]"|";
     printf("%26.26s,", count[i][f]+0);
    }
    printf("\\n");
  }
}
EOF
################################
# End of AWK Scripts           #q
################################


(grep -HEi "enqueue" ${pattern}_networkinputs.txt | grep -v Ack | grep -v Embedded| grep -v EOM;grep -HE "ACK|EOM|EmbeddedMsg" ${pattern}_networkoutputs.txt | grep -E "Send |EmbeddedMsg" )  |   awk -v binsize=$binsize "$scriptVariable" 
