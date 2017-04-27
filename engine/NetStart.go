// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"time"

	"math"

	"bufio"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
)

var _ = fmt.Print

type FactomNode struct {
	Index int
	State *state.State
	Peers []interfaces.IPeer
	MLog  *MsgLog
}

var fnodes []*FactomNode
var mLog = new(MsgLog)
var p2pProxy *P2PProxy
var p2pNetwork *p2p.Controller
var logPort string

func GetFnodes() []*FactomNode {
	return fnodes
}

type FactomParams struct {
	ackbalanceHash           bool
	enableNet                bool
	waitEntries              bool
	listenTo                 int
	cnt                      int
	net                      string
	fnet                     string
	droprate                 int
	journal                  string
	journaling               bool
	follower                 bool
	leader                   bool
	db                       string
	cloneDB                  string
	portOverride             int
	peers                    string
	networkName              string
	networkPortOverride      int
	ControlPanelPortOverride int
	logPort                  string
	blkTime                  int
	faultTimeout             int
	runtimeLog               bool
	netdebug                 int
	exclusive                bool
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

	logportPtr := flag.String("logPort", "6060", "Port for pprof logging")
	portOverridePtr := flag.Int("port", 0, "Port where we serve WSAPI;  default 8088")
	ControlPanelPortOverridePtr := flag.Int("ControlPanelPort", 0, "Port for control panel webserver;  Default 8090")
	networkPortOverridePtr := flag.Int("networkPort", 0, "Port for p2p network; default 8110")

	flag.CommandLine.Parse(args)

	p.ackbalanceHash = *ackBalanceHashPtr
	p.enableNet = *enablenetPtr
	p.waitEntries = *waitEntriesPtr
	p.listenTo = *listenToPtr
	p.cnt = *cntPtr
	p.net = *netPtr
	p.fnet = *fnetPtr
	p.droprate = *dropPtr
	p.journal = *journalPtr
	p.journaling = *journalingPtr
	p.follower = *followerPtr
	p.leader = *leaderPtr
	p.db = *dbPtr
	p.cloneDB = *cloneDBPtr
	p.portOverride = *portOverridePtr
	p.peers = *peersPtr
	p.networkName = *networkNamePtr
	p.networkPortOverride = *networkPortOverridePtr
	p.ControlPanelPortOverride = *ControlPanelPortOverridePtr
	p.logPort = *logportPtr
	p.blkTime = *blkTimePtr
	p.faultTimeout = *faultTimeoutPtr
	p.runtimeLog = *runtimeLogPtr
	p.netdebug = *netdebugPtr
	p.exclusive = *exclusivePtr
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
	return p
}

func NetStart(s *state.State, p *FactomParams, listenToStdin bool) {

	s.PortNumber = 8088
	s.ControlPanelPort = 8090

	messages.AckBalanceHash = p.ackbalanceHash
	// Must add the prefix before loading the configuration.
	s.AddPrefix(p.prefix)
	FactomConfigFilename := util.GetConfigFilename("m2")
	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))
	s.LoadConfig(FactomConfigFilename, p.networkName)
	s.OneLeader = p.rotate
	s.TimeOffset = primitives.NewTimestampFromMilliseconds(uint64(p.timeOffset))
	s.StartDelayLimit = p.startDelay * 1000
	s.Journaling = p.journaling

	// Set the wait for entries flag
	s.WaitForEntries = p.waitEntries

	if 999 < p.portOverride { // The command line flag exists and seems reasonable.
		s.SetPort(p.portOverride)
	} else {
		p.portOverride = s.GetPort()
	}
	if 999 < p.ControlPanelPortOverride { // The command line flag exists and seems reasonable.
		s.ControlPanelPort = p.ControlPanelPortOverride
	} else {
		p.ControlPanelPortOverride = s.ControlPanelPort
	}

	if p.blkTime > 0 {
		s.DirectoryBlockInSeconds = p.blkTime
	} else {
		p.blkTime = s.DirectoryBlockInSeconds
	}

	s.FaultTimeout = p.faultTimeout

	if p.follower {
		p.leader = false
	}
	if p.leader {
		p.follower = false
	}
	if !p.follower && !p.leader {
		panic("Not a leader or a follower")
	}

	if p.journal != "" {
		p.cnt = 1
	}

	if p.rpcUser != "" {
		s.RpcUser = p.rpcUser
	}

	if p.rpcPassword != "" {
		s.RpcPass = p.rpcPassword
	}

	if p.factomdTLS == true {
		s.FactomdTLSEnable = true
	}

	if p.factomdLocations != "" {
		if len(s.FactomdLocations) > 0 {
			s.FactomdLocations += ","
		}
		s.FactomdLocations += p.factomdLocations
	}

	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Net Sim Start!")
	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Listening to Node", p.listenTo)
	fmt.Println(">>>>>>>>>>>>>>>>")

	AddInterruptHandler(func() {
		fmt.Print("<Break>\n")
		fmt.Print("Gracefully shutting down the server...\n")
		for _, fnode := range fnodes {
			fmt.Print("Shutting Down: ", fnode.State.FactomNodeName, "\r\n")
			fnode.State.ShutdownChan <- 0
		}
		if p.enableNet {
			p2pNetwork.NetworkStop()
			// NODE_TALK_FIX
			p2pProxy.stopProxy()
		}
		fmt.Print("Waiting...\r\n")
		time.Sleep(3 * time.Second)
		os.Exit(0)
	})

	if p.journal != "" {
		if s.DBType != "Map" {
			fmt.Println("Journal is ALWAYS a Map database")
			s.DBType = "Map"
		}
	}
	if p.follower {
		s.NodeMode = "FULL"
		leadID := primitives.Sha([]byte(s.Prefix + "FNode0"))
		if s.IdentityChainID.IsSameAs(leadID) {
			s.SetIdentityChainID(primitives.Sha([]byte(time.Now().String()))) // Make sure this node is NOT a leader
		}
	}

	s.KeepMismatch = p.keepMismatch

	if len(p.db) > 0 {
		s.DBType = p.db
	} else {
		p.db = s.DBType
	}

	if len(p.cloneDB) > 0 {
		s.CloneDBType = p.cloneDB
	} else {
		s.CloneDBType = p.db
	}

	pnet := p.net
	if len(p.fnet) > 0 {
		pnet = p.fnet
		p.net = "file"
	}

	go StartProfiler(p.memProfileRate)
	go StartProfiler(p.memProfileRate)

	s.AddPrefix(p.prefix)
	s.SetOut(false)
	s.Init()
	s.SetDropRate(p.droprate)

	mLog.init(p.runtimeLog, p.cnt)

	setupFirstAuthority(s)

	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "Build", Build))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "balancehash", messages.AckBalanceHash))
	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "FNode 0 Salt", s.Salt.String()[:16]))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "enablenet", p.enableNet))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "waitentries", p.waitEntries))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "node", p.listenTo))
	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "prefix", p.prefix))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "node count", p.cnt))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "net spec", pnet))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "Msgs droped", p.droprate))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "journal", p.journal))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database", p.db))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database for clones", p.cloneDB))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "peers", p.peers))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "netdebug", p.netdebug))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%t\"\n", "exclusive", p.exclusive))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "block time", p.blkTime))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "faultTimeout", p.faultTimeout))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "runtimeLog", p.runtimeLog))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "rotate", p.rotate))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "timeOffset", p.timeOffset))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "keepMismatch", p.keepMismatch))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "startDelay", p.startDelay))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "Network", s.Network))
	os.Stderr.WriteString(fmt.Sprintf("%20s %x\n", "customnet", p.customNet))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "deadline (ms)", p.deadline))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "tls", s.FactomdTLSEnable))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "selfaddr", s.FactomdLocations))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "rpcuser", s.RpcUser))
	if "" == s.RpcPass {
		os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "rpcpass", "is blank"))
	} else {
		os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "rpcpass", "is set"))
	}
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "TCP port", s.PortNumber))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "pprof port", logPort))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "Control Panel port", s.ControlPanelPort))

	//************************************************
	// Actually setup the Network
	//************************************************

	// Make p.cnt Factom nodes
	for i := 0; i < p.cnt; i++ {
		makeServer(s) // We clone s to make all of our servers
	}
	// Modify Identities of new nodes
	if len(fnodes) > 1 && len(s.Prefix) == 0 {
		modifyLoadIdentities() // We clone s to make all of our servers
	}

	// Setup the Skeleton Identity
	for i := range fnodes {
		fnodes[i].State.IntiateNetworkSkeletonIdentity()
	}

	// Start the P2P netowork
	var networkID p2p.NetworkID
	var seedURL, networkPort, specialPeers string
	switch s.Network {
	case "MAIN", "main":
		networkID = p2p.MainNet
		seedURL = s.MainSeedURL
		networkPort = s.MainNetworkPort
		specialPeers = s.MainSpecialPeers
	case "TEST", "test":
		networkID = p2p.TestNet
		seedURL = s.TestSeedURL
		networkPort = s.TestNetworkPort
		specialPeers = s.TestSpecialPeers
	case "LOCAL", "local":
		networkID = p2p.LocalNet
		seedURL = s.LocalSeedURL
		networkPort = s.LocalNetworkPort
		specialPeers = s.LocalSpecialPeers
	case "CUSTOM", "custom":
		if bytes.Compare(p.customNet, []byte("\xe3\xb0\xc4\x42")) == 0 {
			panic("Please specify a custom network with -customnet=<something unique here>")
		}
		s.CustomNetworkID = p.customNet
		networkID = p2p.NetworkID(binary.BigEndian.Uint32(p.customNet))
		for i := range fnodes {
			fnodes[i].State.CustomNetworkID = p.customNet
		}
		seedURL = s.LocalSeedURL
		networkPort = s.LocalNetworkPort
		specialPeers = s.LocalSpecialPeers
	default:
		panic("Invalid Network choice in Config File or command line. Choose MAIN, TEST, LOCAL, or CUSTOM")
	}

	connectionMetricsChannel := make(chan interface{}, p2p.StandardChannelSize)
	p2p.NetworkDeadline = time.Duration(p.deadline) * time.Millisecond

	if p.enableNet {
		if 0 < p.networkPortOverride {
			networkPort = fmt.Sprintf("%d", p.networkPortOverride)
		}
		ci := p2p.ControllerInit{
			Port:                     networkPort,
			PeersFile:                s.PeersFile,
			Network:                  networkID,
			Exclusive:                p.exclusive,
			SeedURL:                  seedURL,
			SpecialPeers:             specialPeers,
			ConnectionMetricsChannel: connectionMetricsChannel,
		}
		p2pNetwork = new(p2p.Controller).Init(ci)
		fnodes[0].State.NetworkControler = p2pNetwork
		p2pNetwork.StartNetwork()
		// Setup the proxy (Which translates from network parcels to factom messages, handling addressing for directed messages)
		p2pProxy = new(P2PProxy).Init(fnodes[0].State.FactomNodeName, "P2P Network").(*P2PProxy)
		p2pProxy.FromNetwork = p2pNetwork.FromNetwork
		p2pProxy.ToNetwork = p2pNetwork.ToNetwork
		fnodes[0].Peers = append(fnodes[0].Peers, p2pProxy)
		p2pProxy.SetDebugMode(p.netdebug)
		if 0 < p.netdebug {
			go p2pProxy.PeriodicStatusReport(fnodes)
			p2pNetwork.StartLogging(uint8(p.netdebug))
		} else {
			p2pNetwork.StartLogging(uint8(0))
		}
		p2pProxy.StartProxy()
		// Command line peers lets us manually set special peers
		p2pNetwork.DialSpecialPeersString(p.peers)
		go networkHousekeeping() // This goroutine executes once a second to keep the proxy apprised of the network status.
	}

	switch p.net {
	case "file":
		file, err := os.Open(p.fnet)
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
		side := int(math.Sqrt(float64(p.cnt)))

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
		for i := 1; i < p.cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
		// Make long into a circle
	case "loops":
		fmt.Println("Using loops Network")
		for i := 1; i < p.cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
		for i := 0; (i+17)*2 < p.cnt; i += 17 {
			AddSimPeer(fnodes, i%p.cnt, (i+5)%p.cnt)
		}
		for i := 0; (i+13)*2 < p.cnt; i += 13 {
			AddSimPeer(fnodes, i%p.cnt, (i+7)%p.cnt)
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
		for i := 1; i < p.cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}

	}
	if p.journal != "" {
		go LoadJournal(s, p.journal)
		startServers(false)
	} else {
		startServers(true)
	}

	// Start the webserver
	go wsapi.Start(fnodes[0].State)

	// Start prometheus on port
	launchPrometheus(9876)
	// Start Package's prometheus
	state.RegisterPrometheus()
	p2p.RegisterPrometheus()
	leveldb.RegisterPrometheus()

	go controlPanel.ServeControlPanel(fnodes[0].State.ControlPanelChannel, fnodes[0].State, connectionMetricsChannel, p2pNetwork, Build)
	// Listen for commands:
	SimControl(p.listenTo, listenToStdin)
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
		time.Sleep(10 * time.Millisecond)
		newState.Init()
	}

	fnode := new(FactomNode)
	fnode.State = newState
	fnodes = append(fnodes, fnode)
	fnode.MLog = mLog

	return fnode
}

func startServers(load bool) {
	for i, fnode := range fnodes {
		if i > 0 {
			fnode.State.Init()
		}
		go NetworkProcessorNet(fnode)
		if load {
			go state.LoadDatabase(fnode.State)
		}
		go fnode.State.GoSyncEntries()
		go Timer(fnode.State)
		go fnode.State.ValidatorLoop()
	}
}

func setupFirstAuthority(s *state.State) {
	var id state.Identity
	if networkIdentity := s.GetNetworkBootStrapIdentity(); networkIdentity != nil {
		id.IdentityChainID = networkIdentity
	} else {
		id.IdentityChainID = primitives.NewZeroHash()
	}
	id.ManagementChainID, _ = primitives.HexToHash("88888800000000000000000000000000")
	if pub := s.GetNetworkBootStrapKey(); pub != nil {
		id.SigningKey = pub
	} else {
		id.SigningKey = primitives.NewZeroHash()
	}
	id.MatryoshkaHash = primitives.NewZeroHash()
	id.ManagementCreated = 0
	id.ManagementRegistered = 0
	id.IdentityCreated = 0
	id.IdentityRegistered = 0
	id.Key1 = primitives.NewZeroHash()
	id.Key2 = primitives.NewZeroHash()
	id.Key3 = primitives.NewZeroHash()
	id.Key4 = primitives.NewZeroHash()
	id.Status = 1
	s.Identities = append(s.Identities, &id)

	var auth state.Authority
	auth.Status = 1
	auth.SigningKey = primitives.PubKeyFromString(id.SigningKey.String())
	auth.MatryoshkaHash = primitives.NewZeroHash()
	auth.AuthorityChainID = id.IdentityChainID
	auth.ManagementChainID, _ = primitives.HexToHash("88888800000000000000000000000000")
	s.Authorities = append(s.Authorities, &auth)
}

func networkHousekeeping() {
	for {
		time.Sleep(1 * time.Second)
		p2pProxy.SetWeight(p2pNetwork.GetNumberConnections())
	}
}
