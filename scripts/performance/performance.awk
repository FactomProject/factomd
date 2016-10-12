

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
	block 	= $(8 +statusfield)
	dropped = $(6 +statusfield)
	delay	= $(7 +statusfield)
	fct	= $(25+statusfield)	
	ec	= $(26+statusfield)
	e 	= $(27+statusfield)		
}

isFnode {
	totaltrans	+= $(25+statusfield)	
	totaltrans	+= $(26+statusfield)
	totaltrans	+= $(27+statusfield)		
}



END {	
	rec
	samples = 500
	if (rec < samples) samples = rec
	scale = samples/rec*1.001 # Make sure we get at least our sample count.
	j = 0	
	for(i=0;i<rec;i++){
				
		here = int(j)
		j+=scale
		sCnt[here]      +=1
		sTs[here]       +=ts[i]
		sTotalTps[here] +=totalTps[i]
		sThisTps[here]  +=thisTps[i]
		sTotalAllTps[here] +=totalAllTps[i]
		sThisAllTps[here] +=thisAllTps[i]		
	}
	
	for(i=0;i<samples;i++){
		if (sCnt[i]==0) sCnt[i]=1
		printf("%8d\t%8.3f\t%8.3f\t%8.3f\t%8.3f\t\n", int(sTs[i]/sCnt[i]), sTotalTps[i]/sCnt[i], sThisTps[i]/sCnt[i], sTotalAllTps[i]/sCnt[i], sThisAllTps[i]/sCnt[i])
	}
	for(j=samples;j<500;j++){
		i = samples-1
		printf("%8d\t%8.3f\t%8.3f\t%8.3f\t%8.3f\t\n", int(sTs[i]/sCnt[i]), sTotalTps[i]/sCnt[i], sThisTps[i]/sCnt[i], sTotalAllTps[i]/sCnt[i], sThisAllTps[i]/sCnt[i])
	}
}


