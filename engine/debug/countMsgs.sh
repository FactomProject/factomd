################################
# AWK scripts                  #
################################
pattern="$1"
shift

read -d '' scriptVariable << 'EOF'
 { sub(/:/," "); # fix files where the sequence number grew upto the filename
   hash = substr($0,index($0,"M-"),26)
   endHash = index($0,hash) + length(hash)
   sub(/|0x[0-9a-f]+/,"",$0) # remove address of message if it is present
   rest = substr($0,endHash)
   msgType = substr(rest,0,index(rest,":")-1)
   gsub(/ +/,"",msgType);   # remove spaces from message type
# print;
# print "<"msgType"><"hash">"
   if (tcount[msgType]++ == 0) { # count the types messages
      #print $0
   }
   if(mcount[hash]++ >1){ 
      dcount[msgType]++ # count the duplicate messages based on hash
   }
}
END {
    printf("%-30s %7s %7s\\n", "typr","count","duplicates");
    PROCINFO["sorted_in"] ="@ind_str_asc";
    for(i in tcount) {
       printf("%-30s %7d %7d\\n", i, tcount[i], dcount[i]);
    }
    print "Most duplicated messages"
    PROCINFO["sorted_in"] ="@val_num_desc";
    for(i in mcount) {
       if (mcount[i] >2) {
          printf("%3d %s\\n", mcount[i], i);
       }
       if(x++>10) break;
    }
}
EOF
################################
# End of AWK Scripts           #
################################


grep -H  "$pattern" $@  | awk "$scriptVariable" 
