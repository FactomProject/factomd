# Input:  %CPU  %Memory (Each line represents 10 seconds real time)
#
# Output: Real Time (days)	Time (Days)	%CPU	%MEM
#

#Get all our data
{ 
	if ($1 != 0 && $2 != 0) {
		CPU[cnt]=$1; MEM[cnt]=$2; cnt++
	}
}


#Process

END {
	scale 	 = 500 / cnt
	realdays = 10/60/60/24
	simdays  = realdays / scale

	for (i=0;i<cnt;i++){
		oldptr = ptr		# Remember the old pointer
		ptr    = int(i*scale)	# Pointer into output data
		sumCPU += CPU[i]
		sumMEM += MEM[i]
		sumCnt++
		# print "pointers" oldptr " " ptr		
		if (oldptr != ptr || i+1 == cnt) {
			realtime += sumCnt*realdays
			simtime = sumCnt*simdays
			cpu = sumCPU/sumCnt
			mem = sumMEM/sumCnt
			print realtime "\t" simtime "\t" cpu "\t" mem
			realtime = simtime = 0
		}
	}
} 
