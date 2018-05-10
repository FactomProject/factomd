 {
   sub(/from /,"")
   l = index($0,":")
   fname = substr($0,1,l);
   seq =   substr($0,l+1,20);
   rest = substr($0,l+17)
   m = index(rest,"M-")
   if(m==0) {
     note = rest;
     gsub(/^ +/,"",note);
     msg ="";
   } else {
     note = substr(rest,1,m-1);
     gsub(/^ +/,"",note);
     msg = substr(rest,m);
     gsub(/^ +/,"",msg);
   }
   printf("%s %-30s %-40s %s\n",seq, fname, note, msg);
}
