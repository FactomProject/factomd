// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/constants/runstate"
	. "github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

type FactomNode struct {
	Index    int
	State    *state.State
	Peers    []interfaces.IPeer
	MLog     *MsgLog
	P2PIndex int
}

var fnodes []*FactomNode

var networkpattern string
var mLog = new(MsgLog)
var p2pProxy *P2PProxy
var network *p2p.Network
var logPort string

func GetFnodes() []*FactomNode {
	return fnodes
}

func init() {
	messages.General = new(msgsupport.GeneralFactory)
	primitives.General = messages.General
}

func NetStart(s *state.State, p *FactomParams, listenToStdin bool) {

	s.PortNumber = 8088
	s.ControlPanelPort = 8090
	logPort = p.LogPort

	messages.AckBalanceHash = p.AckbalanceHash
	// Must add the prefix before loading the configuration.
	s.AddPrefix(p.Prefix)
	FactomConfigFilename := util.GetConfigFilename("m2")
	if p.ConfigPath != "" {
		FactomConfigFilename = p.ConfigPath
	}
	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))
	s.LoadConfig(FactomConfigFilename, p.NetworkName)
	s.OneLeader = p.Rotate
	s.TimeOffset = primitives.NewTimestampFromMilliseconds(uint64(p.TimeOffset))
	s.StartDelayLimit = p.StartDelay * 1000
	s.Journaling = p.Journaling
	s.FactomdVersion = FactomdVersion
	s.EFactory = new(electionMsgs.ElectionsFactory)

	log.SetOutput(os.Stdout)
	switch strings.ToLower(p.Loglvl) {
	case "none":
		log.SetOutput(ioutil.Discard)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning", "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
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

	if p.Logjson {
		log.SetFormatter(&log.JSONFormatter{})
	}

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

	if p.Follower {
		p.Leader = false
	}
	if p.Leader {
		p.Follower = false
	}
	if !p.Follower && !p.Leader {
		panic("Not a leader or a follower")
	}

	if p.Journal != "" {
		p.Cnt = 1
	}

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

	p2pconf := p2p.DefaultP2PConfiguration()
	p2pconf.TargetPeers = 32
	p2pconf.DropTo = 30
	p2pconf.MaxPeers = 36
	p2pconf.Fanout = 16
	p2pconf.ChannelCapacity = 5000
	p2pconf.PingInterval = time.Second * 15
	p2pconf.ProtocolVersion = 10

	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Net Sim Start!")
	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Listening to Node", p.ListenTo)
	fmt.Println(">>>>>>>>>>>>>>>>")

	AddInterruptHandler(func() {
		fmt.Print("<Break>\n")
		fmt.Print("Gracefully shutting down the server...\n")
		for _, fnode := range fnodes {
			fnode.State.ShutdownNode(0)
		}
		if p.EnableNet {
			network.Stop()
		}
		fmt.Print("Waiting...\r\n")
		time.Sleep(3 * time.Second)
		os.Exit(0)
	})

	if p.Journal != "" {
		if s.DBType != "Map" {
			fmt.Println("Journal is ALWAYS a Map database")
			s.DBType = "Map"
		}
	}
	if p.Follower {
		s.NodeMode = "FULL"
		leadID := primitives.Sha([]byte(s.Prefix + "FNode0"))
		if s.IdentityChainID.IsSameAs(leadID) {
			s.SetIdentityChainID(primitives.Sha([]byte(time.Now().String()))) // Make sure this node is NOT a leader
		}
	}

	s.KeepMismatch = p.KeepMismatch

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

	pnet := p.Net
	if len(p.Fnet) > 0 {
		pnet = p.Fnet
		p.Net = "file"
	}

	s.UseLogstash = p.UseLogstash
	s.LogstashURL = p.LogstashURL

	go StartProfiler(p.MemProfileRate, p.ExposeProfiling)

	s.AddPrefix(p.Prefix)
	s.SetOut(false)
	s.Init()
	s.SetDropRate(p.DropRate)

	if p.Sync2 >= 0 {
		s.EntryDBHeightComplete = uint32(p.Sync2)
		s.LogPrintf("EntrySync", "Force with Sync2 NetStart EntryDBHeightComplete = %d", s.EntryDBHeightComplete)

	} else {
		height, err := s.DB.FetchDatabaseEntryHeight()
		if err != nil {
			s.LogPrintf("EntrySync", "Error reading EntryDBHeightComplete NetStart EntryDBHeightComplete = %d", s.EntryDBHeightComplete)
			os.Stderr.WriteString(fmt.Sprintf("ERROR reading Entry DBHeight Complete: %v\n", err))
		} else {
			s.EntryDBHeightComplete = height
			s.LogPrintf("EntrySync", "NetStart EntryDBHeightComplete = %d", s.EntryDBHeightComplete)
		}
	}

	mLog.Init(p.RuntimeLog, p.Cnt)

	setupFirstAuthority(s)

	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "Build", Build))
	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "Node name", p.NodeName))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "balancehash", messages.AckBalanceHash))
	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", fmt.Sprintf("%s Salt", s.GetFactomNodeName()), s.Salt.String()[:16]))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "enablenet", p.EnableNet))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "net target", p2pconf.TargetPeers))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "net max", p2pconf.MaxPeers))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "waitentries", p.WaitEntries))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "node", p.ListenTo))
	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "prefix", p.Prefix))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "node count", p.Cnt))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "FastSaveRate", p.FastSaveRate))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "net spec", pnet))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "Msgs droped", p.DropRate))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "journal", p.Journal))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database", p.Db))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database for clones", p.CloneDB))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "peers", p.Peers))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%t\"\n", "exclusive", p.Exclusive))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%t\"\n", "exclusive_in", p.ExclusiveIn))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "block time", p.BlkTime))
	//os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "faultTimeout", p.FaultTimeout)) // TODO old fault timeout mechanism to be removed
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "runtimeLog", p.RuntimeLog))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "rotate", p.Rotate))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "timeOffset", p.TimeOffset))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "keepMismatch", p.KeepMismatch))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "startDelay", p.StartDelay))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "Network", s.Network))
	os.Stderr.WriteString(fmt.Sprintf("%20s %x (%s)\n", "customnet", p.CustomNet, p.CustomNetName))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "deadline (ms)", p.Deadline))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "tls", s.FactomdTLSEnable))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "selfaddr", s.FactomdLocations))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "rpcuser", s.RpcUser))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "corsdomains", s.CorsDomains))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "Start 2nd Sync at ht", s.EntryDBHeightComplete))

	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "faultTimeout", elections.FaultTimeout))

	if "" == s.RpcPass {
		os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "rpcpass", "is blank"))
	} else {
		os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "rpcpass", "is set"))
	}
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "TCP port", s.PortNumber))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "pprof port", logPort))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "Control Panel port", s.ControlPanelPort))

	//************************************************
	// Actually setup the Network
	//************************************************

	// Make p.cnt Factom nodes
	for i := 0; i < p.Cnt; i++ {
		makeServer(s) // We clone s to make all of our servers
	}

	addFnodeName(0) // bootstrap id doesn't change

	// Modify Identities of new nodes
	if len(fnodes) > 1 && len(s.Prefix) == 0 {
		modifyLoadIdentities() // We clone s to make all of our servers
	}

	// Setup the Skeleton Identity & Registration
	for i := range fnodes {
		fnodes[i].State.IntiateNetworkSkeletonIdentity()
		fnodes[i].State.InitiateNetworkIdentityRegistration()
	}

	// Start the P2P network
	var networkID p2p.NetworkID
	var seedURL, networkPort, configPeers string
	switch s.Network {
	case "MAIN", "main":
		networkID = p2p.MainNet
		seedURL = s.MainSeedURL
		networkPort = s.MainNetworkPort
		configPeers = s.MainSpecialPeers
		s.DirectoryBlockInSeconds = 600
	case "TEST", "test":
		networkID = p2p.TestNet
		seedURL = s.TestSeedURL
		networkPort = s.TestNetworkPort
		configPeers = s.TestSpecialPeers
	case "LOCAL", "local":
		networkID = p2p.LocalNet
		seedURL = s.LocalSeedURL
		networkPort = s.LocalNetworkPort
		configPeers = s.LocalSpecialPeers

		// Also update the local constants for custom networks
		fmt.Println("Running on the local network, use local coinbase constants")
		constants.SetLocalCoinBaseConstants()
	case "CUSTOM", "custom":
		if bytes.Compare(p.CustomNet, []byte("\xe3\xb0\xc4\x42")) == 0 {
			panic("Please specify a custom network with -customnet=<something unique here>")
		}
		s.CustomNetworkID = p.CustomNet
		networkID = p2p.NewNetworkID(p.CustomNetName)
		for i := range fnodes {
			fnodes[i].State.CustomNetworkID = p.CustomNet
		}
		seedURL = s.CustomSeedURL
		networkPort = s.CustomNetworkPort
		configPeers = s.CustomSpecialPeers

		// Also update the coinbase constants for custom networks
		fmt.Println("Running on the custom network, use custom coinbase constants")
		constants.SetCustomCoinBaseConstants()
	default:
		panic("Invalid Network choice in Config File or command line. Choose MAIN, TEST, LOCAL, or CUSTOM")
	}

	// use special peers from command line
	if p.Peers != "" {
		configPeers = p.Peers
	}

	p2pconf.Network = networkID
	p2pconf.SeedURL = seedURL
	p2pconf.ListenPort = networkPort
	p2pconf.Special = configPeers

	connectionMetricsChannel := make(chan map[string]p2p.PeerMetrics, 50)
	p2pconf.ReadDeadline = time.Minute * 5
	p2pconf.WriteDeadline = time.Minute * 5

	if p.EnableNet {
		nodeName := fnodes[0].State.FactomNodeName
		if 0 < p.NetworkPortOverride {
			networkPort = fmt.Sprintf("%d", p.NetworkPortOverride)
		}

		if net, err := p2p.NewNetwork(p2pconf); err != nil {
			fmt.Println(err)
			panic("Unable to start p2p network")
		} else {
			network = net

			network.SetMetricsHook(func(pm map[string]p2p.PeerMetrics) {
				select {
				case connectionMetricsChannel <- pm:
				default:
				}
			})
			network.Run()
			fnodes[0].State.NetworkController = net

			p2pProxy = new(P2PProxy).Init(nodeName, "P2P Network").(*P2PProxy)
			p2pProxy.Network = network

			fnodes[0].Peers = append(fnodes[0].Peers, p2pProxy)
			p2pProxy.StartProxy()

			go networkHousekeeping() // This goroutine executes once a second to keep the proxy apprised of the network status.

		}

	}

	// Start live feed service
	config := s.Cfg.(*util.FactomdConfig)
	if config.LiveFeedAPI.EnableLiveFeedAPI || p.EnableLiveFeedAPI {
		s.EventService.ConfigService(s, config, p)
	}

	networkpattern = p.Net

	switch p.Net {
	case "file":
		file, err := os.Open(p.Fnet)
		if err != nil {
			panic(fmt.Sprintf("File network.txt failed to open: %s", err.Error()))
		} else if file == nil {
			panic(fmt.Sprint("File network.txt failed to open, and we got a file of <nil>"))
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var a, b int
			var s string
			fmt.Sscanf(scanner.Text(), "%d %s %d", &a, &s, &b)
			if s == "--" {
				AddSimPeer(fnodes, a, b)
			}
		}
	case "square":
		side := int(math.Sqrt(float64(p.Cnt)))

		for i := 0; i < side; i++ {
			AddSimPeer(fnodes, i*side, (i+1)*side-1)
			AddSimPeer(fnodes, i, side*(side-1)+i)
			for j := 0; j < side; j++ {
				if j < side-1 {
					AddSimPeer(fnodes, i*side+j, i*side+j+1)
				}
				AddSimPeer(fnodes, i*side+j, ((i+1)*side)+j)
			}
		}
	case "long":
		fmt.Println("Using long Network")
		for i := 1; i < p.Cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
		// Make long into a circle
	case "loops":
		fmt.Println("Using loops Network")
		for i := 1; i < p.Cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
		for i := 0; (i+17)*2 < p.Cnt; i += 17 {
			AddSimPeer(fnodes, i%p.Cnt, (i+5)%p.Cnt)
		}
		for i := 0; (i+13)*2 < p.Cnt; i += 13 {
			AddSimPeer(fnodes, i%p.Cnt, (i+7)%p.Cnt)
		}
	case "alot":
		n := len(fnodes)
		for i := 0; i < n; i++ {
			AddSimPeer(fnodes, i, (i+1)%n)
			AddSimPeer(fnodes, i, (i+5)%n)
			AddSimPeer(fnodes, i, (i+7)%n)
		}

	case "alot+":
		n := len(fnodes)
		for i := 0; i < n; i++ {
			AddSimPeer(fnodes, i, (i+1)%n)
			AddSimPeer(fnodes, i, (i+5)%n)
			AddSimPeer(fnodes, i, (i+7)%n)
			AddSimPeer(fnodes, i, (i+13)%n)
		}

	case "tree":
		index := 0
		row := 1
	treeloop:
		for i := 0; true; i++ {
			for j := 0; j <= i; j++ {
				AddSimPeer(fnodes, index, row)
				AddSimPeer(fnodes, index, row+1)
				row++
				index++
				if index >= len(fnodes) {
					break treeloop
				}
			}
			row += 1
		}
	case "circles":
		circleSize := 7
		index := 0
		for {
			AddSimPeer(fnodes, index, index+circleSize-1)
			for i := index; i < index+circleSize-1; i++ {
				AddSimPeer(fnodes, i, i+1)
			}
			index += circleSize

			AddSimPeer(fnodes, index, index-circleSize/3)
			AddSimPeer(fnodes, index+2, index-circleSize-circleSize*2/3-1)
			AddSimPeer(fnodes, index+3, index-(2*circleSize)-circleSize*2/3)
			AddSimPeer(fnodes, index+5, index-(3*circleSize)-circleSize*2/3+1)

			if index >= len(fnodes) {
				break
			}
		}
	default:
		fmt.Println("Didn't understand network type. Known types: mesh, long, circles, tree, loops.  Using a Long Network")
		for i := 1; i < p.Cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}

	}

	var colors []string = []string{"95cde5", "b01700", "db8e3c", "ffe35f"}

	if len(fnodes) > 2 {
		for i, s := range fnodes {
			fmt.Printf("%d {color:#%v, shape:dot, label:%v}\n", i, colors[i%len(colors)], s.State.FactomNodeName)
		}
		fmt.Printf("Paste the network info above into http://arborjs.org/halfviz to visualize the network\n")
	}
	// Initiate dbstate plugin if enabled. Only does so for first node,
	// any more nodes on sim control will use default method
	fnodes[0].State.SetTorrentUploader(p.TorUpload)
	if p.TorManage {
		fnodes[0].State.SetUseTorrent(true)
		manager, err := LaunchDBStateManagePlugin(p.PluginPath, fnodes[0].State.InMsgQueue(), fnodes[0].State, fnodes[0].State.GetServerPrivateKey(), p.MemProfileRate)
		if err != nil {
			panic("Encountered an error while trying to use torrent DBState manager: " + err.Error())
		}
		fnodes[0].State.DBStateManager = manager
	} else {
		fnodes[0].State.SetUseTorrent(false)
	}

	if p.Journal != "" {
		go LoadJournal(s, p.Journal)
		startServers(false)
	} else {
		startServers(true)
	}

	// Anchoring related configurations
	if len(config.App.BitcoinAnchorRecordPublicKeys) > 0 {
		err := s.GetDB().(*databaseOverlay.Overlay).SetBitcoinAnchorRecordPublicKeysFromHex(config.App.BitcoinAnchorRecordPublicKeys)
		if err != nil {
			panic("Encountered an error while trying to set custom Bitcoin anchor record keys from config")
		}
	}
	if len(config.App.EthereumAnchorRecordPublicKeys) > 0 {
		err := s.GetDB().(*databaseOverlay.Overlay).SetEthereumAnchorRecordPublicKeysFromHex(config.App.EthereumAnchorRecordPublicKeys)
		if err != nil {
			panic("Encountered an error while trying to set custom Ethereum anchor record keys from config")
		}
	}
	if p.ReparseAnchorChains {
		fmt.Println("Reparsing anchor chains...")
		err := fnodes[0].State.GetDB().(*databaseOverlay.Overlay).ReparseAnchorChains()
		if err != nil {
			panic("Encountered an error while trying to re-parse anchor chains: " + err.Error())
		}
	}

	// Start the webserver
	wsapi.Start(fnodes[0].State)
	if fnodes[0].State.DebugExec() && messages.CheckFileName("graphData.txt") {
		go printGraphData("graphData.txt", 30)
	}

	// Start prometheus on port
	launchPrometheus(9876)
	// Start Package's prometheus
	state.RegisterPrometheus()
	//	p2pold.RegisterPrometheus()
	leveldb.RegisterPrometheus()
	RegisterPrometheus()

	go controlPanel.ServeControlPanel(fnodes[0].State.ControlPanelChannel, fnodes[0].State, connectionMetricsChannel, network, Build, p.NodeName)

	go SimControl(p.ListenTo, listenToStdin)

}

func printGraphData(filename string, period int) {
	downscale := int64(1)
	messages.LogPrintf(filename, "\t%9s\t%9s\t%9s\t%9s\t%9s\t%9s", "Dbh-:-min", "Node", "ProcessCnt", "ListPCnt", "UpdateState", "SleepCnt")
	for {
		for _, f := range fnodes {
			s := f.State
			messages.LogPrintf(filename, "\t%9s\t%9s\t%9d\t%9d\t%9d\t%9d", fmt.Sprintf("%d-:-%d", s.LLeaderHeight, s.CurrentMinute), s.FactomNodeName, s.StateProcessCnt/downscale, s.ProcessListProcessCnt/downscale, s.StateUpdateState/downscale, s.ValidatorLoopSleepCnt/downscale)
		}
		time.Sleep(time.Duration(period) * time.Second)
	} // for ever ...
}

//**********************************************************************
// Functions that access variables in this method to set up Factom Nodes
// and start the servers.
//**********************************************************************
func makeServer(s *state.State) *FactomNode {
	// All other states are clones of the first state.  Which this routine
	// gets passed to it.
	newState := s

	if len(fnodes) > 0 {
		newState = s.Clone(len(fnodes)).(*state.State)
		newState.EFactory = new(electionMsgs.ElectionsFactory) // not an elegant place but before we let the messages hit the state
		time.Sleep(10 * time.Millisecond)
		newState.Init()
		newState.EFactory = new(electionMsgs.ElectionsFactory)
	}

	fnode := new(FactomNode)
	fnode.State = newState
	fnodes = append(fnodes, fnode)
	fnode.MLog = mLog

	return fnode
}

func startServers(load bool) {
	for i, fnode := range fnodes {
		startServer(i, fnode, load)
	}
}

func startServer(i int, fnode *FactomNode, load bool) {
	fnode.State.RunState = runstate.Booting
	if i > 0 {
		fnode.State.Init()
	}
	NetworkProcessorNet(fnode)
	if load {
		go state.LoadDatabase(fnode.State)
	}
	go fnode.State.GoSyncEntries()
	go Timer(fnode.State)
	go elections.Run(fnode.State)
	go fnode.State.ValidatorLoop()

	// moved StartMMR here to ensure Init goroutine only called once and not twice (removed from state.go)
	go fnode.State.StartMMR()
	go fnode.State.MissingMessageResponseHandler.Run()
}

func setupFirstAuthority(s *state.State) {
	if len(s.IdentityControl.Authorities) > 0 {
		//Don't initialize first authority if we are loading during fast boot
		//And there are already authorities present
		return
	}

	s.IdentityControl.SetBootstrapIdentity(s.GetNetworkBootStrapIdentity(), s.GetNetworkBootStrapKey())
}

func networkHousekeeping() {
	for {
		time.Sleep(1 * time.Second)
		p2pProxy.SetWeight(network.GetInfo().Peers)
	}
}

func AddNode() {

	fnodes := GetFnodes()
	s := fnodes[0].State
	i := len(fnodes)

	makeServer(s)
	modifyLoadIdentities()

	fnodes = GetFnodes()
	fnodes[i].State.IntiateNetworkSkeletonIdentity()
	fnodes[i].State.InitiateNetworkIdentityRegistration()
	AddSimPeer(fnodes, i, i-1) // KLUDGE peer w/ only last node
	startServer(i, fnodes[i], true)
}
