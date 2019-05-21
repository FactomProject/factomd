#!/usr/bin/env bash

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

func initialize(whichArray) {

    for (i in age[whichArray]) {
        larg[whichArray]=age[whichArray][i];
        smal[whichArray]=age[whichArray][i];
        totaladd[whichArray]=0;
        break;
    }
    
    for (j in age[whichArray]) {
        if (age[whichArray][j] > larg[whichArray]) {
            larg[whichArray]=age[whichArray][j];
        }
        if (age[whichArray][j] < smal[whichArray]) {
            smal[whichArray]=age[whichArray][j];
        }
    
        totaladd[whichArray]+=age[whichArray][j]; 
    }
    
    for(i=0;i<bucks;i++){
        histogram[whichArray][i]=0;
    }
   
    for (i in age[whichArray]) {
        inde=int((age[whichArray][i]-smal["combined"])/((larg["combined"]-smal["combined"])/(bucks-1)));
        histogram[whichArray][inde]++;
    }
}

func PrintHisto(whichArray) {

    PROCINFO["sorted_in"] ="@val_num_asc";
    for (x=0;x<bucks;x++) {
        print value[whichArray][x], "     ", histogram[whichArray][x];
    }
}


BEGIN {count=0; bucks=100; typeOfLog[0]=0; holding[0]= "foo"; delete holding[0];
    
}

  { printf("%30s:%d                  \\r", FILENAME, FNR) > "/dev/stderr"; }
  { printf("%5d %s %d \\n", FNR, $3, length(holding)); }

$4~/add/ {
    count++;
    hash = substr($0,index($0,"R-"),8);   
#    print "add", hash
    if(hash in everything) {
      print "Duplicate", hash, FILENAME,FNR;
    } else {
      everything[hash]=0;
    }
    
    timestamp[hash]=$2
    everything[hash]++;
    holding[hash] = $0;
}

$4~/delete/ {
    hash = substr($0,index($0,"R-"),8);   
#    print "del", hash
    delete holding[hash];
    age["combined"][hash]=time2sec($2) - time2sec(timestamp[hash])
    
    rest= substr($0,index($0,"M-")+39); 
    match(rest,/[^ ][^\]]+\]/); 
    tag = substr(rest,RSTART,RLENGTH); 
    gsub(/\\[ /, "[", tag);
    gsub(/ /, "-", tag);
    
    age[tag][hash]=time2sec($2) - time2sec(timestamp[hash]);

}


END {
    print "count   =", count;
    print "current =", length(holding)
    
    initialize("combined");
    for (i in age) {
        initialize(i);
    }
    
    for(i=0;i<bucks;i++){
           value[i]=smal["combined"] + i*((larg["combined"]-smal["combined"])/(bucks-1));
    }
   
    PROCINFO["sorted_in"] ="@val_num_asc";
    printf("%19s ","index")
    for(i in histogram) {
        printf("%19s ",i);
    }
    print "";

    PROCINFO["sorted_in"] ="@ind_num_asc";
    for (k in histogram["combined"]) {
        printf("%19.3f:", value[k])
        PROCINFO["sorted_in"] ="@val_num_asc";
        for(j in histogram){
            printf("%19d ",histogram[j][k]);
        }
        print "";
    }
    print "";
    
    printf("%13s", ""); 
    for(i in histogram) {
        printf("%19s ",i);
    }
    print "";
    
    PROCINFO["sorted_in"] ="@val_str_asc";
    printf("Average Time:")
    for (k in histogram) {
        printf("%19.4f ", totaladd[k]/length(age[k]))
    }
    print "";
    
    printf("Largest Time:")
    for (k in histogram) {
        printf("%19.4f ", larg[k])
    }
    print "";
    
    printf("Smallest Time:")
    for (k in histogram) {
        printf("%19.4f ", smal[k])
    }
    print "";
    
    print "Holding", length(holding)
    for(i in holding) {
      print holding[i]
    }
}


EOF
################################
# End of AWK Scripts           #
################################
gawk  "$scriptVariable" $@ | less -R
