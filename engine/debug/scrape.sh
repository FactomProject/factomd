#!/usr/bin/env bash
#"===SummaryStart===\n    FNode0[9987d4] L_W_vm14
#2018-10-01T00:13:25.573175842Z
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
 
# {print};

/[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+/ {
  match($0,/[0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+/);
  ip = substr($0,RSTART,RLENGTH);
  match($0,/FNode0\\[......\\] ......../); 
  foo = substr($0,RSTART+7,RLENGTH-7); 
  gsub(/_/," ",foo);
  #gsub(/vm/," ",foo);
  gsub(/]/," ",foo);  
  match($0,/[0-9]+-:-[0-9]/); 
  bar=substr($0,RSTART,RLENGTH); 
#  printf("1- %16s %s %s %s\\n", ip, foo, bar, dtt);
}

/2018/ {
  match($0,/2018.*Z/);
  dtt=substr($0,RSTART,RLENGTH);
#  printf("2- %16s %s %s %s\\n", ip, foo, bar, dtt);
}


END {printf("%16s %s %s %s\\n", ip, foo, bar, dtt);}

EOF
################################
# End of AWK Scripts           #
################################}

(echo  -n $1 " " ;curl -s -m 5  "$1:8090/factomd?item=dataDump&value=" | tac | tac | jq '.DataDump1.RawDump'; echo " "; curl -s -m 5  "$1:8090/factomd?item=peers" | jq '.[] | .Connection.MomentConnected' | sort | grep -m1 "")  | awk "$scriptVariable" 


