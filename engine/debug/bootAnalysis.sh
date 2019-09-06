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
    if(x!=3) {printf("time2sec(%s) bad split got %d fields. %s:%d", t , x ,FILENAME,FNR); print $0; exit;}
    sec = (ary[1]*60+ary[2])*60+ary[3];
    #printf("time2sec(%s) %02d:%02d:%02d= %d\\n",t, ary[1]+0, ary[2]+0,ary[3]+0,sec);
    return sec;
}


#    31959 11:45:58.548  207001-:-0 FollowerExecute[-1]                    M-2e4c4f|R-5485cf|H-2e4c4f|0xc002a33980                    DBState[20]:DBState: dbht:207001 [size:      40,420] dblock 27f52e admin 674a18 fb ac263b ec bcb1b0 hash 2e4c4f ts:2019-08-24 08:32:00 InDB true IsLast false Sigs 29 RandomPeer 
/InDB true/ && saveStateBlock=="" {
     saveStateBlock=substr($7,6);
     saveStateBlockTime=$2;
}
/IsLast true/{
     topOfDataBase=substr($7,6);
     topOfDataBaseTime=$2;
}
/InDB false/ && firstBlockFromNet == "" {
     firstBlockFromNet=substr($7,6);
     firstBlockFromNetTime=$2;
}
/InDB false/ {
     lastBlockFromNet=substr($7,6);
     lastBlockFromNetTime=$2;
}

#   576450 11:57:03.947  208734-:-0 Enqueue                                M-a0b8d3|R-a0b8d3|H-a0b8d3|0xc00d946820               Missing Data[17]:MissingData: [1d2a961954] RandomPeer 
/MissingData:/ && firstMissingDataTime == "" {
    firstMissingDataTime = $2;
}

/MissingData:/ {
    lastMissingDataTime = $2;
}

#  2584743 12:08:00.493  208748-:-10 AddDBState(isNew true, directoryBlock 208748 50cc8b8a, adminBlock 74c668f5, factoidBlock 69e1841c, entryCreditBlock 467C9A26, eBlocks 5, entries 0) 
/-:-10 AddDBState.isNew true/ && firstBuiltBlock=="" {
    firstBuiltBlock = $7;
    firstBuiltBlockTime = $2;
}
/-:-10 AddDBState.isNew true/{
    lastBuiltBlock = $7;
    lastBuiltBlockTime = $2;
}


END {

    print "LoadFromDisk";
    print "saveStateBlock   ","DBHT:", saveStateBlock, saveStateBlockTime;
    print "topOfDataBase    ","DBHT:", topOfDataBase, topOfDataBaseTime, time2sec(topOfDataBaseTime)-time2sec(saveStateBlockTime),"seconds";
    print "FirstPassSync";
    print "firstBlockFromNet","DBHT:", firstBlockFromNet, firstBlockFromNetTime;
    print "lastBlockFromNet ","DBHT:", lastBlockFromNet, lastBlockFromNetTime, time2sec(lastBlockFromNetTime)-time2sec(firstBlockFromNetTime),"seconds";;
    print "SecondPassSync";
    print "firstMissingDataTime:", firstMissingDataTime;
    print "lastMissingDataTime: ", lastMissingDataTime, time2sec(lastMissingDataTime)-time2sec(firstMissingDataTime),"seconds";;
    print "FollowByMinutes";    
    print "firstBuiltBlock  ","DBHT:", firstBuiltBlock, firstBuiltBlockTime;
    print "lastBuiltBlock   ","DBHT:", lastBuiltBlock, lastBuiltBlockTime, time2sec(lastBuiltBlockTime)-time2sec(firstBuiltBlockTime),"seconds";;
     
}


EOF
################################
# End of AWK Scripts           #
################################
   

echo "bootAnalysis.sh" ${fnode:-fnode0}

(grep "AddDBState" ${fnode:-fnode0}_dbstateprocess.txt; grep DBState: ${fnode:-fnode0}_executemsg.txt; grep MissingData: ${fnode:-fnode0}_networkoutputs.txt)  | tee x | awk "$scriptVariable"
