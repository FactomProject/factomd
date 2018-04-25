# find . -name vendor -prune -o -name .git -prune -o -name \*.go | grep -v vendor | grep -v git | xargs awk -i inplace -f ~/fixups.awk
# find . -name vendor -prune -o -name .git -prune -o -name \*.go | grep -v vendor | grep -v git | xargs goimports -w
# find . -name vendor -prune -o -name .git -prune -o -name \*.go | grep -v vendor | grep -v git | xargs gofmt -w
# git reset --hard


# case preserving fixup
func fix(bad,good) {
 rval =0
    rval += gsub(bad,good)
    bad = toupper(substr(bad,1,1)) substr(bad,2)
    good = toupper(substr(good,1,1)) substr(good,2)
    rval += gsub(bad,good)
    return rval
}


  {  
againsttt
 fixes += fix("accending" , "ascending" )
 fixes += fix("accomadate" , "accommodate" )
 fixes += fix("accomadation" , "accommodation" )
 fixes += fix("acknowledgemets","acknowledgements");
 fixes += fix("agains ","against");
 fixes += fix("aksed","asked");
 fixes += fix("appication" , "application" )
 fixes += fix("applicaiton" , "application" )
 fixes += fix("asyncronous" , "asynchronous" )
 fixes += fix("athoritites","authorities");
 fixes += fix("authorites" , "authorities" )
 fixes += fix("authoritites","authorities");
 fixes += fix("boundries","boundaries");
 fixes += fix("boundry" , "boundary" )
 fixes += fix("broadcasted","broadcast");
 fixes += fix("caculates","calculates");
 fixes += fix("caluclate" , "calculate" )
 fixes += fix("caluclate","calculate");
 fixes += fix(" cant "," can't ");
 fixes += fix("catagories" , "categories" )
 fixes += fix("chainges","changes");
 fixes += fix("comming" , "coming" )
 fixes += fix("concatinated" , "contaminated" )
 fixes += fix("concensus" , "consensus" )
 fixes += fix("controler","controller");
 fixes += fix("cooresponding" , "corresponding" )
 fixes += fix("corresponsing" , "corresponding" )
 fixes += fix("detatched" , "detached" )
 fixes += fix("detatched","detached")
 fixes += fix("discoverd" , "discovered" )
 fixes += fix("dont'","do not");
 fixes += fix("exmaple" , "example" )
 fixes += fix("fufill" , "fulfill" )
 fixes += fix("functionaility","functionality");
 fixes += fix("guage" , "gauge" )
 fixes += fix("hieght" , "height" )
 fixes += fix("identitiy","identity");
 fixes += fix("in comming","incoming");
 fixes += fix("inital" , "initial" )
 fixes += fix("inital","initial");
 fixes += fix("initalize" , "initialize" )
 fixes += fix("initalizes" , "initializes" )
 fixes += fix("journalling","journaling");
 fixes += fix("markes" , "marks" )
 fixes += fix("marshaled","marshalled");
 fixes += fix("negotiationg" , "negotiating" )
 fixes += fix("netowrking" , "networking" )
 fixes += fix("netowrk" , "network" )
 fixes += fix("occurances" , "occurrences" )
 fixes += fix("occured" , "occurred" )
 fixes += fix("partical" , "particular" )
 fixes += fix("paticular" , "particular" )
 fixes += fix("preceeding" , "preceding" )
 fixes += fix("readible","readable");
 fixes += fix("recieved" , "received" )
 fixes += fix("recieve" , "receive" )
 fixes += fix("recieve","receive");
 fixes += fix("recieves" , "receives" )
 fixes += fix("recieving" , "receiving" )
 fixes += fix("relevent" , "relevant" )
 fixes += fix("responsiblity" , "responsibility" )
 fixes += fix("selectivly" , "selectively" )
 fixes += fix("seperated" , "separated" )
 fixes += fix("seperate" , "separate" )
 fixes += fix("signiture","signature");
 fixes += fix("simpy","simply");
 fixes += fix("specifiy" , "specify" )
 fixes += fix("syncronization" , "synchronization" )
 fixes += fix("targetted" , "targeted" )
 fixes += fix("transaciton" , "transactions" )
 fixes += fix("transctions" , "transitions" )
 fixes += fix("unkown" , "unknown" )
 fixes += fix("unnecesary" , "unnecessary" )
 fixes += fix("varaible" , "variable" )
 fixes += fix("werid","weird")
 fixes += fix("withing" , "within" )


# special case with mixed case to insure capitalization
 fixes += fix(" factom "," Factom ");


  }

 {comment =""}
# strip eol comment
/[/][/]/ {start = index($0,"//"); comment = substr($0,start); $0 = substr($0,1,start-1)}


# # Use debugMutexes instead of syncMutex
# FILENAME!~/(atomic)|(rateCalculator)/ { fixes += gsub(/sync.Mutex/,"atomic.DebugMutex");}
# /import \(/   {inimport=1}
# /import \(\)/ {inimport=0} # HANDLE FILES THAT DON'T IMPORT ANYTING
# $0~"github.com/FactomProject/factomd/util/atomic" {inimport=0} # don't duplicate the atomic import
# FILENAME~/(atomic)|(rateCalculator)/ && inimport {inimport=0} # don't import atomic in these files that use sync on purpose
# /^)$/ && inimport {inimport=0; 	print "\"github.com/FactomProject/factomd/util/atomic\"";}# add atomic to import, goimport will toss it if it's not needed
# # End Use debugMutexes instead of syncMutex
# 
# # EntryDBHeightComplete atomic.AtomicUint32
# # wsapiNode, ListenTo atomic.AtomicInt
# # DBFinished, OutputAllowed, NetStateOff, ControlPanelDataRequest atomic.AtomicBool
# 
# 
# # do the work for atomic stores
# func dostore(name) { # assumes EOL comments are stripped
#  sub("*" name, name);
#  match($0, / *= */);
#  rest = substr($0,RSTART+RLENGTH)
#  $0 = substr($0,1,RSTART-1) ".Store(" rest ")"
#  fixes++
# }
# 
# 
# 
# # atomic status access
# /Status +uint8/ 			    {fixes += gsub(/uint8/,"atomic.AtomicUint8")}
# /((oneID)|(newAuth)|(.*[iI]dent.*)|(.*Auth.*)|([^a-z][abei])|(auth)|(id))\.Status +=[^=]/ 		    {dostore()}
# /((oneID)|(newAuth)|(.*[iI]dent.*)|(.*Auth.*)|([^a-z][abei])|(auth)|(id))\.Status +((==)|[><)]|(<=)|(>=))/ {fixes += gsub(/Status /,"Status.Load() ")}
# /[ \t(]((oneID)|(newAuth)|(.*[iI]dent.*)|(.*Auth.*)|([^a-z][abei])|(auth)|(id))\.Status[)]/ {fixes += gsub(/Status[)]/,"Status.Load())")}
# /[(][abei].Status[)]/ {fixes += gsub(/Status[)]/,"Status.Load())")}
# /status :=.*\.Status/ {fixes += gsub(/Status$/,"Status.Load()")}
# # end atomic status access
# 
# 
# # EntryDBHeightComplete atomic.AtomicUint32
# 
# /!.*\.DBFinished/     	{fixes += gsub(/DBFinished/,"DBFinished.Load() ")}
# /DBFinished +bool/ 	{fixes += gsub(/bool/,"atomic.AtomicBool")}
# /DBFinished +=[^=]/ 	{dostore("DBFinished")}
# /DBFinished == true/    {fixes += gsub(/DBFinished == true/,"DBFinished.Load() ")}
# #
# /OutputAllowed +bool/ 	{fixes += gsub(/bool/,"atomic.AtomicBool")}
# /OutputAllowed +=[^=]/ 	{dostore("OutputAllowed")}
# /[^ \t] *OutputAllowed([^.]|$)/ {fixes += gsub(/OutputAllowed/,"OutputAllowed.Load() ")}
# 
# #
# /[^t]NetStateOff +bool/ 	{fixes += gsub(/bool/,"atomic.AtomicBool")}
# /[^t]NetStateOff +=[^=]/ 	{dostore()}
# /[^t\"]NetStateOff([^.]|$)/ {fixes += gsub(/\.NetStateOff/,".NetStateOff.Load() ")}
# 
# #
# /[^t]ControlPanelDataRequest +bool/ 	{fixes += gsub(/bool/,"atomic.AtomicBool")}
# /[^t]ControlPanelDataRequest +=[^=]/ 	{dostore("ControlPanelDataRequest")}
# /[^t\"]ControlPanelDataRequest(( [{])|( *[^. a])|$)/ {fixes += gsub(/\.ControlPanelDataRequest/,".ControlPanelDataRequest.Load() ")}
# #
# /[^t]wsapiNode +int/ 	{fixes += gsub(/int/,"atomic.AtomicInt")}
# /[^t]wsapiNode +=[^=]/ 	{dostore("wsapiNode")}
# /\[wsapiNode/ {fixes += gsub(/wsapiNode/,"wsapiNode.Load() ")}
# /wsapiNode \*int/ {fixes += gsub(/wsapiNode \*int/,"wsapiNode *atomic.AtomicInt")}
# /\*wsapiNode/ {fixes += gsub(/\*wsapiNode/,"wsapiNode.Load()")}
# 
# #
# /\[ListenTo/ {fixes += gsub(/ListenTo/,"ListenTo.Load()")}
# /[^t]ListenTo +=[^=]/ 	{dostore("ListenTo")}
# /[^t]ListenTo +int/ 	{fixes += gsub(/int/,"atomic.AtomicInt")}
# 
# /[<>] ListenTo/ {fixes += gsub(/ListenTo/,"ListenTo.Load()")}
# /ListenTo [<>]/ {fixes += gsub(/ListenTo/,"ListenTo.Load()")}
# /[^&]ListenTo[,)]/ {fixes += gsub(/ListenTo/,"ListenTo.Load()")}
# /ListenTo\+\+/ {fixes += gsub(/ListenTo\+\+/,"ListenTo.Store(ListenTo.Load()+1)")}
# 
# #exclude the ISR routine
# /func.*((InstantaneousStatusReport)|(SimControl))/ {inISR=1;}
# /^}/ {inISR=0}
# 
# /listenTo \*int/        {if(inISR==0){fixes += gsub(/listenTo \*int/,"listenTo *atomic.AtomicInt")}}
# /[^.]listenTo +=[^=]/ 	{if(inISR==0){dostore("listenTo")}}
# /\*listenTo[^P.]/       {if(inISR==0){fixes += gsub(/\*listenTo/,"listenTo.Load()")}}
# /\[listenTo[^.]/        {if(inISR==0){fixes += gsub(/listenTo/,"listenTo.Load()")}}
# 

#
 { printf("%s%s\n", $0, comment); 
   comment =""
   if(fixes!=lfixes){ 
	printf("\r%3d", fixes) > "/dev/stderr";
   } 
 }

