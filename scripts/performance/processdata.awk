

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
	block 	= $(10 +statusfield)
	dropped = $(8 +statusfield)
	delay	= $(9 +statusfield)
	fct	= $(27+statusfield)
	ec	= $(28+statusfield)
	e 	= $(29+statusfield)
}





END {	
	print(  "nodes\t"       ,maxNum+1,
		"\tleaders\t" 	,leadercnt)

	print(	"blocks\t" 	,block,
		"\tDropped\t"	,dropped,
		"\tDelay\t"	,delay)
	
	print(	"FCT\t"		,fct,
		"\tEC\t"	,ec,
		"\tE\t"		,e)
}


