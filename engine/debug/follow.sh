#!/usr/bin/env bash
pattern="$1"
shift
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
{  fname = substr($0,1,index($0,":"));
   rest = substr($0,length(fname)+1); # +1 to exclude the :
   sub(/fnode0_/,"fnode0__",fname);
   sub(/\\.txt/,"",fname);
   sub(/^ +/,"",rest); #trim leading spaces if any
   seq =   substr(rest,0,index(rest," "));
   rest = substr(rest,length(seq)+1)
   m = index(rest,"M-")
   if(m==0) {
     note = rest;
     gsub(/^ +/,"",note); # trim leading spaces
     msg ="";
   } else {
     note = substr(rest,1,m-1);
     gsub(/^ +/,"",note);
     msg = substr(rest,m);
     gsub(/^ +/,"",msg);
   }
   printf("%8s %-30s %-45s %s\\n",seq, fname, note, msg);
}
EOF
################################
# End of AWK Scripts           #
################################

################################
# AWK scripts                  #
################################
read -d '' scriptVariable2 << 'EOF'
 {  print;
 }
EOF
################################
# End of AWK Scripts           #
################################


(grep -HE "Send.*$pattern" fnode*_networkoutputs.txt  
grep -HE "enqueue.*$pattern" fnode*_networkinputs.txt 
ls fnode*_executemsg.txt | xargs -n1 --delimiter "\n" -I%  sh -c "grep -HE \"Execute.*$pattern\" %"
grep -HE "Add.*$pattern" fnode*_processlist.txt 
grep -HE "done.*$pattern" fnode*_process.txt) |  awk "$scriptVariable" | sort -n | awk "$scriptVariable2" | grep -E "$pattern" --color='always' | less -R

