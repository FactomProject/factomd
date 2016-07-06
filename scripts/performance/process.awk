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
	realdays = 10/60/60/24  	   # Value of each data point in real time.
	simdays  = realdays * 10           # We assume 1 minute blocks;  maybe not a good assumption
		
	for (i=0;i<cnt;i++){
		oldptr = ptr		# Remember the old pointer
		ptr    = int(i*scale)	# Pointer into output data
		sumCPU += CPU[i]
		sumMEM += MEM[i]
		sumCnt++
		# print "pointers" oldptr " " ptr		
		if (oldptr != ptr || i+1 == cnt) {
			realtime = i * realdays
			simtime = i * simdays
			cpu = sumCPU/sumCnt
			mem = sumMEM/sumCnt
			outstr = sprintf(" %8.4f\t%8.4f\t%6.2f\t%6.2f",realtime,simtime,cpu,mem) 
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
