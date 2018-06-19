package globals

import "time"

var FnodeNames map[string]string = make(map[string]string) /// use by MessageTrace debug code
var Params FactomParams
var StartTime time.Time
var LastDebugLogRegEx string // used to detect if FactomParams.DebugLogRegEx was changed by the control panel

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
	LogPort                  string
	BlkTime                  int
	FaultTimeout             int
	RuntimeLog               bool
	Exclusive                bool
	ExclusiveIn              bool
	Prefix                   string
	Rotate                   bool
	TimeOffset               int
	KeepMismatch             bool
	StartDelay               int64
	Deadline                 int
	CustomNet                []byte
	RpcUser                  string
	RpcPassword              string
	FactomdTLS               bool
	FactomdLocations         string
	MemProfileRate           int
	Fast                     bool
	FastLocation             string
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

}
