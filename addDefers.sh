#/bin/sh
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
BEGIN {
  pattern = "(func \\\\(.*\\\\*(.*)\\\\) (.*)\\\\(.*\\\\)) interfaces.IHash \\\\{$";
  replacment = "\\\\1(rval interfaces.IHash) {\\ndefer func() {\\n		if rval != nil \\\\&\\\\& reflect.ValueOf(rval).IsNil() {\\n		rval = nil // convert an interface that is nil to a nil interface\\n			primitives.LogNilHashBug(\\"\\\\2.\\\\3() saw an interface that was nil\\")\\\n		}\\n	}()\\n"
#  print "<"pattern"><"replacment">"

}


 {
  if(NR%1000 == 0) printf("\\r%8d %s       ",FNR,FILENAME) > "/dev/stderr";
  if(match($0,pattern)) {
    r = gensub(pattern,replacment,1);
    print r
    } else {
     print;
    }

}

 
EOF
################################
# End of AWK Scripts           #
################################


find .  -name \*.go | xargs grep -lE "func \(.*\) .*\(.*\).* interfaces.IHash {"  | xargs -n 1 gawk  -i inplace  "$scriptVariable" 

#echo  ./common/entryBlock/entry.go  | xargs -n 1 gawk  -i inplace  "$scriptVariable" 

