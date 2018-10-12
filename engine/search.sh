#/bin/sh
#set -x

# git log v5.4.2 | head -n 1092 | grep -B5 " Merge " | grep -E "^commit" | awk ' {print $2;}' > x

test="$1"
shift
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'

func mysystem(cmd) {
    print cmd;
    print "\n",cmd,"\n" ;
    rc = system(cmd " 2>&1");
    if(rc!=0) {
      print "Exit", rc;
      exit(rc);
    }
    return rc;
}


/^#/ {print "Skipping", $1; next;}

 {branches[x++] = $0;}


END {
 PROCINFO["@ind_num_asc"]
    for(i in branches) {
      branch = branches[i];
      print "Testing Branch ", branch;
      mysystem("git reset --hard");
      mysystem("git checkout " branch);
      mysystem("git checkout Rolling_DBSIG_Test_v5.2 factomd_test.go");
      rc = system("go test --run " test);
      print i, branch, "Exit Code", rc
    }
 }
EOF
################################
# End of AWK Scripts           #
################################
rm -f runlog.txt
awk -v test=$test "$scriptVariable" $@ 2>&1 | tee -i runlog.txt  | grep -E "Testing Branch|Exit Code"


