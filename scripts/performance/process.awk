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
	simscale = 600/60           # assume 60 second blocks
	if (blktime>0){
	    simscale = 600/blktime
	}
}


#Process

END {
	scale 	  = 500 / cnt
	realhours = 10/60/60  	   # Value of each data point in real time.
		
	for (i=0;i<cnt;i++){
		oldptr = ptr		# Remember the old pointer
		ptr    = int(i*scale)	# Pointer into output data
		sumCPU += CPU[i]
		sumMEM += MEM[i]
		sumCnt++
		# print "pointers" oldptr " " ptr		
		if (oldptr != ptr || i+1 == cnt) {
			realtime = i * realhours
			cpu = sumCPU/sumCnt
			mem = sumMEM/sumCnt
			outstr = sprintf(" %8.4f\t%8.4f\t%6.2f\t%6.2f",realtime,realtime,cpu,mem) 
			print outstr
			ot++
			sumCPU=0
			sumMEM=0
			sumCnt=0
		}
	}
	for (;ot<500;ot++){
		print outstr
	}

} 
