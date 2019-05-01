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



 {}

#   149355 08:46:33.139  189378-:-0 Send P2P P2P Network                   M-931c6a|R-931c6a|H-931c6a|0xc422020270               Missing Data[17]:MissingData: [fc73485610] RandomPeer 
/MissingData/ {
    ts = time2sec($2);
    id = substr($10,2,10);
    print "md",id;
    asks[id]++;
    
    if(!(id in first_ask)) {
      first_ask[id] = ts;
      first_ask_r[id] = $2;
      inflight++;
      stat[ts]=inflight;
    }
    last_ask[i] = ts;
}

#   809450 08:48:24.772  190349-:-0 peer-0, enqueue                        M-50a5a2|R-50a5a2|H-50a5a2|0xc428604700              Data Response[18]:DataResponse Type:  0 Hash: 22a964a30064c63ce7a48bcd022e64ef70db1897073d46521c406e36b4d23ba0 FNode00 
/DataResponse/ {
    ts = time2sec($2);
    id = substr($12,1,10);
    print "dr",id;
    resps[id]++;

    if(!(id in first_resp)) {
        first_resp[id] = ts;
        delay[id] = ts - first_ask[id];
        inflight--;
        stat[ts]=inflight;
    }
    last_resp[id] = ts;
}

END {

    sum_of_asks = sum_of_array(asks);
    if(length(asks)==0) {
      average_ask = "no asks"
    } else {
      average_ask = sum_of_asks/length(asks);
    }
    print "EntryRequests:   total=", sum_of_asks, "unique=", length(asks), "min=", min_of_array(asks), "aver=", average_ask, "max=", max_of_array(asks);

    sum_of_resps = sum_of_array(resps);
    if(length(resps)==0) {
      average_resp = "no resp"
    } else {
      average_resp = sum_of_resps/length(resps);
    }
    print "EntryResponces:  total=", sum_of_resps, "unique=", length(resps), "min=", min_of_array(resps), "aver=", average_resp, "max=", max_of_array(resps);

    sum_of_delay = sum_of_array(delay);
    max_delay = max_of_array(delay);
    min_delay =  min_of_array(delay);
    
    
    if(length(delay)==0) {
      average_delay = "no delays"
    } else {
      average_delay = sum_of_delay/length(delay)
    }
    printf("Time for reply:  total= %s        min=%g aver=%g max=%g\\n", time2str(sum_of_delay), min_delay, average_delay, max_delay);

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
    
    
    
    sum_of_stat = sum_of_array(stat);
    if(length(stat)==0) {
      average_stat = "no stat"
    } else {
      average_stat = sum_of_delay/length(stat)
    }
    print "unique requests inflight average", average_stat, "max=", max_of_array(stat);
    
    min_inflight = min_index_of_array(stat); # time first ask 
    max_inflight = max_index_of_array(stat); # time of last responce
    sample_period = (max_inflight-min_inflight)/10 +.000000001;
    
    print "Start of trace", time2str(min_inflight), "duration", time2str(max_inflight-min_inflight), "sample period", sample_period;
    for(i in stat) {
        sample_idx = int((i-min_inflight)/sample_period);
        histogram2[sample_idx]+=stat[i];
        count[sample_idx]++;
    }
    
    PROCINFO["sorted_in"] ="@ind_num_asc";
    print "histogram of inflight requests in", sample_period, "second buckets"
    print "time          count";
    for(i in histogram2){
       printf("%s %d\\n", time2str(i * sample_period),histogram2[i]/count[i]);
    }
    print "---------------------"

    
        #for(i in stat) {print i, stats[i];}

    
}
EOF
################################
# End of AWK Scripts           #
################################

 
(grep -hE "Send.*MissingData" $1_networkoutputs.txt; grep -hE "enqueue.*DataResponse" $1_networkinputs.txt) | sort -n | awk "$scriptVariable"

