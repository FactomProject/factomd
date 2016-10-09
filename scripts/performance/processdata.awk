

BEGIN{
	FS="[ [/]+"
}

$3 == "f" || $3 == "w" || $3 == "fw" {
	statusfield=1
}

{
	print(	"blocks\t" 	,$(8 +statusfield), 
		"\tDropped\t"	,$(6 +statusfield),
		"\tDelay\t"	,$(7 +statusfield))
	
	print(	"FCT\t"		,$(25+statusfield),
		"\tEC\t"	,$(26+statusfield),
		"\tE\t"		,$(27+statusfield))
}


