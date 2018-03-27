# find . -name vendor -prune -o -name .git -prune -o -name \*.go | grep -v vendor | grep -v git | xargs awk -i inplace -f ~/fix2.awk



func fix(bad) {
    orig = $0
    good = toupper(substr(bad,1,1)) substr(bad,2)

    if (FILENAME ~ /factomParams\.go/) {
        rval += gsub(bad,good,$0)
    }

    rval += gsub("\\."bad,"."good,$0)
}


{
	fix("prefix");                      
	fix("rotate");                      
	fix("timeOffset");            
	fix("keepMismatch");            
	fix("customNet");            
	fix("deadline");             
	fix("exposeProfiling");      
	fix("factomdLocations");     
	fix("factomdTLS");           
	fix("fast");                 
	fix("fastLocation");         
	fix("logjson");              
	fix("loglvl");               
	fix("logstashURL");          
	fix("memProfileRate");       
	fix("pluginPath");           
	fix("rpcPassword");          
	fix("rpcUser");              
	fix("svm");                  
	fix("torManage");            
	fix("torUpload");            
	fix("useLogstash");  
	if (rval != prval) {
		printf("\r%5d",rval) > "/dev/stderr";
		prval = rval;
        }
	print;        
}

