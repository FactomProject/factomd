

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
	block 	= $(8 +statusfield)
	dropped = $(6 +statusfield)
	delay	= $(7 +statusfield)
	fct	= $(25+statusfield)	
	ec	= $(26+statusfield)
	e 	= $(27+statusfield)
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


