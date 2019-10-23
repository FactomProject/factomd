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
function print_unfilledgaps(){
    PROCINFO["sorted_in"] = "@ind_num_asc"
    
    for(dbht in fill) {
       for(vm in fill[dbht]) {
          for(h in fill[dbht][vm]) {
              if(fill[dbht][vm][h] < 0) {
                  printf("%6d/%02d/%-5d \\n", dbht,vm,h);
              }
              #printf("  %6d/%02d/%-5d %s\\n", dbht,vm,h,processlist[dbht][vm][h]);
          }
          #print "";
       }
    }
}

func time2sec(t) {    
  x = split(t,ary,":");
  if(x!=3) {printf("time2sec(%s) bad split got %d fields. %d", t , x ,NR); print $0; exit;}
  sec = (ary[1]*60+ary[2])*60+ary[3];
#  printf("time2sec(%s) %02d:%02d:%02d= %d\\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
  return sec;
}

BEGIN {
    if(time ~ /[0-9][0-9]:[0-9][0-9]:[0-9][0-9](.[0-9])?/) {
        time_regex = "^ *[0-9]+ " time;
    } else {
        time_regex = "^ *" time " ";
    }
}

/Added/ {
    split($6,ht,"/"); # split dbht/vm/height
    message = substr($0,index($0,"M-"));
    #print "Add " ht[1] "/" ht[2] "/" ht[3] ":" message;

    # REVIEW: is there a way to set an empty array?
    processlist[ht[1]][ht[2]][ht[3]] = message;

    if(length(processlist[ht[1]][ht[2]]) == 1 ) {
        # started a new block/vm
        # printf("new vm %s \\n", ht[2]);
        position[ht[1]][ht[2]] = 0;
    } 

    plht = position[ht[1]][ht[2]]

    if(plht == ht[3]) {
        position[ht[1]][ht[2]] = plht + 1;
    } else {
        if(plht < ht[3]) {

            for(i=plht; i < ht[3]; ++i) {
                fill[ht[1]][ht[2]][i] = -1;
                gap[ht[1]][ht[2]][i] = time2sec($2);
                # found ht[3] expected plht
                printf("gap  %-5d %6d/%02d/%-5d %0.2f\\n", ht[3], ht[1], ht[2], i, 0);
            }
            position[ht[1]][ht[2]] = ht[3] + 1;
        } else {
            fill[ht[1]][ht[2]][ht[3]] = time2sec($2);
            delta = fill[ht[1]][ht[2]][ht[3]] - gap[ht[1]][ht[2]][ht[3]]
            printf("%s %-5d %6d/%02d/%-5d %0.2f\\n", "fill", plht, ht[1], ht[2], ht[3], delta);
        }
    }

    next;
}

END {
    print  "\\n\\nfound end of time at " Line, NR":", $1;
    print "\\nunfilled gaps:\\n---------------"
    print_unfilledgaps();
    exit(0);
}
EOF
################################
# End of AWK Scripts           #
################################
echo grep -hE -E "${time:-max}|Added at ${height:-[0-9]+}/${vms:-[0-9]+}/|done ${height:-[0-9]+}/${vms:-[0-9]+}/" ${fnode:-fnode0}_processlist.txt 
grep -hE -E "${time:-max}|Added at ${height:-[0-9]+}/${vms:-[0-9]+}/|done ${height:-[0-9]+}/${vms:-[0-9]+}/" ${fnode:-fnode0}_processlist.txt | tee x | awk -v time=${time:-max} -v height=${height:-[0-9]+} -v vms=${vms:-[0-9]+} "$scriptVariable"

