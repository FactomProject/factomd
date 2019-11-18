#!/usr/bin/env bash

# processlist.sh <node> <height> <vms>

if [ "$#" -gt 4  ]; then
    echo "processlist.sh <time> <height> <vms> <node> "
fi

time=$1
shift
height=$1
shift
vms=$1
shift
fnode=$1
shift


echo "processlist.sh" ${time:-max} ${height:-[0-9]+} ${fnode:-fnode0} ${vms:-[0-9]+}


################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'

function print_processlists(){
    printf("\\r%8d\\n",NR)>"/dev/stderr";
    PROCINFO["sorted_in"] = "@ind_num_asc"
    
    for(dbht in processlist) {
       for(vm in processlist[dbht]) {
          for(h in processlist[dbht][vm]) {
            if(dbht in donelist && vm in donelist[dbht] && h in donelist[dbht][vm]) {
              printf("P %6d/%02d/%-5d %s\\n", dbht,vm,h,processlist[dbht][vm][h]);
            } else {
              printf("  %6d/%02d/%-5d %s\\n", dbht,vm,h,processlist[dbht][vm][h]);
            }
          }
          print "";
       }
    }
}

# {print NR,$0;}

BEGIN {
    if(time ~ /[0-9][0-9]:[0-9][0-9]:[0-9][0-9](.[0-9])?/) {
        time_regex = "^ *[0-9]+ " time;
    } else {
        time_regex = "^ *" time " ";
    }
    
    print "time",time, "height", height, "vms",vms;
    print "time_regex", time_regex;
}

#warm fuzzy while working
 {if(NR%1000==0) {printf("\\r%8d",NR)>"/dev/stderr";}}

$0 ~ time_regex {
    print  "found end of time at " Line, NR":", $0;
    print_processlists();
    exit(0);
}

END {
    print_processlists();
    exit(0);
}



#  1616362 00:04:54.018  205916-:-0 Added 205917/10/5                  M-da9376|R-bbae5d|H-da9376|0xc008c9b7c0                        EOM[ 0]:   EOM-DBh/VMh/h 205917/10/-- minute 4 FF  0 --Leader[598240] hash[da9376]   

/Added/ {
    split($5,ht,"/"); # split dbht/vm/height
    message = substr($0,index($0,"M-"));
    #print "Add " ht[1] "/" ht[2] "/" ht[3] ":" message;
    processlist[ht[1]][ht[2]][ht[3]] = message;
    next;
}


#  3178219 00:31:31.771  205919-:-0 done  205917/10/5                  M-0af8f2|R-8f90c3|H-0af8f2|0xc0064bc500                        EOM[ 0]:   EOM-DBh/VMh/h 205919/23/-- minute 0 FF  0 --Leader[9b2295] hash[0af8f2]   
/done/ {
    split($5,ht,"/"); # split dbht/vm/height
    message = substr($0,index($0,"M-"));
    #print "done" ht[1] "/" ht[2] "/" ht[3] ":" message;
    donelist[ht[1]][ht[2]][ht[3]] = message;
    next;
}



EOF
################################
# End of AWK Scripts           #
################################

echo grep -hE -E "${time:-max}|Added ${height:-[0-9]+}/${vms:-[0-9]+}/|done ${height:-[0-9]+}/${vms:-[0-9]+}/" ${fnode:-fnode0}_processlist.txt 

grep -hE "${time:-max}|Added ${height:-[0-9]+}/${vms:-[0-9]+}/|done ${height:-[0-9]+}/${vms:-[0-9]+}/" ${fnode:-fnode0}_processlist.txt | tee x | awk -v time=${time:-max} -v height=${height:-[0-9]+} -v vms=${vms:-[0-9]+} "$scriptVariable"
