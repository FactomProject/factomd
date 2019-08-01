package globals

import (
	"io"
	"sync"
	"time"
)

var FnodeNames map[string]string = make(map[string]string) /// use by MessageTrace debug code
var Params FactomParams
var StartTime time.Time
var LastDebugLogRegEx string      // used to detect if FactomParams.DebugLogRegEx was changed by the control panel
var InputChan = make(chan string) // Get commands here

type FactomParams struct {
	AckbalanceHash           bool
	EnableNet                bool
	WaitEntries              bool
	ListenTo                 int
	Cnt                      int
	Net                      string
	Fnet                     string
	DropRate                 int
	Journal                  string
	Journaling               bool
	Follower                 bool
	Leader                   bool
	Db                       string
	CloneDB                  string
	PortOverride             int
	Peers                    string
	NetworkName              string
	NetworkPortOverride      int
	ControlPanelPortOverride int
	EnableLiveFeedAPI        bool
	EventReceiverProtocol    string
	EventReceiverAddress     string
	EventReceiverPort        int
	EventFormat              string
	LogPort                  string
	BlkTime                  int
	FaultTimeout             int
	RoundTimeout             int
	RuntimeLog               bool
	Exclusive                bool
	ExclusiveIn              bool
	P2PIncoming              int
	P2POutgoing              int
	Prefix                   string
	Rotate                   bool
	TimeOffset               int
	KeepMismatch             bool
	StartDelay               int64
	Deadline                 int
	CustomNet                []byte
	CustomNetName            string
	RpcUser                  string
	RpcPassword              string
	FactomdTLS               bool
	FactomdLocations         string
	MemProfileRate           int
	Fast                     bool
	FastLocation             string
	FastSaveRate             int
	Loglvl                   string
	Logjson                  bool
	Svm                      bool
	PluginPath               string
	TorManage                bool
	TorUpload                bool
	Sim_Stdin                bool
	ExposeProfiling          bool
	UseLogstash              bool
	LogstashURL              string
	Sync2                    int
	DebugConsole             string
	StdoutLog                string
	StderrLog                string
	DebugLogRegEx            string
	ConfigPath               string
	CheckChainHeads          bool // Run checkchain heads on boot
	FixChainHeads            bool // Only matters if CheckChainHeads == true
	ControlPanelSetting      string
	WriteProcessedDBStates   bool // Write processed DBStates to debug file
	NodeName                 string
	FactomHome               string
	FullHashesLog            bool // Log all unique full hashes
	DebugLogLocation         string
	ReparseAnchorChains      bool
}

/****************************************************************
	DEBUG logging to keep full hash. Turned on from command line
 ****************************************************************/
var HashMutex sync.Mutex
var Hashlog io.Writer
var Hashes map[[32]byte]bool
var HashesInOrder [10000]*[32]byte
var HashNext int
