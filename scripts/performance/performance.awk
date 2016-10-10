

BEGIN{
	FS="[ [/]+"
}

{
  statusfield=0
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

		sumFct += fct-lastfct
		sumEc += ec-lastec
		sumE += e-laste
		
		lastfct = fct
		lastec = ec
		laste = e
		
		ts[rec]=time
		totalTps[rec]=(sumFct+sumEc+sumE)/($2-startTime)
		thisTps[rec]=(diffFct+diffEc+diffE)/(time-lastTime)
		rec++
		lastTime = time
	}
}

{ 
  node  = $(3+statusfield) 
  if (length(node) > 5 && substr(node,1,5)=="FNode") {
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
	block 	= $(8 +statusfield)
	dropped = $(6 +statusfield)
	delay	= $(7 +statusfield)
	fct	= $(25+statusfield)	
	ec	= $(26+statusfield)
	e 	= $(27+statusfield)		
}





END {	
	samples = 200
	scale = samples/rec*.99999 # Make sure we get at least our sample count.
	j = 0	
	for(i=0;i<rec;i++){
				
		here = int(j)
		j+=scale
		sCnt[here]      +=1
		sTs[here]       +=ts[i]
		sTotalTps[here] +=totalTps[i]
		sThisTps[here]  +=thisTps[i]
		
	}
	
	for(i=0;i<samples;i++){
		print int(sTs[i]/sCnt[i]) "\t" sTotalTps[i]/sCnt[i] "\t" sThisTps[i]/sCnt[i]
	}
}


