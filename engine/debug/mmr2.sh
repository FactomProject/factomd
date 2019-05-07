#!/usr/bin/env bash
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'
func time2sec(t) {	
  x = split(t,ary,":");
  if(x!=3) {print "bad time",NR, $0; exit;}
  sec = (ary[1]*60+ary[2])*60+ary[3];
#  printf("time2sec(%s) %02d:%02d:%02g= %d\\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
  return sec;
}

func time2str(seconds) {
 return sprintf("%02d:%02d:%02d.%03d", seconds/3600, (seconds/60)%60, seconds%60, int((seconds-int(seconds))*1000));
}

func sum_of_array(a) {
    s = 0;
    for(i in a){s+=a[i];}
    return s;
}

func min_of_array(a) {
    for(i in a){min=a[i];break;} # set min to a valid value
    for(i in a){if(min > a[i]){min = a[i];}}
    return min;
}

func max_of_array(a) {
    for(i in a){max=a[i];break;} # set max to a valid value
    for(i in a){if(max < a[i]){max = a[i];}}
    return max;
}

func min_index_of_array(a) {
    for(i in a){min=i;break;} # set min to a valid value
    for(i in a){if(min > i){min = i;}}
    return min;
}

func max_index_of_array(a) {
    for(i in a){max=i;break;} # set max to a valid value
    for(i in a){if(max < i){max = i;}}
    return max;
}



NR==1 {fts = time2sec($2)}

#    87670 16:32:13.982       5-:-0 Send P2P FNode03                       M-76bd96|R-76bd96|H-76bd96|0xc43c2e0a00                Missing Msg[16]:MissingMsg --> 455b7b<FNode0> asking for DBh/VMh/h[5/0/1, ] Sys: 0 msgHash[76bd96] from peer-0  RandomPeer
/MissingMsg / {
    ts = time2sec($2);

    cnt = match($0,/([0-9]+\\/[0-9]\\/[0-9]+, )+/,ary);
    list = substr($0,RSTART,RLENGTH);
    n = split(list,ary,", ");
    for( i in ary ) {
      id = ary[i];
      if(id~/[0-9]+\\/[0-9]+\\/[0-9]+/) {
        asks[id]++;
        if(!(id in first_ask)) {
          first_ask[id] = ts;
          first_ask_r[id] = $2;
          inflight++;
          stat[ts-fts]=inflight;
        }
        last_ask[i] = ts;
      }
    }
}

#144123 16:32:44.878       7-:-1 peer-1, enqueue                        M-4c200f|R-4c200f|H-4c200f|0xc57ef183c0       Missing Msg Response[19]:MissingMsgResponse <-- DBh/VMh/h[          7/1/2] message    EOM-     DBh/VMh/h 7/1/-- minute 1 FF  0 --Leader[aeaac8<FNode03>] hash[aa99a7]  msgHash[4c200f] to peer-2  FNode01
/MissingMsgResponse/ {
    ts = time2sec($2);
    id = substr($12,1,length($12)-1);
    resps[id]++;
    if(!(id in first_resp)) {
        first_resp[id] = ts;
        delay[id] = ts - first_ask[id];
        inflight--;
        stat[ts-fts]=inflight;
    }
    last_resp[id] = ts;
}

END {

    sum_of_asks = sum_of_array(asks);
    print "MMRequests:   total=", sum_of_asks, "unique=", length(asks), "min=", min_of_array(asks), "aver=", sum_of_asks/length(asks), "max=", max_of_array(asks);

    sum_of_resps = sum_of_array(resps);
    print "MMResponces:  total=", sum_of_resps, "unique=", length(resps), "min=", min_of_array(resps), "aver=", sum_of_resps/length(resps), "max=", max_of_array(resps);

    sum_of_delay = sum_of_array(delay);
    max_delay = max_of_array(delay);
    min_delay =  min_of_array(delay);
    printf("Time for reply:  total= %s        min=%g aver=%g max=%g\\n", time2str(sum_of_delay), min_delay, sum_of_delay/length(delay), max_delay);

    sample_period = (max_delay-min_delay)/10;
    
    for(i in delay) {
        sample_idx = int((delay[i]-min_delay)/sample_period);
        histogram[sample_idx]++;
    }
    
    PROCINFO["sorted_in"] ="@ind_num_asc";
    print "histogram of delays in", sample_period, "second buckets"
    print "seconds count";
    for(i in histogram){printf("%7g %d\\n", i*sample_period + sample_period/2, histogram[i]);}
    print "---------------------"
    
    
    
    
    print "unique requests in-flight average", sum_of_array(stat)/length(stat), "max=", max_of_array(stat);
    min_inflight = min_index_of_array(stat); # time first ask 
    max_inflight = max_index_of_array(stat); # time of last response
    sample_period = (max_inflight-min_inflight)/10;
    
    print "Start of trace", time2str(min_inflight), "duration", time2str(max_inflight-min_inflight), "sample period", sample_period;
    for(i in stat) {
        sample_idx = int((i-min_inflight)/sample_period);
        histogram2[sample_idx]+=stat[i];
        count[sample_idx]++;
    }
    
    PROCINFO["sorted_in"] ="@ind_num_asc";
    print "histogram of in-flight requests in", sample_period, "second buckets"
    print "time          count";
    for(i in histogram2){
       printf("%s %d\\n", time2str(i * sample_period),histogram2[i]/count[i]);
    }
    print "---------------------"



    PROCINFO["sorted_in"] ="@val_num_desc";
    print "Asks ----------------"
    x =0;
    for(i in asks) { print i, asks[i]; x++; if(x==15) break;}
    print "---------------------"

    print "Delays --------------"
    x =0;
    for(i in delay) { print i, delay[i]; x++; if(x==10) break;}
    print "---------------------"

    print "inflight ------------"
    x =0;
    for(i in stat) { print i, stat[i]; x++; if(x==10) break;}
    print "---------------------"

        #for(i in stat) {print i, stats[i];}

    
}
EOF
################################
# End of AWK Scripts           #
################################

 
(grep -hE "Send.*:MissingMsg " $1_networkoutputs.txt; grep -hE "enqueue.*:MissingMsgResponse" $1_networkinputs.txt) | sort -n | awk "$scriptVariable"

