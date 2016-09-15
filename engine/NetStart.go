// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"flag"
	"fmt"
	"os"
	"time"

	"math"

	"bufio"
	"bytes"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
)

var _ = fmt.Print

type FactomNode struct {
	State *state.State
	Peers []interfaces.IPeer
	MLog  *MsgLog
}

var fnodes []*FactomNode
var mLog = new(MsgLog)
var p2pProxy *P2PProxy
var p2pNetwork *p2p.Controller

func NetStart(s *state.State) {
	enablenetPtr := flag.Bool("enablenet", true, "Enable or disable networking")
	listenToPtr := flag.Int("node", 0, "Node Number the simulator will set as the focus")
	cntPtr := flag.Int("count", 1, "The number of nodes to generate")
	netPtr := flag.String("net", "tree", "The default algorithm to build the network connections")
	fnetPtr := flag.String("fnet", "", "Read the given file to build the network connections")
	dropPtr := flag.Int("drop", 0, "Number of messages to drop out of every thousand")
	journalPtr := flag.String("journal", "", "Rerun a Journal of messages")
	journalingPtr := flag.Bool("journaling", false, "Write a journal of all messages recieved. Default is off.")
	followerPtr := flag.Bool("follower", false, "If true, force node to be a follower.  Only used when replaying a journal.")
	leaderPtr := flag.Bool("leader", true, "If true, force node to be a leader.  Only used when replaying a journal.")
	dbPtr := flag.String("db", "", "Override the Database in the Config file and use this Database implementation")
	cloneDBPtr := flag.String("clonedb", "", "Override the main node and use this database for the clones in a Network.")
	portOverridePtr := flag.Int("port", 0, "Address to serve WSAPI on")
	networkNamePtr := flag.String("network", "", "Network to join: MAIN, TEST or LOCAL")
	networkPortOverridePtr := flag.Int("networkPort", 0, "Address for p2p network to listen on.")
	peersPtr := flag.String("peers", "", "Array of peer addresses. ")
	blkTimePtr := flag.Int("blktime", 0, "Seconds per block.  Production is 600.")
	runtimeLogPtr := flag.Bool("runtimeLog", false, "If true, maintain runtime logs of messages passed.")
	netdebugPtr := flag.Int("netdebug", 0, "0-5: 0 = quiet, >0 = increasing levels of logging")
	exclusivePtr := flag.Bool("exclusive", false, "If true, we only dial out to special/trusted peers.")
	prefixNodePtr := flag.String("prefix", "", "Prefix the Factom Node Names with this value; used to create leaderless networks.")
	rotatePtr := flag.Bool("rotate", false, "If true, responsiblity is owned by one leader, and rotated over the leaders.")
	timeOffsetPtr := flag.Int("timedelta", 0, "Maximum timeDelta in milliseconds to offset each node.  Simulates deltas in system clocks over a network.")
	keepMismatchPtr := flag.Bool("keepmismatch", false, "If true, do not discard DBStates even when a majority of DBSignatures have a different hash")
	startDelayPtr := flag.Int("startdelay", 20, "Delay to start processing messages, in seconds")
	deadlinePtr := flag.Int("deadline", 10, "Timeout Delay in milliseconds used on Reads and Writes to the network comm")

	flag.Parse()

	enableNet := *enablenetPtr
	listenTo := *listenToPtr
	cnt := *cntPtr
	net := *netPtr
	fnet := *fnetPtr
	droprate := *dropPtr
	journal := *journalPtr
	journaling := *journalingPtr
	follower := *followerPtr
	leader := *leaderPtr
	db := *dbPtr
	cloneDB := *cloneDBPtr
	portOverride := *portOverridePtr
	peers := *peersPtr
	networkName := *networkNamePtr
	networkPortOverride := *networkPortOverridePtr
	blkTime := *blkTimePtr
	runtimeLog := *runtimeLogPtr
	netdebug := *netdebugPtr
	exclusive := *exclusivePtr
	prefix := *prefixNodePtr
	rotate := *rotatePtr
	timeOffset := *timeOffsetPtr
	keepMismatch := *keepMismatchPtr
	startDelay := int64(*startDelayPtr)
	deadline := *deadlinePtr

	// Must add the prefix before loading the configuration.
	s.AddPrefix(prefix)
	FactomConfigFilename := util.GetConfigFilename("m2")
	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))
	s.LoadConfig(FactomConfigFilename, networkName)
	s.OneLeader = rotate
	s.TimeOffset = primitives.NewTimestampFromMilliseconds(uint64(timeOffset))
	s.StartDelayLimit = startDelay * 1000
	s.Journaling = journaling

	if 999 < portOverride { // The command line flag exists and seems reasonable.
		s.SetPort(portOverride)
	}

	if blkTime != 0 {
		s.DirectoryBlockInSeconds = blkTime
	} else {
		blkTime = s.DirectoryBlockInSeconds
	}

	if follower {
		leader = false
	}
	if leader {
		follower = false
	}
	if !follower && !leader {
		panic("Not a leader or a follower")
	}

	if journal != "" {
		cnt = 1
	}

	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Net Sim Start!")
	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Listening to Node", listenTo)
	fmt.Println(">>>>>>>>>>>>>>>>")

	AddInterruptHandler(func() {
		fmt.Print("<Break>\n")
		fmt.Print("Gracefully shutting down the server...\n")
		for _, fnode := range fnodes {
			fmt.Print("Shutting Down: ", fnode.State.FactomNodeName, "\r\n")
			fnode.State.ShutdownChan <- 0
		}
		if enableNet {
			p2pNetwork.NetworkStop()
			// NODE_TALK_FIX
			p2pProxy.stopProxy()
		}
		fmt.Print("Waiting...\r\n")
		time.Sleep(3 * time.Second)
		os.Exit(0)
	})

	if journal != "" {
		if s.DBType != "Map" {
			fmt.Println("Journal is ALWAYS a Map database")
			s.DBType = "Map"
		}
	}
	if follower {
		s.NodeMode = "FULL"
		leadID := primitives.Sha([]byte(s.Prefix + "FNode0"))
		if s.IdentityChainID.IsSameAs(leadID) {
			s.SetIdentityChainID(primitives.Sha([]byte(time.Now().String()))) // Make sure this node is NOT a leader
		}
	}
	if leader {
		if len(s.Prefix) == 0 {
			s.SetIdentityChainID(primitives.Sha([]byte(s.Prefix + "FNode0"))) // Make sure this node is a leader
			s.NodeMode = "SERVER"
		}
	}

	s.KeepMismatch = keepMismatch

	if len(db) > 0 {
		s.DBType = db
	} else {
		db = s.DBType
	}

	if len(cloneDB) > 0 {
		s.CloneDBType = cloneDB
	} else {
		s.CloneDBType = db
	}

	pnet := net
	if len(fnet) > 0 {
		pnet = fnet
		net = "file"
	}

	go StartProfiler()

	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "enablenet", enableNet))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "node", listenTo))
	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "prefix", prefix))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "node count", cnt))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "net spec", pnet))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "Msgs droped", droprate))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "journal", journal))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database", db))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database for clones", cloneDB))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "port", s.PortNumber))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "peers", peers))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "netdebug", netdebug))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%t\"\n", "exclusive", exclusive))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "block time", blkTime))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "runtimeLog", runtimeLog))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "rotate", rotate))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "timeOffset", timeOffset))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "keepMismatch", keepMismatch))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "startDelay", startDelay))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "Network", s.Network))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "deadline (ms)", deadline))

	s.AddPrefix(prefix)
	s.SetOut(false)
	s.Init()
	s.SetDropRate(droprate)

	mLog.init(runtimeLog, cnt)

	setupFirstAuthority(s)

	//************************************************
	// Actually setup the Network
	//************************************************

	// Make cnt Factom nodes
	for i := 0; i < cnt; i++ {
		makeServer(s) // We clone s to make all of our servers
	}
	// Modify Identities of new nodes
	if len(fnodes) > 1 && len(s.Prefix) == 0 {
		modifyLoadIdentities() // We clone s to make all of our servers
	}

	// Sort the FNodes by ID
	for i := 0; i < len(fnodes)-1; i++ {
		for j := 0; j < len(fnodes)-1-i; j++ {
			if bytes.Compare(fnodes[j].State.IdentityChainID.Bytes(), fnodes[j+1].State.IdentityChainID.Bytes()) > 0 {
				tmp := fnodes[j]
				fnodes[j] = fnodes[j+1]
				fnodes[j+1] = tmp
			}
		}
	}
	for i := 0; i < len(fnodes); i++ {
		fnodes[i].State.FactomNodeName = fmt.Sprintf("FNode%d", i)
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
		networkID = p2p.LocalNet
		seedURL = s.LocalSeedURL
		networkPort = s.LocalNetworkPort
		specialPeers = s.LocalSpecialPeers
	default:
		panic("Invalid Network choice in Config File. Choose MAIN, TEST or LOCAL")
	}

	connectionMetricsChannel := make(chan interface{}, p2p.StandardChannelSize)
	p2p.Deadline = time.Duration(deadline) * time.Millisecond

	if enableNet {

		if 0 < networkPortOverride {
			networkPort = fmt.Sprintf("%d", networkPortOverride)
		}
		ci := p2p.ControllerInit{
			Port:                     networkPort,
			PeersFile:                s.PeersFile,
			Network:                  networkID,
			Exclusive:                exclusive,
			SeedURL:                  seedURL,
			SpecialPeers:             specialPeers,
			ConnectionMetricsChannel: connectionMetricsChannel,
		}
		p2pNetwork = new(p2p.Controller).Init(ci)
		p2pNetwork.StartNetwork()
		// Setup the proxy (Which translates from network parcels to factom messages, handling addressing for directed messages)
		p2pProxy = new(P2PProxy).Init(fnodes[0].State.FactomNodeName, "P2P Network").(*P2PProxy)
		p2pProxy.FromNetwork = p2pNetwork.FromNetwork
		p2pProxy.ToNetwork = p2pNetwork.ToNetwork
		fnodes[0].Peers = append(fnodes[0].Peers, p2pProxy)
		p2pProxy.SetDebugMode(netdebug)
		if 0 < netdebug {
			go p2pProxy.PeriodicStatusReport(fnodes)
			p2pNetwork.StartLogging(uint8(netdebug))
		} else {
			p2pNetwork.StartLogging(uint8(0))
		}
		p2pProxy.StartProxy()
		// Command line peers lets us manually set special peers
		p2pNetwork.DialSpecialPeersString(peers)
	}

	switch net {
	case "file":
		file, err := os.Open(fnet)
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
		side := int(math.Sqrt(float64(cnt)))

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
		for i := 1; i < cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
		// Make long into a circle
		AddSimPeer(fnodes, 0, cnt-1)
	case "loops":
		fmt.Println("Using loops Network")
		for i := 1; i < cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
		for i := 0; (i+17)*2 < cnt; i += 17 {
			AddSimPeer(fnodes, i%cnt, (i+5)%cnt)
		}
		for i := 0; (i+13)*2 < cnt; i += 13 {
			AddSimPeer(fnodes, i%cnt, (i+7)%cnt)
		}
	case "alot":
		n := len(fnodes)
		for i := 0; i < n; i++ {
			AddSimPeer(fnodes, i, (i+1)%n)
			AddSimPeer(fnodes, i, (i+5)%n)
			AddSimPeer(fnodes, i, (i+7)%n)
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
		for i := 1; i < cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}

	}
	if journal != "" {
		go LoadJournal(s, journal)
		startServers(false)
	} else {
		startServers(true)
	}

	// Start the webserver
	go wsapi.Start(fnodes[0].State)

	go controlPanel.ServeControlPanel(fnodes[0].State.ControlPanelChannel, fnodes[0].State, connectionMetricsChannel, p2pNetwork, Build)
	// Listen for commands:
	SimControl(listenTo)
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
		number := fmt.Sprintf("%d", len(fnodes))
		newState = s.Clone(number).(*state.State)
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
		go Timer(fnode.State)
		go fnode.State.ValidatorLoop()
	}
}

func setupFirstAuthority(s *state.State) {
	var id state.Identity
	id.IdentityChainID, _ = primitives.HexToHash("38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9") //s.IdentityChainID
	id.ManagementChainID, _ = primitives.HexToHash("88888800000000000000000000000000")
	pub := primitives.PubKeyFromString("cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a")
	data, _ := pub.MarshalBinary()
	id.SigningKey = primitives.NewHash(data)
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
	s.Identities = append(s.Identities, id)

	var auth state.Authority
	auth.Status = 1
	auth.SigningKey = primitives.PubKeyFromString("cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a")
	auth.MatryoshkaHash = primitives.NewZeroHash()
	auth.AuthorityChainID, _ = primitives.HexToHash("38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9") //s.IdentityChainID
	auth.ManagementChainID, _ = primitives.HexToHash("88888800000000000000000000000000")
	s.Authorities = append(s.Authorities, auth)
}
