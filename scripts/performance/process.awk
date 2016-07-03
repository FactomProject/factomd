# Input:  %CPU  %Memory (Each line represents 10 seconds real time)
#
# Output: Real Time (days)	Time (Days)	%CPU	%MEM
#

#Get all our data
{ 
	if ($1+0 != 0 && $2+0 != 0) {
		CPU[cnt]=$1; 
		MEM[cnt]=$2; 
		cnt++
	}
}


#Process

END {
	scale 	 = 500 / cnt
	realdays = 1/6/60/24
	simdays  = 10/60/24
		
	for (i=0;i<cnt;i++){
		oldptr = ptr		# Remember the old pointer
		ptr    = int(i*scale)	# Pointer into output data
		sumCPU += CPU[i]
		sumMEM += MEM[i]
		sumCnt++
		periodCnt++
		# print "pointers" oldptr " " ptr		
		if (oldptr != ptr || i+1 == cnt) {
			realtime += (i-(periodCnt/2) ) * realdays
			simtime += (i-(periodCnt/2) ) * simdays
			cpu = sumCPU/periodCnt
			mem = sumMEM/periodCnt
			outstr = realtime "\t" simtime "\t" cpu "\t" mem "\t" sumCPU "\t" periodCnt
			print outstr
			ot++
			sumCPU=0
			sumMem=0
			periodCnt=0
		}
	}
	for (;ot<500;ot++){
		print outstr
	}

} 
