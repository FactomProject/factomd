#!/usr/bin/env bash
# import data into first sheet in https://docs.google.com/spreadsheets/d/1sUYnMiI02f4hkYFoppb6DHup4eZv3wIAm4JZmNccnxk/edit?usp=sharing
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'

func time2sec(t) {    
    x = split(t,ary,":");
    if(x!=3) {printf("time2sec(%s) bad split got %d fields. %s:%d", t , x ,FILENAME,FNR); print $0; exit;}
    sec = (ary[1]*60+ary[2])*60+ary[3];
    #printf("time2sec(%s) %02d:%02d:%02d= %d\\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
    return sec;
}

# {print;}

NR%1000 == 0 { printf("%20d\\r", NR) > "/dev/stderr"; }

BEGIN {
    printf("%10s\\t%10s\\t%10s\\t%6s\\t%s\\t%s\\t%s\\n", "hash","Q","A", "delay","Qcount","Acount","inflight");
}
  
#   267679 16:28:58.824  203562-:-0 Send P2P P2P Network                   M-2d050e|R-2d050e|H-2d050e|0xc00133f450               Missing Data[17]:MissingData: [3d3b5aa78f] RandomPeer 
/Send.*MissingData:/ {
  hash = substr($11,2,10);
  if(!(hash in requests)) {
     inflight++;
     requests[hash] = time2sec($2);
     reqcount[hash] = 1;
  } else {
     reqcount[hash]++;
  }
}


#   267540 16:28:58.687  203561-:-0 peer-0, enqueue                        M-3b7d2a|R-3b7d2a|H-3b7d2a|0xc009b77180              Data Response[18]:DataResponse Type:  0 Hash: ae761a7d66393b703195bbcf4c0a2143c1a8b3900e6d56c29aed8090a38f1cbf FNode00 
/enqueue.*DataResponse/ {
  hash = substr($12,1,10);
#  print hash;
  if(!(hash in respcount)) {
    inflight--;
    response[hash]=time2sec($2);
    respcount[hash] = 1;
  } else {
    respcount[hash]++;
  }
  printf("%s\\t%10.3f\\t%10.3f\\t%6d\\t%s\\t%s\\t%s\\n",hash,requests[hash],response[hash],(response[hash]-requests[hash])*1000, reqcount[hash],respcount[hash],inflight);
}




EOF
################################
# End of AWK Scripts           #
################################
(grep -E "Send.*MissingData:" fnode0_networkoutputs.txt; grep -E "enqueue.*DataResponse" fnode0_networkinputs.txt) | sort -n | gawk  "$scriptVariable" $@ > sync.csv
