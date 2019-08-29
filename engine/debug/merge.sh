#/bin/sh
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
/[0-9]{4}-[0-9]{2}-[0-9]{2} / {next;} # drop file time stamp
{
   sub(/from /,"")
   l = index($0,":") # find the end of the file name
   fname = substr($0,1,l); #seperate that
   seq =   substr($0,l+1,9); #grab the sequence number
   rest = substr($0,l+11) 
   gsub(/^ +/,"",rest); # compress leading spaces 
   
#   printf("%d <%s><%s><%s>\\n", l, fname, seq, rest);
   
   m = index(rest,"M-") # find the message hash
   if(m==0) {
     note = rest;
     msg ="";
   } else {
     note = substr(rest,1,m-1); # seperate the note
     msg = substr(rest,m);      # from the message
   }
   printf("%-30s %s  %-40s %s\\n", fname, seq, note, msg);
}

EOF
################################
# End of AWK Scripts           #
################################
if [$# -ne 0]; then
grep -HE . "$@"  | awk  "$scriptVariable" | sort -nk2 | less -R
else
awk  "$scriptVariable" | sort -nk2 | less -R
fi
