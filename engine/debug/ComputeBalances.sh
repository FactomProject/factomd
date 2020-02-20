#/bin/sh
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'

func time2sec(t) {	
  x = split(t,ary,":");
  if(x!=3) {printf("time2sec(%s) bad split got %d fields. %d", t , x ,NR); print $0; exit;}
  sec = (ary[1]*60+ary[2])*60+ary[3];
#  printf("time2sec(%s) %02d:%02d:%02d= %d\\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
  return sec;
}

 {
   if(NR%1000==0) {printf("\\r%7d",NR)>"/dev/stderr";}
}

# 438 17:33:43.063     256-:-0 Loading 35 FTC balances from DBH 256 

/Loading [0-9]+ FTC balances from DBH / {
    DBHT = $11
    count = $6
}


# 439 17:33:43.063     256-:-0 17e5dc5999de0116483d64982e9a821e00b45f5f31e3f414e54547e78b1bde90<FA29bzmmasdRRcLjW7ZJsqyQXFTpZyUUnvRbxGGZJLxCM9kS7Snq> = 2000000000000 
      
/[0-9a-zA-Z]+<FA[[0-9a-zA-Z]+> = / {
    match($5,/<.*>/)
   addr = substr($5,RSTART+1,RLENGTH-2)
   balance[addr] = $7/100000000
   t = $4
   printf("bal @ %7s %s %15.7f\\n", t, addr,  balance[addr] )
}


#626 17:33:43.186       1-:-0 At 0 process rt =false Transaction TXID: 914333898b4cd3a87091ced94d6276090a1a266e1f4b7578e2b036cfaf9aaf3e (size 1096): 
/Transaction TXID: / {
    t = $4;
}

# 626 17:33:43.186       input:       500000.00253    FA2nmHSnesq3KQSDGXS2qMT9Jk4U4cJzWGxQ55urbJdcSRMmtXvG

/input:/ {
    addr = $6
    balance[addr]-=$5
   printf("bal @ %7s %s %15.7f\\n", t, addr,  balance[addr] )
}

# 626 17:33:43.186      output:        20000.0        FA2Ucjw1pBzrCxDJP82yCbhzk9BFU2aWdb1VqitBot3mFAseW6Uo
/output: +[0-9]+\.[0-9]+ +FA/ {
    addr = $6
    balance[addr]+=$5
   printf("bal @ %7s %s %15.7f\\n", t, addr,  balance[addr] )
}


END {
    printf("\\r%7d\\n",NR)>"/dev/stderr"
  
}

 
EOF
################################
# End of AWK Scripts           #
################################

 


grep -H . $@_factoids.txt| gawk  "$scriptVariable" 
