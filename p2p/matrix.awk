


func findcalls(line,list) {
  delete list
  do {  
	m = match(line,/c\.[A-Za-z0-9_.]+/,a)
#	printf("Match %d\n", m)
	if(m) {
  	  ms = substr(line,RSTART,RLENGTH)
#	  printf("<%s>\n",ms)
	  list[length(list)] = ms
	  line = substr(line,RSTART+RLENGTH)
#          printf(" after <%s>\n",line)
	}
  }while (m);
  
}



/func / {
	match($0,/func \(.*\) (.*)\(.*\{/,list)
#	print $0,"->",list[1]
	current_func = list[1]
	funcs[current_func] = current_func ":"
	next;
}

/c\./ {
#   print $0
   findcalls($0,list); 
   for( i in list) {
      f = list[i]
      funcs[current_func] = funcs[current_func] ", "  f
#      printf("%s %d <%s> [%s]\n", current_func, i, f, funcs[current_func]);
    }
#   print ""
}


END {
    printf("Functions:,");
  
    for(i in funcs) {
	printf("%s\n",funcs[i]);
	}	
     print "";
}
