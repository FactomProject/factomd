#/bin/sh
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
 {
    n = substr($1,1,1); 
    print n; 
    cmd = "tar xzvf "$1;
    print cmd;
    system(cmd); 
    foo = sprintf("%02d", n)
    cmd = "ls *fnode0_*.txt | awk ' {f = $1; g = tolower(f); sub(/0_/,\\"" foo "_\\",g); if(f!=g) {cmd=\\\"mv -v \\\" f \\\" \\\" g; print cmd; system(cmd);}}'";
    print cmd, n;
    system(cmd); 
    cmd = "mv out.txt out" foo " .txt"
    print cmd, n;
    system(cmd); 
    cmd = "mv err.txt err" foo " .txt"
    print cmd, n;
    system(cmd); 
   

}
EOF
################################
# End of AWK Scripts           #
################################



 ls -r *.tgz | awk "$scriptVariable"
