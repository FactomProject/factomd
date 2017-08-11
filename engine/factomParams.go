package engine

import (
	"flag"
	"os"

	"github.com/FactomProject/factomd/common/primitives"
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
	Netdebug                 int
	Exclusive                bool
	prefix                   string
	rotate                   bool
	timeOffset               int
	keepMismatch             bool
	startDelay               int64
	deadline                 int
	customNet                []byte
	rpcUser                  string
	rpcPassword              string
	factomdTLS               bool
	factomdLocations         string
	memProfileRate           int
	fast                     bool
	fastLocation             string
	loglvl                   string
	logjson                  bool
	svm                      bool
	pluginPath               string
	torManage                bool
	torUpload                bool
	disableSimControl        bool
	exposeProfiling          bool
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
	journalingPtr := flag.Bool("journaling", false, "Write a journal of all messages recieved. Default is off.")
	followerPtr := flag.Bool("follower", false, "If true, force node to be a follower.  Only used when replaying a journal.")
	leaderPtr := flag.Bool("leader", true, "If true, force node to be a leader.  Only used when replaying a journal.")
	dbPtr := flag.String("db", "", "Override the Database in the Config file and use this Database implementation. Options Map, LDB, or Bolt")
	cloneDBPtr := flag.String("clonedb", "", "Override the main node and use this database for the clones in a Network.")
	networkNamePtr := flag.String("network", "", "Network to join: MAIN, TEST or LOCAL")
	peersPtr := flag.String("peers", "", "Array of peer addresses. ")
	blkTimePtr := flag.Int("blktime", 0, "Seconds per block.  Production is 600.")
	faultTimeoutPtr := flag.Int("faulttimeout", 60, "Seconds before considering Federated servers at-fault. Default is 60.")
	runtimeLogPtr := flag.Bool("runtimeLog", false, "If true, maintain runtime logs of messages passed.")
	netdebugPtr := flag.Int("netdebug", 0, "0-5: 0 = quiet, >0 = increasing levels of logging")
	exclusivePtr := flag.Bool("exclusive", false, "If true, we only dial out to special/trusted peers.")
	prefixNodePtr := flag.String("prefix", "", "Prefix the Factom Node Names with this value; used to create leaderless networks.")
	rotatePtr := flag.Bool("rotate", false, "If true, responsiblity is owned by one leader, and rotated over the leaders.")
	timeOffsetPtr := flag.Int("timedelta", 0, "Maximum timeDelta in milliseconds to offset each node.  Simulates deltas in system clocks over a network.")
	keepMismatchPtr := flag.Bool("keepmismatch", false, "If true, do not discard DBStates even when a majority of DBSignatures have a different hash")
	startDelayPtr := flag.Int("startdelay", 10, "Delay to start processing messages, in seconds")
	deadlinePtr := flag.Int("deadline", 1000, "Timeout Delay in milliseconds used on Reads and Writes to the network comm")
	customNetPtr := flag.String("customnet", "", "This string specifies a custom blockchain network ID.")
	rpcUserflag := flag.String("rpcuser", "", "Username to protect factomd local API with simple HTTP authentication")
	rpcPasswordflag := flag.String("rpcpass", "", "Password to protect factomd local API. Ignored if rpcuser is blank")
	factomdTLSflag := flag.Bool("tls", false, "Set to true to require encrypted connections to factomd API and Control Panel") //to get tls, run as "factomd -tls=true"
	factomdLocationsflag := flag.String("selfaddr", "", "comma seperated IPAddresses and DNS names of this factomd to use when creating a cert file")
	memProfileRate := flag.Int("mpr", 512*1024, "Set the Memory Profile Rate to update profiling per X bytes allocated. Default 512K, set to 1 to profile everything, 0 to disable.")
	exposeProfilePtr := flag.Bool("exposeprofiler", false, "Setting this exposes the profiling port to outside localhost.")
	factomHomePtr := flag.String("factomhome", "", "Set the factom home directory. The .factom folder will be placed here if set, otherwise it will default to $HOME")

	logportPtr := flag.String("logPort", "6060", "Port for pprof logging")
	portOverridePtr := flag.Int("port", 0, "Port where we serve WSAPI;  default 8088")
	ControlPanelPortOverridePtr := flag.Int("ControlPanelPort", 0, "Port for control panel webserver;  Default 8090")
	networkPortOverridePtr := flag.Int("networkPort", 0, "Port for p2p network; default 8110")

	fastPtr := flag.Bool("fast", true, "If true, factomd will fast-boot from a file.")
	fastLocationPtr := flag.String("fastlocation", "", "Directory to put the fast-boot file in.")

	logLvlPtr := flag.String("loglvl", "none", "Set log level to either: none, debug, info, warning, error, fatal or panic")
	logJsonPtr := flag.Bool("logjson", false, "Use to set logging to use a json formatting")

	superVerboseMessages := flag.Bool("svm", false, "If true, print out every single message as you receive it.")

	disableSimControlPtr := flag.Bool("nosim", false, "Using this flag disables sim control")

	// Plugins
	pluginPath := flag.String("plugin", "", "Input the path to any plugin binaries")

	// 	Torrent Plugin
	tormanager := flag.Bool("tormanage", false, "Use torrent dbstate manager. Must have plugin binary installed and in $PATH")
	torUploader := flag.Bool("torupload", false, "Be a torrent uploader")

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
	p.FaultTimeout = *faultTimeoutPtr
	p.RuntimeLog = *runtimeLogPtr
	p.Netdebug = *netdebugPtr
	p.Exclusive = *exclusivePtr
	p.prefix = *prefixNodePtr
	p.rotate = *rotatePtr
	p.timeOffset = *timeOffsetPtr
	p.keepMismatch = *keepMismatchPtr
	p.startDelay = int64(*startDelayPtr)
	p.deadline = *deadlinePtr
	p.customNet = primitives.Sha([]byte(*customNetPtr)).Bytes()[:4]
	p.rpcUser = *rpcUserflag
	p.rpcPassword = *rpcPasswordflag
	p.factomdTLS = *factomdTLSflag
	p.factomdLocations = *factomdLocationsflag
	p.memProfileRate = *memProfileRate
	p.fast = *fastPtr
	p.fastLocation = *fastLocationPtr
	p.loglvl = *logLvlPtr
	p.logjson = *logJsonPtr
	p.disableSimControl = *disableSimControlPtr
	p.exposeProfiling = *exposeProfilePtr

	p.svm = *superVerboseMessages
	p.pluginPath = *pluginPath
	p.torManage = *tormanager
	p.torUpload = *torUploader

	if *factomHomePtr != "" {
		os.Setenv("FACTOM_HOME", *factomHomePtr)
	}

	return p
}
