	{
	gsub(/[:{]/," "); 
        gsub(/ FNode../,"");
	if($1!= fname) {fname = $1}
	list[fname] = list[fname] substr($10,6,8) "\n";
	}

END 	{
          for(fname in list) {
		printf("--%30s--\n%s\n",fname, list[fname]);
	  }
	  print "";
      for(fname in list) {
		printf("%30s ",fname);
	  }
	  print "";
	  count = split(list[fname],msgs,"\n")
          for(i=0;i<count;i++){
	    for(fname in list) {
		split(list[fname],msgs,"\n")
                printf("%30s ",msgs[i]);
	    }
	    print "";
          }
	}
