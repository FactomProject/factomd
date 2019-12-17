#!/usr/bin/env bash

# processlist.sh <node> <height> <vms>

if [ "$#" -gt 1  ]; then
    echo "bootAnalysis.sh <node> "
fi
fnode=$1
shift
################################
# AWK scripts                  #
################################
read -d '' scriptVariable << 'EOF'

func time2sec(t) {    
    x = split(t,ary,":");
    if(x!=3) {
        printf("time2sec(%s) bad split got %d fields.\\n",t, x)
        printf("Line:  %s:%d\\n", FILENAME,FNR); 
        print FNR,"<"$0">"; 
        exit;
    }
    sec = (ary[1]*60+ary[2])*60+ary[3];
    #printf("time2sec(%s) %02d:%02d:%02d= %d\\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
    return sec;
}

func timeDiff(t1,t2){

    tDiff = time2sec(t1)-time2sec(t2);
    if(tDiff < 0) {
        tDiff = tDiff+24*60*60;
    }
    return sprintf("%d seconds %2d:%02d:%02d.%03d H:M:S", tDiff, (tDiff/(60*60)),(tDiff/60)%60,tDiff%60,(tDiff - int(tDiff))*1000 );
}

BEGIN {
    saveStateBlockTime = "unset";
    topOfDataBaseTime = "unset";
    firstBlockFromNetTime = "unset";
    lastBlockFromNetTime = "unset";
    firstMissingDataTime = "unset";
    lastMissingDataTime = "unset";
    firstBuiltBlockTime = "unset";
    lastBuiltBlockTime = "unset";
    firstExecBlockTime = "unset";
    lastExecBlockTime = "unset";
}

func checkunset(a,b,c,d,e,f){
    return a!="unset"&&b!="unset"&&c!="unset"&&d!="unset"&&e!="unset"&&f!="unset";
}


    {if($2=="") next; time2sec($2);}

#    31959 11:45:58.548  207001-:-0 FollowerExecute[-1]                    M-2e4c4f|R-5485cf|H-2e4c4f|0xc002a33980                    DBState[20]:DBState: dbht:207001 [size:      40,420] dblock 27f52e admin 674a18 fb ac263b ec bcb1b0 hash 2e4c4f ts:2019-08-24 08:32:00 InDB true IsLast false Sigs 29 RandomPeer 
/InDB true/ && saveStateBlockTime =="unset" {
     saveStateBlock=substr($7,6);
     saveStateBlockTime=$2;
}
/IsLast true/{
     topOfDataBase=substr($7,6);
     topOfDataBaseTime=$2;
}
/InDB false/ && firstBlockFromNetTime == "unset" {
     firstBlockFromNet=substr($7,6);
     firstBlockFromNetTime=$2;
}
/InDB false/ {
     lastBlockFromNet=substr($7,6);
     lastBlockFromNetTime=$2;
}

#   576450 11:57:03.947  208734-:-0 Enqueue                                M-a0b8d3|R-a0b8d3|H-a0b8d3|0xc00d946820               Missing Data[17]:MissingData: [1d2a961954] RandomPeer 
/MissingData:/ && firstMissingDataTime == "unset" {
    firstMissingDataTime = $2;
}

/MissingData:/ {
    lastMissingDataTime = $2;
}

#  2584743 12:08:00.493  208748-:-10 AddDBState(isNew true, directoryBlock 208748 50cc8b8a, adminBlock 74c668f5, factoidBlock 69e1841c, entryCreditBlock 467C9A26, eBlocks 5, entries 0) 
/95575-:-10 AddDBState.isNew true/ && firstBuiltBlockTime=="unset" {
    firstBuiltBlock = $7;
    firstBuiltBlockTime = $2;
}
/-:-10 AddDBState.isNew true/{
    lastBuiltBlock = $7;
    lastBuiltBlockTime = $2;
}

# 16444463 18:32:00.561  217796-:-2 done 217796/17/3                       M-3bc1e0|R-715edb|H-3bc1e0|0xc001cb9e00                        EOM[ 0]:   EOM-DBh/VMh/h 217796/17/-- minute  2 FF  0 --Leader[e3eded] hash[3bc1e0] ts 1573173120000 2019-11-07 18:32:00   
/EOM/ && firstExecBlockTime=="unset" {
    firstExecBlock = $3;
    firstExecBlockTime = $2;
}
/EOM/{
    lastExecBlock = $3;
    lastExecBlockTime = $2;
}



/enqueue.* Ack/ && firstNetworkHeight=="" {
    firstNetworkHeight = int($10)
}

/enqueue.* Ack/ {
    networkHeight = int($10)
}

END {


    print "saveStateBlockTime", saveStateBlockTime;
    print "topOfDataBaseTime", topOfDataBaseTime;
    print "firstBlockFromNetTime", firstBlockFromNetTime;
    print "lastBlockFromNetTime", lastBlockFromNetTime;
    print "firstMissingDataTime", firstMissingDataTime;
    print "lastMissingDataTime", lastMissingDataTime;
    print "firstExecBlockTime", firstExecBlockTime;
    print "lastExecBlockTime", lastExecBlockTime;
    print "firstBuiltBlockTime", firstBuiltBlockTime;
    print "lastBuiltBlockTime", lastBuiltBlockTime;


    if(checkunset(saveStateBlockTime)) {
    print "LoadFromDisk";
       print "saveStateBlock   ","DBHT:", saveStateBlock, saveStateBlockTime;
        if(checkunset(topOfDataBaseTime)) {
       print "topOfDataBase    ","DBHT:", topOfDataBase, topOfDataBaseTime, timeDiff(topOfDataBaseTime, saveStateBlockTime);
       rate = (topOfDataBase - saveStateBlock)/timeDiff(topOfDataBaseTime, saveStateBlockTime)
       print "Rate = ", rate, "blocks per second or", 1/rate, " seconds per block";
    }
    }  else {
        print "Load from disk never started"
    }
    if(checkunset(firstBlockFromNetTime)) {
        print "FirstPassSync";
        print "firstBlockFromNet","DBHT:", firstBlockFromNet, firstBlockFromNetTime;
        if(checkunset(lastBlockFromNetTime)) {
            print "lastBlockFromNet ","DBHT:", lastBlockFromNet, "First Network Height", firstNetworkHeight, "Network Height", networkHeight, lastBlockFromNetTime, timeDiff(lastBlockFromNetTime, firstBlockFromNetTime);
        rate = (lastBlockFromNet - firstBlockFromNet)/timeDiff(lastBlockFromNetTime, firstBlockFromNetTime);
        print "Rate = ", rate, "blocks per second or", 1/rate, " seconds per block";
    }
    } else {
        print "First pass never started"
    }
    if(checkunset(firstMissingDataTime)) {
        print "SecondPassSync";
        print "firstMissingDataTime:", firstMissingDataTime;
        if(checkunset(lastMissingDataTime)) {
        print "lastMissingDataTime: ", lastMissingDataTime, timeDiff(lastMissingDataTime, firstMissingDataTime);
        rate = (lastBlockFromNet - firstBlockFromNet)/timeDiff(lastMissingDataTime, firstMissingDataTime);
        print "Rate = ", rate, "blocks per second or", 1/rate, " seconds per block";
    }
    } else {
        print "Second pass never started"
    }
    if(checkunset(firstExecBlockTime)) {
        print "FollowByMinutes(execute)";    
        print "firstExecBlock  ","DBHT:", firstExecBlock, firstExecBlockTime, "Network Height", networkHeight;
        if(checkunset(lastExecBlockTime)) {
            print "lastExecBlock   ","DBHT:", lastExecBlock, lastExecBlockTime, timeDiff(lastExecBlockTime, firstExecBlockTime);
            rate = (lastExecBlock - firstExecBlock)/timeDiff(lastExecBlockTime, firstExecBlockTime);
            print "Rate = ", rate, "blocks per second or", 1/rate, " seconds per block";
        }
    } else {
        print "Follow(execute) by minutes never started"
    }
    if(checkunset(firstBuiltBlockTime)) {
        print "FollowByMinutes";    
        print "firstBuiltBlock  ","DBHT:", firstBuiltBlock, firstBuiltBlockTime;
        if(checkunset(lastBuiltBlockTime)) {
        print "lastBuiltBlock   ","DBHT:", lastBuiltBlock, lastBuiltBlockTime, timeDiff(lastBuiltBlockTime, firstBuiltBlockTime);
        rate = (lastBuiltBlock - firstBuiltBlock)/timeDiff(lastBuiltBlockTime, firstBuiltBlockTime);
        print "Rate = ", rate, "blocks per second or", 1/rate, " seconds per block";
    } 
    } else {
        print "Follow(save) by minutes never started"
    }
}


EOF
################################
# End of AWK Scripts           #
################################
   

echo "bootAnalysis.sh" ${fnode:-fnode0}

(grep -e "done.* EOM" ${fnode:-fnode0}_process.txt; grep -E "enqueue.* Ack" ${fnode:-fnode0}_networkinputs.txt;grep "AddDBState" ${fnode:-fnode0}_dbstateprocess.txt; grep DBState: ${fnode:-fnode0}_executemsg.txt; grep MissingData: ${fnode:-fnode0}_networkoutputs.txt)  | tee x | awk "$scriptVariable"
