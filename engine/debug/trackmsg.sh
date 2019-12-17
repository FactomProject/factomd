#!/usr/bin/env bash

# trackmsg.sh <pattern> <node>

if [ "$#" -gt 2 ] || [ "$#" -lt 1 ] ; then
    echo "trackmsg.sh <pattern> <node>"
    exit 1
fi


pattern="$1"
shift
fnode=$1
shift

################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
{  fname = substr($0,1,index($0,":"));
   rest = substr($0,length(fname)+1); # +1 to exclude the :
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
   printf("%s %-30s %-45s %s\\n",seq, fname, note, msg);
}
EOF
################################
# End of AWK Scripts           #
################################
echo "grep -E \"$pattern\" ${fnode:-fnode0_}* | awk -f msgOrder.awk | sort -n | grep -E \"$pattern\" --color='always' | less -R"
grep -H -E "$pattern" ${fnode:-fnode0_}* | awk "$scriptVariable" | sort -n | grep -E "$pattern" --color='always' | less -R
