#!/usr/bin/env bash

pattern="$1"
shift
################################
# AWK scripts                  #
################################
#filname:2599920 09:59:00 159028-:-0 peer-0, enqueue M-466a4e|R-7b886f|H-466a4e                        EOM[ 0]:   EOM-DBh/VMh/h 159029/13/-- minute 0 FF  0 --Leader[07e4f3] hash[466a4e] 
read -d '' scriptVariable << 'EOF'
{  fname = substr($0,1,index($0,":"));
   rest = substr($0,length(fname)+1); # +1 to exclude the :
   sub(/fnode0_/,"fnode0__",fname);
   sub(/\\.txt/,"",fname);
   sub(/^ +/,"",rest); #trim leading spaces if any
   seq =   substr(rest,0,index(rest," "));
   rest = substr(rest,length(seq)+1);
      
   match(rest,/[0-9]+.[0-9]+.--/,result);
   split(result[0],t,"/");
   dbh = t[1];
   vm = t[2];

   match(rest,/minute [0-9]/,result);
   minute = result[0];
   sub(/minute /,"",minute);
   
   sync[dbh][minute][vm] = vm;
   
}

dbh != prev_dbh {
  PROCINFO["sorted_in"] = "@ind_num_asc";

  if (prev_dbh  != "") {
    print "New DBH", dbh;
   
    for(m in sync[prev_dbh]) {
       i = 0;
       printf("%8d-:-%d ",prev_dbh,m);
       for(vm in sync[prev_dbh][m]) {
          while(i<vm) {printf(".. ");i++;}
          printf("%02d ",vm);
          i++;
       }
       printf("\\n");
    }
  }
   
   prev_dbh = dbh;
}

EOF
################################
# End of AWK Scripts           #
################################



grep -HE "EOM\[" fnode0_networkinputs.txt | grep -v Drop | awk "$scriptVariable"| less

