

BEGIN{
	FS="[ [/]+"
}

{
  statusfield=0
  isFnode = 0
}

$3 == "f" || $3 == "w" || $3 == "fw" {
	statusfield=1
}

/Time:/ {
 	leadercnt=0
	if (startTime == 0) {
		startTime = $2
		lastTime = 0
	}
	time = $2-startTime
	diff = time-lastTime
	if (diff > 20) {
		diffFct = fct-lastfct
		diffEc = ec-lastec
		diffE = e-laste
		diffAll = totaltrans-lastall

		sumFct += fct-lastfct
		sumEc += ec-lastec
		sumE += e-laste
		sumTrans += totaltrans - lastall

		lastfct = fct
		lastec = ec
		laste = e
		lastall = totaltrans
		
		ts[rec]=time

		totalTps[rec]=(sumFct+sumEc+sumE)/time
		thisTps[rec]=(diffFct+diffEc+diffE)/(time-lastTime)

		totalAllTps[rec]=(sumTrans)/time
		thisAllTps[rec]=(diffAll)/(time-lastTime)

		rec++
		lastTime = time
	}
}

{ 
  node  = $(3+statusfield) 
  if (length(node) > 5 && substr(node,1,5)=="FNode") {
     isFnode = 1
     nodeNum = 	substr(node,6)
     if (maxNum < nodeNum) {
        maxNum = nodeNum
     }
     status = substr($(5+statusfield),1,1)
     if (status == "L") {
	leadercnt++
     }
  }
}

node =="FNode0" {
	totaltrans = 0
	block 	= $(10 +statusfield)
	dropped = $(8 +statusfield)
	delay	= $(9 +statusfield)
	fct	= $(27+statusfield)
	ec	= $(28+statusfield)
	e 	= $(29+statusfield)
}

isFnode {
	totaltrans	+= $(27+statusfield)
	totaltrans	+= $(28+statusfield)
	totaltrans	+= $(29+statusfield)
}



END {	
	rec
	samples = 500
	if (rec < samples) samples = rec
	scale = samples/rec*1.001 # Make sure we get at least our sample count.
	j = 0
	maxtps = thisTps[0]
	mintps = maxtps
	allmaxtps = thisAllTps[0]
	allmintps = allmaxtps


	for(i=0;i<rec;i++){
	    if (maxtps < thisTps[i]) {
	    	maxtps = thisTps[i]
	    }
	    if (mintps > thisTps[i]) {
	        mintps = thisTps[i]
	    }

        if (allmaxtps < thisAllTps[i]) {
	    	allmaxtps = thisAllTps[i]
	    }
	    if (allmintps > thisAllTps[i]) {
	        allmintps = thisAllTps[i]
	    }

        oldhere = here
		here = int(j)
		if (here != oldhere) {
		    maxtps = thisTps[i]
		    mintps = maxtps
		    allmaxtps = thisAllTps[i]
		    allmintps = allmaxtps
		}

		j+=scale
		sCnt[here]      +=1
		sTs[here]       +=ts[i]
		sTotalTps[here] +=totalTps[i]
		sTotalAllTps[here] +=totalAllTps[i]
		if (here%2 == 0) {
		    sThisTps[here]=maxtps
		    sThisAllTps[here]=allmaxtps
		}else{
		    sThisTps[here]=mintps
		    sThisAllTps[here]=allmintps
		}
	}
	
	for(i=0;i<samples;i++){
		if (sCnt[i]==0) sCnt[i]=1
		printf("%8d\t%8.3f\t%8.3f\t%8.3f\t%8.3f\t\n", int(sTs[i]/sCnt[i]), sTotalTps[i]/sCnt[i], sThisTps[i], sTotalAllTps[i]/sCnt[i], sThisAllTps[i])
	}
	for(j=samples;j<500;j++){
		i = samples-1
		printf("%8d\t%8.3f\t%8.3f\t%8.3f\t%8.3f\t\n", int(sTs[i]/sCnt[i]), sTotalTps[i]/sCnt[i], sThisTps[i], sTotalAllTps[i]/sCnt[i], sThisAllTps[i])
	}
}


