package globals

import (
	"fmt"
	"os"
	"reflect"
	"text/tabwriter"
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

	// LiveFeed API params
	EnableLiveFeedAPI        bool
	EventReceiverProtocol    string
	EventReceiverHost        string
	EventReceiverPort        int
	EventSenderPort          int
	EventFormat              string
	EventSendStateChange     bool
	EventBroadcastContent    string
	EventReplayDuringStartup bool
	PersistentReconnect      bool
}

// PrettyPrint will print all the struct fields and their values to Stdout
func (p *FactomParams) PrettyPrint() {
	fmt.Println("Parameters:")
	tw := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	s := reflect.ValueOf(p).Elem()
	t := s.Type()

	for i := 0; i < s.NumField(); i++ {
		fmt.Fprintf(tw, "%2d:\t%25s\t%s\t=\t%v\t\n",
			i, t.Field(i).Name, s.Field(i).Type(), s.Field(i).Interface())
	}

	tw.Flush()
}
