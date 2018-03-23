package engine

import (
	"flag"
	"os"

	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
)

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
}

func (f *FactomParams) Init() { // maybe used by test code
	f.AckbalanceHash = true
	f.EnableNet = true
	f.WaitEntries = false
	f.ListenTo = 0
	f.Cnt = 1
	f.Net = "tree"
	f.Fnet = ""
	f.DropRate = 0
	f.Journal = ""
	f.Journaling = false
	f.Follower = false
	f.Leader = true
	f.Db = ""
	f.CloneDB = ""
	f.PortOverride = 0
	f.Peers = ""
	f.NetworkName = ""
	f.NetworkPortOverride = 0
	f.ControlPanelPortOverride = 0
	f.LogPort = "6060"
	f.BlkTime = 0
	f.FaultTimeout = 99999 //TODO: REMOVE  Old Fault Mechanism
	f.RuntimeLog = false
	f.Exclusive = false
	f.Prefix = ""
	f.Rotate = false
	f.TimeOffset = 0
	f.KeepMismatch = false
	f.StartDelay = 10
	f.Deadline = 1000
	f.CustomNet = primitives.Sha([]byte("")).Bytes()[:4]
	f.RpcUser = ""
	f.RpcPassword = ""
	f.FactomdTLS = false
	f.FactomdLocations = ""
	f.MemProfileRate = 512 * 1024
	f.Fast = true
	f.FastLocation = ""
	f.Loglvl = "none"
	f.Logjson = false
	f.Svm = false
	f.PluginPath = ""
	f.TorManage = false
	f.TorUpload = false
	f.Sim_Stdin = true
	f.ExposeProfiling = false
	f.Sync2 = -1
	f.DebugConsole = "foobar" //TODO: pretty sure this value is overridden by the default in the flag -- clay
	f.StdoutLog = "out.txt"
	f.StderrLog = "err.txt"
}

func ParseCmdLine(args []string) *FactomParams {
	p := new(FactomParams)

	ackBalanceHashPtr := flag.Bool("balancehash", true, "If false, then don't pass around balance hashes")
	enablenetPtr := flag.Bool("enablenet", true, "Enable or disable networking")
	waitEntriesPtr := flag.Bool("waitentries", false, "Wait for Entries to be validated prior to execution of messages")
	listenToPtr := flag.Int("node", 0, "Node Number the simulator will set as the focus")
	cntPtr := flag.Int("count", 1, "The number of nodes to generate")
	netPtr := flag.String("net", "tree", "The default algorithm to build the network connections")
	fnetPtr := flag.String("fnet", "", "Read the given file to build the network connections")
	dropPtr := flag.Int("drop", 0, "Number of messages to drop out of every thousand")
	journalPtr := flag.String("journal", "", "Rerun a Journal of messages")
	journalingPtr := flag.Bool("journaling", false, "Write a journal of all messages received. Default is off.")
	followerPtr := flag.Bool("follower", false, "If true, force node to be a follower.  Only used when replaying a journal.")
	leaderPtr := flag.Bool("leader", true, "If true, force node to be a leader.  Only used when replaying a journal.")
	dbPtr := flag.String("db", "", "Override the Database in the Config file and use this Database implementation. Options Map, LDB, or Bolt")
	cloneDBPtr := flag.String("clonedb", "", "Override the main node and use this database for the clones in a Network.")
	networkNamePtr := flag.String("network", "", "Network to join: MAIN, TEST or LOCAL")
	peersPtr := flag.String("peers", "", "Array of peer addresses. ")
	blkTimePtr := flag.Int("blktime", 0, "Seconds per block.  Production is 600.")
	// TODO: Old fault mechanism -- remove
	//	faultTimeoutPtr := flag.Int("faulttimeout", 99999, "Seconds before considering Federated servers at-fault. Default is 30.")
	runtimeLogPtr := flag.Bool("runtimeLog", false, "If true, maintain runtime logs of messages passed.")
	exclusivePtr := flag.Bool("exclusive", false, "If true, we only dial out to special/trusted peers.")
	PrefixNodePtr := flag.String("Prefix", "", "Prefix the Factom Node Names with this value; used to create leaderless networks.")
	RotatePtr := flag.Bool("Rotate", false, "If true, responsibility is owned by one leader, and Rotated over the leaders.")
	TimeOffsetPtr := flag.Int("timedelta", 0, "Maximum timeDelta in milliseconds to offset each node.  Simulates deltas in system clocks over a network.")
	KeepMismatchPtr := flag.Bool("keepmismatch", false, "If true, do not discard DBStates even when a majority of DBSignatures have a different hash")
	startDelayPtr := flag.Int("startdelay", 10, "Delay to start processing messages, in seconds")
	DeadlinePtr := flag.Int("Deadline", 1000, "Timeout Delay in milliseconds used on Reads and Writes to the network comm")
	CustomNetPtr := flag.String("customnet", "", "This string specifies a custom blockchain network ID.")
	RpcUserflag := flag.String("rpcuser", "", "Username to protect factomd local API with simple HTTP authentication")
	RpcPasswordflag := flag.String("rpcpass", "", "Password to protect factomd local API. Ignored if rpcuser is blank")
	FactomdTLSflag := flag.Bool("tls", false, "Set to true to require encrypted connections to factomd API and Control Panel") //to get tls, run as "factomd -tls=true"
	FactomdLocationsflag := flag.String("selfaddr", "", "comma separated IPAddresses and DNS names of this factomd to use when creating a cert file")
	MemProfileRate := flag.Int("mpr", 512*1024, "Set the Memory Profile Rate to update profiling per X bytes allocated. Default 512K, set to 1 to profile everything, 0 to disable.")
	exposeProfilePtr := flag.Bool("exposeprofiler", false, "Setting this exposes the profiling port to outside localhost.")
	factomHomePtr := flag.String("factomhome", "", "Set the Factom home directory. The .factom folder will be placed here if set, otherwise it will default to $HOME")

	logportPtr := flag.String("logPort", "6060", "Port for pprof logging")
	portOverridePtr := flag.Int("port", 0, "Port where we serve WSAPI;  default 8088")
	ControlPanelPortOverridePtr := flag.Int("ControlPanelPort", 0, "Port for control panel webserver;  Default 8090")
	networkPortOverridePtr := flag.Int("networkPort", 0, "Port for p2p network; default 8110")

	FastPtr := flag.Bool("Fast", true, "If true, factomd will Fast-boot from a file.")
	FastLocationPtr := flag.String("Fastlocation", "", "Directory to put the Fast-boot file in.")

	logLvlPtr := flag.String("Loglvl", "none", "Set log level to either: none, debug, info, warning, error, fatal or panic")
	logJsonPtr := flag.Bool("Logjson", false, "Use to set logging to use a json formatting")

	sim_stdinPtr := flag.Bool("sim_stdin", true, "If true, sim control reads from stdin.")

	// Plugins
	PluginPath := flag.String("plugin", "", "Input the path to any plugin binaries")

	// 	Torrent Plugin
	tormanager := flag.Bool("tormanage", false, "Use torrent dbstate manager. Must have plugin binary installed and in $PATH")
	TorUploader := flag.Bool("torupload", false, "Be a torrent uploader")

	// Logstash connection (if used)
	logstash := flag.Bool("logstash", false, "If true, use Logstash")
	LogstashURL := flag.String("logurl", "localhost:8345", "Endpoint URL for Logstash")

	sync2Ptr := flag.Int("sync2", -1, "Set the initial blockheight for the second Sync pass. Used to force a total sync, or skip unnecessary syncing of entries.")

	DebugConsolePtr := flag.String("debugconsole", "", "Enable DebugConsole on port. localhost:8093 open 8093 and spawns a telnet console, remotehost:8093 open 8093")

	StdoutLogPtr := flag.String("stdoutlog", "", "Log stdout to a file")
	StderrLogPtr := flag.String("stderrlog", "", "Log stderr to a file, optionally the same file as stdout")
	flag.StringVar(&globals.DebugLogRegEx, "debuglog", "off", "regex to pick which logs to save")
	flag.IntVar(&elections.FaultTimeout, "faulttimeout", 30, "Seconds before considering Federated servers at-fault. Default is 30.")
	flag.CommandLine.Parse(args)

	p.AckbalanceHash = *ackBalanceHashPtr
	p.EnableNet = *enablenetPtr
	p.WaitEntries = *waitEntriesPtr
	p.ListenTo = *listenToPtr
	p.Cnt = *cntPtr
	p.Net = *netPtr
	p.Fnet = *fnetPtr
	p.DropRate = *dropPtr
	p.Journal = *journalPtr
	p.Journaling = *journalingPtr
	p.Follower = *followerPtr
	p.Leader = *leaderPtr
	p.Db = *dbPtr
	p.CloneDB = *cloneDBPtr
	p.PortOverride = *portOverridePtr
	p.Peers = *peersPtr
	p.NetworkName = *networkNamePtr
	p.NetworkPortOverride = *networkPortOverridePtr
	p.ControlPanelPortOverride = *ControlPanelPortOverridePtr
	p.LogPort = *logportPtr
	p.BlkTime = *blkTimePtr
	//	p.FaultTimeout = *faultTimeoutPtr
	p.RuntimeLog = *runtimeLogPtr
	p.Exclusive = *exclusivePtr
	p.Prefix = *PrefixNodePtr
	p.Rotate = *RotatePtr
	p.TimeOffset = *TimeOffsetPtr
	p.KeepMismatch = *KeepMismatchPtr
	p.StartDelay = int64(*startDelayPtr)
	p.Deadline = *DeadlinePtr
	p.CustomNet = primitives.Sha([]byte(*CustomNetPtr)).Bytes()[:4]
	p.RpcUser = *RpcUserflag
	p.RpcPassword = *RpcPasswordflag
	p.FactomdTLS = *FactomdTLSflag
	p.FactomdLocations = *FactomdLocationsflag
	p.MemProfileRate = *MemProfileRate
	p.Fast = *FastPtr
	p.FastLocation = *FastLocationPtr
	p.Loglvl = *logLvlPtr
	p.Logjson = *logJsonPtr
	p.Sim_Stdin = *sim_stdinPtr
	p.ExposeProfiling = *exposeProfilePtr

	p.PluginPath = *PluginPath
	p.TorManage = *tormanager
	p.TorUpload = *TorUploader

	p.UseLogstash = *logstash
	p.LogstashURL = *LogstashURL

	p.Sync2 = *sync2Ptr
	p.DebugConsole = *DebugConsolePtr
	p.StdoutLog = *StdoutLogPtr
	p.StderrLog = *StderrLogPtr

	if *factomHomePtr != "" {
		os.Setenv("FACTOM_HOME", *factomHomePtr)
	}

	return p
}
