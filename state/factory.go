package state

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FactomProject/factomd/Utilities/CorrectChainHeads/correctChainHeads"
	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/modules/logging"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/util"
)

func (s *State) LoadConfigFromFile(filename string, networkFlag string) {
	s.ConfigFilePath = filename
	s.ReadCfg(filename)

	// Get our factomd configuration information.
	cfg := s.GetCfg().(*util.FactomdConfig)

	s.Network = cfg.App.Network
	if 0 < len(networkFlag) { // Command line overrides the config file.
		s.Network = networkFlag
		globals.Params.NetworkName = networkFlag // in case it did not come from there.
	} else {
		globals.Params.NetworkName = s.Network
	}
	fmt.Printf("\n\nNetwork : %s\n", s.Network)

	networkName := strings.ToLower(s.Network) + "-"
	// TODO: improve the paths after milestone 1
	cfg.App.LdbPath = cfg.App.HomeDir + networkName + cfg.App.LdbPath
	cfg.App.BoltDBPath = cfg.App.HomeDir + networkName + cfg.App.BoltDBPath
	cfg.App.DataStorePath = cfg.App.HomeDir + networkName + cfg.App.DataStorePath
	cfg.Log.LogPath = cfg.App.HomeDir + networkName + cfg.Log.LogPath
	cfg.App.ExportDataSubpath = cfg.App.HomeDir + networkName + cfg.App.ExportDataSubpath
	cfg.App.PeersFile = cfg.App.HomeDir + networkName + cfg.App.PeersFile
	cfg.App.ControlPanelFilesPath = cfg.App.HomeDir + cfg.App.ControlPanelFilesPath

	s.LogPath = cfg.Log.LogPath + s.Prefix
	s.LdbPath = cfg.App.LdbPath + s.Prefix
	s.BoltDBPath = cfg.App.BoltDBPath + s.Prefix
	s.LogLevel = cfg.Log.LogLevel
	s.ConsoleLogLevel = cfg.Log.ConsoleLogLevel
	s.NodeMode = cfg.App.NodeMode
	s.DBType = cfg.App.DBType
	s.ExportData = cfg.App.ExportData // bool
	s.ExportDataSubpath = cfg.App.ExportDataSubpath
	s.MainNetworkPort = cfg.App.MainNetworkPort
	s.PeersFile = cfg.App.PeersFile
	s.MainSeedURL = cfg.App.MainSeedURL
	s.MainSpecialPeers = cfg.App.MainSpecialPeers
	s.TestNetworkPort = cfg.App.TestNetworkPort
	s.TestSeedURL = cfg.App.TestSeedURL
	s.TestSpecialPeers = cfg.App.TestSpecialPeers
	s.CustomBootstrapIdentity = cfg.App.CustomBootstrapIdentity
	s.CustomBootstrapKey = cfg.App.CustomBootstrapKey
	s.LocalNetworkPort = cfg.App.LocalNetworkPort
	s.LocalSeedURL = cfg.App.LocalSeedURL
	s.LocalSpecialPeers = cfg.App.LocalSpecialPeers
	s.LocalServerPrivKey = cfg.App.LocalServerPrivKey
	s.CustomNetworkPort = cfg.App.CustomNetworkPort
	s.CustomSeedURL = cfg.App.CustomSeedURL
	s.CustomSpecialPeers = cfg.App.CustomSpecialPeers
	s.FactoshisPerEC = cfg.App.ExchangeRate
	s.DirectoryBlockInSeconds = cfg.App.DirectoryBlockInSeconds
	s.PortNumber = cfg.App.PortNumber
	s.ControlPanelPort = cfg.App.ControlPanelPort
	s.RpcUser = cfg.App.FactomdRpcUser
	s.RpcPass = cfg.App.FactomdRpcPass
	// if RequestTimeout is not set by the configuration it will default to 0.
	//		If it is 0, the loop that uses it will set it to the blocktime/20
	//		We set it there, as blktime might change after this function (from mainnet selection)
	s.RequestTimeout = time.Duration(cfg.App.RequestTimeout) * time.Second
	s.RequestLimit = cfg.App.RequestLimit

	s.StateSaverStruct.FastBoot = cfg.App.FastBoot
	s.StateSaverStruct.FastBootLocation = cfg.App.FastBootLocation
	s.FastBoot = cfg.App.FastBoot
	s.FastBootLocation = cfg.App.FastBootLocation

	// to test run curl -H "Origin: http://anotherexample.com" -H "Access-Control-Request-Method: POST" /
	//     -H "Access-Control-Request-Headers: X-Requested-With" -X POST /
	//     --data-binary '{"jsonrpc": "2.0", "id": 0, "method": "heights"}' -H 'content-type:text/plain;'  /
	//     --verbose http://localhost:8088/v2

	// while the config file has http://anotherexample.com in parameter CorsDomains the response should contain the string
	// < Access-Control-Allow-Origin: http://anotherexample.com

	if len(cfg.App.CorsDomains) > 0 {
		domains := strings.Split(cfg.App.CorsDomains, ",")
		s.CorsDomains = make([]string, len(domains))
		for _, domain := range domains {
			s.CorsDomains = append(s.CorsDomains, strings.Trim(domain, " "))
		}
	}
	s.FactomdTLSEnable = cfg.App.FactomdTlsEnabled

	FactomdTLSKeyFile := cfg.App.FactomdTlsPrivateKey
	if cfg.App.FactomdTlsPrivateKey == "/full/path/to/factomdAPIpriv.key" {
		FactomdTLSKeyFile = fmt.Sprint(cfg.App.HomeDir, "factomdAPIpriv.key")
	}
	if s.FactomdTLSKeyFile != FactomdTLSKeyFile {
		if s.FactomdTLSEnable {
			if _, err := os.Stat(FactomdTLSKeyFile); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Configured file does not exits: %s\n", FactomdTLSKeyFile)
			}
		}
		s.FactomdTLSKeyFile = FactomdTLSKeyFile // set state
	}

	FactomdTLSCertFile := cfg.App.FactomdTlsPublicCert
	if cfg.App.FactomdTlsPublicCert == "/full/path/to/factomdAPIpub.cert" {
		s.FactomdTLSCertFile = fmt.Sprint(cfg.App.HomeDir, "factomdAPIpub.cert")
	}
	if s.FactomdTLSCertFile != FactomdTLSCertFile {
		if s.FactomdTLSEnable {
			if _, err := os.Stat(FactomdTLSCertFile); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Configured file does not exits: %s\n", FactomdTLSCertFile)
			}
		}
		s.FactomdTLSCertFile = FactomdTLSCertFile // set state
	}

	s.FactomdTLSEnable = cfg.App.FactomdTlsEnabled
	s.FactomdTLSKeyFile = cfg.App.FactomdTlsPrivateKey

	externalIP := strings.Split(cfg.Walletd.FactomdLocation, ":")[0]
	if externalIP != "localhost" {
		s.FactomdLocations = externalIP
	}

	switch cfg.App.ControlPanelSetting {
	case "disabled":
		s.ControlPanelSetting = 0
	case "readonly":
		s.ControlPanelSetting = 1
	case "readwrite":
		s.ControlPanelSetting = 2
	default:
		s.ControlPanelSetting = 1
	}
	s.FERChainId = cfg.App.ExchangeRateChainId
	s.ExchangeRateAuthorityPublicKey = cfg.App.ExchangeRateAuthorityPublicKey
	identity, err := primitives.HexToHash(cfg.App.IdentityChainID)
	if err != nil {
		s.IdentityChainID = primitives.Sha([]byte(s.FactomNodeName))
		s.LogPrintf("AckChange", "Bad IdentityChainID  in config \"%v\"", cfg.App.IdentityChainID)
		s.LogPrintf("AckChange", "Default2 IdentityChainID \"%v\"", s.IdentityChainID.String())
	} else {
		s.IdentityChainID = identity
		s.LogPrintf("AckChange", "Load IdentityChainID \"%v\"", s.IdentityChainID.String())
	}

	if cfg.App.P2PIncoming > 0 {
		p2p.MaxNumberIncomingConnections = cfg.App.P2PIncoming
	}
	if cfg.App.P2POutgoing > 0 {
		p2p.NumberPeersToConnect = cfg.App.P2POutgoing
	}
}

func (s *State) LoadConfigDefaults() {
	s.LogPath = "database/"
	s.LdbPath = "database/ldb"
	s.BoltDBPath = "database/bolt"
	s.LogLevel = "none"
	s.ConsoleLogLevel = "standard"
	s.NodeMode = "SERVER"
	s.DBType = "Map"
	s.ExportData = false
	s.ExportDataSubpath = "data/export"
	s.Network = "TEST"
	s.MainNetworkPort = "8108"
	s.PeersFile = "peers.json"
	s.MainSeedURL = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/mainseed.txt"
	s.MainSpecialPeers = ""
	s.TestNetworkPort = "8109"
	s.TestSeedURL = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/testseed.txt"
	s.TestSpecialPeers = ""
	s.LocalNetworkPort = "8110"
	s.LocalSeedURL = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/localseed.txt"
	s.LocalSpecialPeers = ""

	s.LocalServerPrivKey = "4c38c72fc5cdad68f13b74674d3ffb1f3d63a112710868c9b08946553448d26d"
	s.FactoshisPerEC = 006666
	s.FERChainId = "111111118d918a8be684e0dac725493a75862ef96d2d3f43f84b26969329bf03"
	s.ExchangeRateAuthorityPublicKey = "3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"
	s.DirectoryBlockInSeconds = 6
	s.PortNumber = 8088
	s.ControlPanelPort = 8090
	s.ControlPanelSetting = 1

	// TODO:  Actually load the IdentityChainID from the config file
	s.IdentityChainID = primitives.Sha([]byte(s.FactomNodeName))
	s.LogPrintf("AckChange", "Default IdentityChainID %v", s.IdentityChainID.String())
}

func (s *State) LoadConfig(filename string, networkFlag string) {
	if len(filename) > 0 {
		s.LoadConfigFromFile(filename, networkFlag)
	} else {
		s.LoadConfigDefaults()
	}
	s.updateNetworkControllerConfig()
}

// original constructor
func NewState(p *globals.FactomParams, FactomdVersion string) *State {
	s := new(State)
	// Must add the prefix before loading the configuration.
	s.AddPrefix(p.Prefix)
	// Setup the name to catch any early logging
	s.FactomNodeName = p.Prefix + "FNode0"
	//s.NameInit(common.NilName, s.FactomNodeName+"State", reflect.TypeOf(s).String())
	s.logging = logging.NewLayerLogger(log.GlobalLogger, map[string]string{"fnode": s.FactomNodeName})

	// print current dbht-:-minute
	s.logging.AddPrintField("dbht",
		func(interface{}) string { return fmt.Sprintf("%7d-:-%-2d", *&s.LLeaderHeight, *&s.CurrentMinute) },
		"")
	s.TimestampAtBoot = primitives.NewTimestampNow()
	preBootTime := new(primitives.Timestamp)
	preBootTime.SetTimeMilli(s.TimestampAtBoot.GetTimeMilli() - 20*60*1000)
	s.SetLeaderTimestamp(s.TimestampAtBoot)
	s.SetMessageFilterTimestamp(preBootTime)
	s.RunState = runstate.New

	// build a timestamp 20 minutes before boot so we will accept inMessages from nodes who booted before us.
	s.PortNumber = 8088
	s.ControlPanelPort = 8090

	FactomConfigFilename := util.GetConfigFilename("m2")
	if p.ConfigPath != "" {
		FactomConfigFilename = p.ConfigPath
	}
	s.LoadConfig(FactomConfigFilename, p.NetworkName)
	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))

	s.TimeOffset = primitives.NewTimestampFromMilliseconds(uint64(p.TimeOffset))
	s.StartDelayLimit = p.StartDelay * 1000
	s.FactomdVersion = FactomdVersion

	// Set the wait for entries flag
	s.WaitForEntries = p.WaitEntries

	if 999 < p.PortOverride { // The command line flag exists and seems reasonable.
		s.SetPort(p.PortOverride)
	} else {
		p.PortOverride = s.GetPort()
	}
	if 999 < p.ControlPanelPortOverride { // The command line flag exists and seems reasonable.
		s.ControlPanelPort = p.ControlPanelPortOverride
	} else {
		p.ControlPanelPortOverride = s.ControlPanelPort
	}

	if p.BlkTime > 0 {
		s.DirectoryBlockInSeconds = p.BlkTime
	} else {
		p.BlkTime = s.DirectoryBlockInSeconds
	}

	s.FaultTimeout = 9999999 //todo: Old Fault Mechanism -- remove

	if p.RpcUser != "" {
		s.RpcUser = p.RpcUser
	}

	if p.RpcPassword != "" {
		s.RpcPass = p.RpcPassword
	}

	if p.FactomdTLS == true {
		s.FactomdTLSEnable = true
	}

	if p.FactomdLocations != "" {
		if len(s.FactomdLocations) > 0 {
			s.FactomdLocations += ","
		}
		s.FactomdLocations += p.FactomdLocations
	}

	if p.Fast == false {
		s.StateSaverStruct.FastBoot = false
	}
	if p.FastLocation != "" {
		s.StateSaverStruct.FastBootLocation = p.FastLocation
	}
	if p.FastSaveRate < 2 || p.FastSaveRate > 5000 {
		panic("FastSaveRate must be between 2 and 5000")
	}
	s.FastSaveRate = p.FastSaveRate

	s.CheckChainHeads.CheckChainHeads = p.CheckChainHeads
	s.CheckChainHeads.Fix = p.FixChainHeads

	if p.P2PIncoming > 0 {
		p2p.MaxNumberIncomingConnections = p.P2PIncoming
	}
	if p.P2POutgoing > 0 {
		p2p.NumberPeersToConnect = p.P2POutgoing
	}

	// Command line override if provided
	switch p.ControlPanelSetting {
	case "disabled":
		s.ControlPanelSetting = 0
	case "readonly":
		s.ControlPanelSetting = 1
	case "readwrite":
		s.ControlPanelSetting = 2
	}

	s.UseLogstash = p.UseLogstash
	s.LogstashURL = p.LogstashURL

	if len(p.Db) > 0 {
		s.DBType = p.Db
	} else {
		p.Db = s.DBType
	}

	if len(p.CloneDB) > 0 {
		s.CloneDBType = p.CloneDB
	} else {
		s.CloneDBType = p.Db
	}

	s.AddPrefix(p.Prefix)
	s.SetOut(false)
	s.SetDropRate(p.DropRate)

	// publishing hooks for new modules

	return s
}

func Clone(s *State, cloneNumber int) interfaces.IState {
	newState := new(State)
	newState.StateConfig = s.StateConfig
	number := fmt.Sprintf("%02d", cloneNumber)
	newState.FactomNodeName = s.Prefix + "FNode" + number
	// the DBHT value is replaced by the result of running the formatter for dbht which has the current value
	newState.logging = logging.NewLayerLogger(log.GlobalLogger, map[string]string{"fnode": newState.FactomNodeName, "dbht": "unused"})
	newState.logging.AddPrintField("dbht",
		func(interface{}) string { return fmt.Sprintf("%7d-:-%-2d", *&s.LLeaderHeight, *&s.CurrentMinute) },
		"") // the
	simConfigPath := util.GetHomeDir() + "/.factom/m2/simConfig/"
	configfile := fmt.Sprintf("%sfactomd%03d.conf", simConfigPath, cloneNumber)

	if cloneNumber >= 1 {
		os.Stderr.WriteString(fmt.Sprintf("Looking for Config File %s\n", configfile))
	}
	if _, err := os.Stat(simConfigPath); os.IsNotExist(err) {
		os.Stderr.WriteString("Creating simConfig directory\n")
		os.MkdirAll(simConfigPath, 0775)
	}

	config := false
	if _, err := os.Stat(configfile); !os.IsNotExist(err) {
		os.Stderr.WriteString(fmt.Sprintf("   Using the %s config file.\n", configfile))
		newState.LoadConfig(configfile, s.GetNetworkName())
		config = true
	}

	if s.LogPath == "stdout" {
		newState.LogPath = "stdout"
	} else {
		newState.LogPath = s.LogPath + "/Sim" + number
	}

	newState.RunState = runstate.New // reset runstate since this clone will be started by sim node
	newState.LdbPath = s.LdbPath + "/Sim" + number
	newState.BoltDBPath = s.BoltDBPath + "/Sim" + number
	newState.ExportDataSubpath = s.ExportDataSubpath + "sim-" + number
	newState.IdentityControl = s.IdentityControl.Clone() // FIXME relocate

	if !config {
		// FIXME: add hack so wew can do Fnode00, Fnode01, ...
		newState.IdentityChainID = primitives.Sha([]byte(newState.FactomNodeName))
		s.LogPrintf("AckChange", "Default3 IdentityChainID %v", s.IdentityChainID.String())

		//generate and use a new deterministic PrivateKey for this clone
		shaHashOfNodeName := primitives.Sha([]byte(newState.FactomNodeName)) //seed the private key with node name
		clonePrivateKey := primitives.NewPrivateKeyFromHexBytes(shaHashOfNodeName.Bytes())
		newState.LocalServerPrivKey = clonePrivateKey.PrivateKeyString()
		s.initServerKeys()
	}

	// FIXME change to use timestamp.Clone
	newState.TimestampAtBoot = primitives.NewTimestampFromMilliseconds(s.TimestampAtBoot.GetTimeMilliUInt64())
	newState.LeaderTimestamp = primitives.NewTimestampFromMilliseconds(s.LeaderTimestamp.GetTimeMilliUInt64())
	newState.SetMessageFilterTimestamp(s.GetMessageFilterTimestamp())
	switch newState.DBType {
	case "LDB":
		newState.StateSaverStruct.FastBoot = s.StateSaverStruct.FastBoot
		newState.StateSaverStruct.FastBootLocation = newState.LdbPath
		break
	case "Bolt":
		newState.StateSaverStruct.FastBoot = s.StateSaverStruct.FastBoot
		newState.StateSaverStruct.FastBootLocation = newState.BoltDBPath
		break
	}
	if globals.Params.WriteProcessedDBStates {
		path := filepath.Join(newState.LdbPath, newState.Network, "dbstates")
		os.MkdirAll(path, 0775)
	}
	return newState
}

func (s *State) Initialize(o common.NamedObject, electionFactory interfaces.IElectionsFactory) {
	if s.Salt == nil {
		b := make([]byte, 32)
		_, err := rand.Read(b)
		if err != nil {
			panic("Random Number Failure")
		}
		s.Salt = primitives.Sha(b)
	}

	fmt.Printf("The Instance ID of this node is %s\n", s.Salt.String()[:16])

	s.StartDelay = s.GetTimestamp().GetTimeMilli() // We can't start as a leader until we know we are upto date
	s.RunLeader = false
	s.IgnoreMissing = true
	s.BootTime = s.GetTimestamp().GetTimeSeconds()
	s.TimestampAtBoot = primitives.NewTimestampNow()
	s.ProcessTime = s.TimestampAtBoot
	if s.LogPath == "stdout" {
		//wsapi.InitLogs(s.LogPath, s.LogLevel)
	} else {
		er := os.MkdirAll(s.LogPath, 0775)
		if er != nil {
			panic("Could not create " + s.LogPath + "\n error: " + er.Error())
		}
		//wsapi.InitLogs(s.LogPath+s.FactomNodeName+".log", s.LogLevel)
	}

	s.Hold = NewHoldingList(s)                                // setup the dependent holding map
	s.TimeOffset = new(primitives.Timestamp)                  //interfaces.Timestamp(int64(rand.Int63() % int64(time.Microsecond*10)))
	s.InvalidMessages = make(map[[32]byte]interfaces.IMsg, 0) //
	s.ShutdownChan = make(chan int, 1)                        //SubChannel to gracefully shut down.
	s.tickerQueue = make(chan int, 100)                       //ticks from a clock
	s.timerMsgQueue = make(chan interfaces.IMsg, 100)         //incoming eom notifications, used by leaders
	//	s.ControlPanelChannel = make(chan DisplayState, 20)                     //
	s.networkInvalidMsgQueue = make(chan interfaces.IMsg, 100)              //incoming message queue from the network inMessages
	s.networkOutMsgQueue = NewNetOutMsgQueue(s, constants.INMSGQUEUE_MED)   //Messages to be broadcast to the network
	s.inMsgQueue = NewInMsgQueue(s, constants.INMSGQUEUE_HIGH)              //incoming message queue for Factom application inMessages
	s.inMsgQueue2 = NewInMsgQueue2(s, constants.INMSGQUEUE_HIGH)            //incoming message queue for Factom application inMessages
	s.electionsQueue = NewElectionQueue(s, constants.INMSGQUEUE_HIGH)       //incoming message queue for Factom application inMessages
	s.apiQueue = NewAPIQueue(s, constants.INMSGQUEUE_HIGH)                  //incoming message queue from the API
	s.ackQueue = make(chan interfaces.IMsg, 50)                             //queue of Leadership inMessages
	s.msgQueue = make(chan interfaces.IMsg, 50)                             //queue of Follower inMessages
	s.prioritizedMsgQueue = make(chan interfaces.IMsg, 50)                  //a prioritized queue of Follower inMessages (from mmr.go)
	s.MissingEntries = make(chan *MissingEntry, constants.INMSGQUEUE_HIGH)  //Entries I discover are missing from the database
	s.UpdateEntryHash = make(chan *EntryUpdate, constants.INMSGQUEUE_HIGH)  //Handles entry hashes and updating Commit maps.
	s.WriteEntry = make(chan interfaces.IEBEntry, constants.INMSGQUEUE_LOW) //Entries to be written to the database
	s.RecentMessage.NewMsgs = make(chan interfaces.IMsg, 100)

	// Set up struct to stop replay attacks
	s.Replay = new(Replay)
	s.Replay.s = s
	s.Replay.name = "Replay"

	s.FReplay = new(Replay)
	s.FReplay.s = s
	s.FReplay.name = "FReplay"

	// Set up maps for the followers
	s.Holding = make(map[[32]byte]interfaces.IMsg)
	s.HoldingList = make(chan [32]byte, 4000)
	s.Acks = make(map[[32]byte]interfaces.IMsg)
	s.Commits = NewSafeMsgMap("commits", s) //make(map[[32]byte]interfaces.IMsg)

	// Setup the FactoidState and Validation Service that holds factoid and entry credit balances
	s.FactoidBalancesP = map[[32]byte]int64{}
	s.ECBalancesP = map[[32]byte]int64{}

	fs := new(FactoidState)
	fs.State = s
	s.FactoidState = fs

	// Allocate the original set of Process Lists
	s.ProcessLists = NewProcessLists(s)
	s.FaultWait = 3
	s.LastTiebreak = 0
	s.EOMfaultIndex = 0

	s.DBStates = new(DBStateList)
	s.DBStates.State = s
	s.DBStates.DBStates = make([]*DBState, 0)

	s.StatesMissing = NewStatesMissing()
	s.StatesWaiting = NewStatesWaiting()
	s.StatesReceived = NewStatesReceived()

	switch s.NodeMode {
	case "FULL":
		s.Leader = false
		s.Println("\n   +---------------------------+")
		s.Println("   +------ Follower Only ------+")
		s.Print("   +---------------------------+\n\n")
	case "SERVER":
		s.Println("\n   +-------------------------+")
		s.Println("   |       Leader Node       |")
		s.Print("   +-------------------------+\n\n")
	default:
		panic("Bad Node Mode (must be FULL or SERVER)")
	}

	//Database
	switch s.DBType {
	case "LDB":
		if err := s.InitLevelDB(); err != nil {
			panic(fmt.Sprintf("Error initializing the database: %v", err))
		}
	case "Bolt":
		if err := s.InitBoltDB(); err != nil {
			panic(fmt.Sprintf("Error initializing the database: %v", err))
		}
	case "Map":
		if err := s.InitMapDB(); err != nil {
			panic(fmt.Sprintf("Error initializing the database: %v", err))
		}
	default:
		panic("No Database type specified")
	}

	if s.CheckChainHeads.CheckChainHeads {
		if s.CheckChainHeads.Fix {
			// Set dblock head to 184 if 184 is present and head is not 184
			d, err := s.DB.FetchDBlockHead()
			if err != nil {
				// We should have a dblock head...
				panic(fmt.Errorf("Error loading dblock head: %s\n", err.Error()))
			}

			if d != nil {
				if d.GetDatabaseHeight() == 160183 {
					// Our head is less than 160184, do we have 160184?
					if d2, err := s.DB.FetchDBlockByHeight(160184); d2 != nil && err == nil {
						err := s.DB.(*databaseOverlay.Overlay).SaveDirectoryBlockHead(d2)
						if err != nil {
							panic(err)
						}
					}
				}
			}
		}
		correctChainHeads.FindHeads(s.DB.(*databaseOverlay.Overlay), correctChainHeads.CorrectChainHeadConfig{
			PrintFreq: 5000,
			Fix:       s.CheckChainHeads.Fix,
		})
	}
	if s.ExportData {
		s.DB.SetExportData(s.ExportDataSubpath)
	}

	// Cross Boot Replay
	switch s.DBType {
	case "Map":
		s.SetupCrossBootReplay("Map")
	default:
		s.SetupCrossBootReplay("Bolt")
	}

	//Network
	switch s.Network {
	case "MAIN":
		s.NetworkNumber = constants.NETWORK_MAIN
		s.DirectoryBlockInSeconds = 600
	case "TEST":
		s.NetworkNumber = constants.NETWORK_TEST
	case "LOCAL":
		s.NetworkNumber = constants.NETWORK_LOCAL
	case "CUSTOM":
		s.NetworkNumber = constants.NETWORK_CUSTOM
	default:
		panic("Bad value for Network in factomd.conf")
	}

	s.Println("\nRunning on the ", s.Network, "Network")
	s.Println("\nExchange rate chain id set to ", s.FERChainId)
	s.Println("\nExchange rate Authority Public Key set to ", s.ExchangeRateAuthorityPublicKey)

	s.AuditHeartBeats = make([]interfaces.IMsg, 0)

	// If we cloned the Identity control of another node, don't reset!
	if s.IdentityControl == nil {
		s.IdentityControl = identity.NewIdentityManager()
	}
	s.initServerKeys()
	s.AuthorityServerCount = 0

	//LoadIdentityCache(s)
	//StubIdentityCache(s)
	//needed for multiple nodes with FER.  remove for singe node launch
	if s.FERChainId == "" {
		s.FERChainId = "111111118d918a8be684e0dac725493a75862ef96d2d3f43f84b26969329bf03"
	}
	if s.ExchangeRateAuthorityPublicKey == "" {
		s.ExchangeRateAuthorityPublicKey = "3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29"
	}
	// end of FER removal
	s.Starttime = time.Now()
	// Allocate the MMR queues
	s.asks = make(chan askRef, 100) // Should be > than the number of VMs so each VM can have at least one outstanding ask.
	s.adds = make(chan plRef, 50)   // No good rule of thumb on the size of this
	s.dbheights = make(chan int, 1)
	s.rejects = make(chan MsgPair, 1) // Messages rejected from process list

	// Allocate the missing message handler
	s.MissingMessageResponseHandler = NewMissingMessageReponseCache(s)
	// Election factory was created and passed int to avoid import loop
	s.EFactory = electionFactory

	if s.StateSaverStruct.FastBoot {
		d, err := s.DB.FetchDBlockHead()
		if err != nil {
			panic(err)
		}

		if d == nil || int(d.GetDatabaseHeight()) < s.FastSaveRate {
			//If we have less than whatever our block rate is, we wipe SaveState
			//This is to ensure we don't accidentally keep SaveState while deleting a database
			s.StateSaverStruct.DeleteSaveState(s.Network)
		} else {
			err = s.StateSaverStruct.LoadDBStateList(s, s.DBStates, s.Network)
			if err != nil {
				s.StateSaverStruct.DeleteSaveState(s.Network)
				s.LogPrintf("faulting", "Database load failed %v", err)
			}
			if err == nil {
				for _, dbstate := range s.DBStates.DBStates {
					if dbstate != nil {
						dbstate.SaveStruct.Commits.s = s
					}
				}
			}
		}
	}

	if globals.Params.WriteProcessedDBStates {
		path := filepath.Join(s.LdbPath, s.Network, "dbstates")
		os.MkdirAll(path, 0775)
	}

	// Setup the Skeleton Identity & Registration
	s.IntiateNetworkSkeletonIdentity()
	s.InitiateNetworkIdentityRegistration()
}
