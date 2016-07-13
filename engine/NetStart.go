// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"math"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
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
var network p2p.Controller
var p2pProxy *P2PProxy

func NetStart(s *state.State) {

	listenToPtr := flag.Int("node", 0, "Node Number the simulator will set as the focus")
	cntPtr := flag.Int("count", 1, "The number of nodes to generate")
	netPtr := flag.String("net", "tree", "The default algorithm to build the network connections")
	dropPtr := flag.Int("drop", 0, "Number of messages to drop out of every thousand")
	journalPtr := flag.String("journal", "", "Rerun a Journal of messages")
	followerPtr := flag.Bool("follower", false, "If true, force node to be a follower.  Only used when replaying a journal.")
	leaderPtr := flag.Bool("leader", true, "If true, force node to be a leader.  Only used when replaying a journal.")
	dbPtr := flag.String("db", "", "Override the Database in the Config file and use this Database implementation")
	cloneDBPtr := flag.String("clonedb", "", "Override the main node and use this database for the clones in a Network.")
	folderPtr := flag.String("folder", "", "Directory in .factom to store nodes. (eg: multiple nodes on one filesystem support)")
	portOverridePtr := flag.Int("port", 0, "Address to serve WSAPI on")
	networkPortPtr := flag.String("p2pPort", "8108", "Port to listen for peers on.")
	peersPtr := flag.String("peers", "", "Array of peer addresses. ")
	blkTimePtr := flag.Int("blktime", 0, "Seconds per block.  Production is 600.")
	runtimeLogPtr := flag.Bool("runtimeLog", false, "If true, maintain runtime logs of messages passed.")
	netdebugPtr := flag.Int("netdebug", 0, "0-5: 0 = quiet, >0 = increasing levels of logging")
	exclusivePtr := flag.Bool("exclusive", false, "If true, we only dial out to special/trusted peers.")
	prefixNodePtr := flag.String("prefix", "", "Prefix the Factom Node Names with this value; used to create leaderless networks.")
	rotatePtr := flag.Bool("rotate", false, "If true, responsiblity is owned by one leader, and rotated over the leaders.")
	timeOffsetPtr := flag.Int("timedelta", 0, "Maximum timeDelta in milliseconds to offset each node.  Simulates deltas in system clocks over a network.")
	keepMismatchPtr := flag.Bool("keepmismatch", false, "If true, do not discard DBStates even when a majority of DBSignatures have a different hash")

	flag.Parse()

	listenTo := *listenToPtr
	cnt := *cntPtr
	net := *netPtr
	droprate := *dropPtr
	journal := *journalPtr
	follower := *followerPtr
	leader := *leaderPtr
	db := *dbPtr
	cloneDB := *cloneDBPtr
	folder := *folderPtr
	portOverride := *portOverridePtr
	networkPort := *networkPortPtr
	peers := *peersPtr
	blkTime := *blkTimePtr
	runtimeLog := *runtimeLogPtr
	netdebug := *netdebugPtr
	exclusive := *exclusivePtr
	prefix := *prefixNodePtr
	rotate := *rotatePtr
	timeOffset := *timeOffsetPtr
	keepMismatch := *keepMismatchPtr

	// Must add the prefix before loading the configuration.
	s.AddPrefix(prefix)
	FactomConfigFilename := util.GetConfigFilename("m2")
	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))
	s.LoadConfig(FactomConfigFilename, folder)

	s.OneLeader = rotate
	s.TimeOffset = primitives.NewTimestampFromMilliseconds(uint64(timeOffset))

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
		network.NetworkStop()
		// NODE_TALK_FIX
		p2pProxy.stopProxy()
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
		s.SetIdentityChainID(primitives.Sha([]byte(time.Now().String()))) // Make sure this node is NOT a leader
	}
	if leader {
		s.SetIdentityChainID(primitives.Sha([]byte(s.Prefix + "FNode0"))) // Make sure this node is a leader
		s.NodeMode = "SERVER"
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

	go StartProfiler()

	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "node", listenTo))
	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "prefix", prefix))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "node count", cnt))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "net type", net))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "Msgs droped", droprate))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "journal", journal))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database", db))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database for clones", cloneDB))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "folder", folder))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "port", s.PortNumber))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "networkPort", networkPort))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "peers", peers))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%d\"\n", "netdebug", netdebug))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%t\"\n", "exclusive", exclusive))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "block time", blkTime))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "runtimeLog", runtimeLog))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "rotate", rotate))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "timeOffset", timeOffset))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "keepMismatch", keepMismatch))

	s.AddPrefix(prefix)
	s.SetOut(false)
	s.Init()
	s.SetDropRate(droprate)

	mLog.init(runtimeLog, cnt)

	setupBlankAuthority(s)

	//************************************************
	// Actually setup the Network
	//************************************************

	// Make cnt Factom nodes
	for i := 0; i < cnt; i++ {
		makeServer(s) // We clone s to make all of our servers
	}

	// Start the P2P netowork
	var networkID p2p.NetworkID
	switch s.Network {
	case "MAIN", "main":
		networkID = p2p.MainNet
	case "LOCAL", "local":
		networkID = p2p.LocalNet
	case "TEST", "test":
		networkID = p2p.TestNet
	default:
		panic("Invalid Network choice in Config File. Choose MAIN, TEST or LOCAL")
	}
	ci := p2p.ControllerInit{
		Port:      networkPort,
		PeersFile: s.PeersFile,
		Network:   networkID,
		Exclusive: exclusive,
		SeedURL:   s.SeedURL,
	}
	p2pController := new(p2p.Controller).Init(ci)
	network = *p2pController
	network.StartNetwork()
	// Setup the proxy (Which translates from network parcels to factom messages, handling addressing for directed messages)
	p2pProxy = new(P2PProxy).Init(fnodes[0].State.FactomNodeName, "P2P Network").(*P2PProxy)
	p2pProxy.FromNetwork = network.FromNetwork
	p2pProxy.ToNetwork = network.ToNetwork
	fnodes[0].Peers = append(fnodes[0].Peers, p2pProxy)
	p2pProxy.SetDebugMode(netdebug)
	if 0 < netdebug {
		go PeriodicStatusReport(fnodes)
		go p2pProxy.ProxyStatusReport(fnodes)
		network.StartLogging(uint8(netdebug))
	} else {
		network.StartLogging(uint8(0))
	}
	p2pProxy.startProxy()
	// Command line peers lets us manually set special peers
	// Parse the peers into an array.
	parseFunc := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
	}
	peerAddresses := strings.FieldsFunc(peers, parseFunc)
	for _, peerAddress := range peerAddresses {
		fmt.Println("Dialing Peer: ", peerAddress)
		ipPort := strings.Split(peerAddress, ":")
		peer := new(p2p.Peer).Init(ipPort[0], ipPort[1], 0, p2p.SpecialPeer, 0)
		peer.Source["Command Line"] = time.Now()
		network.DialPeer(*peer, true) // these are persistent connections
	}

	switch net {
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

func setupBlankAuthority(s *state.State) {
	var id state.Identity
	id.IdentityChainID, _ = primitives.HexToHash("38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9") //s.IdentityChainID
	id.ManagementChainID, _ = primitives.HexToHash("88888800000000000000000000000000")
	//fmt.Printf("DEBUG: State Public: %x\n", s.GetServerPublicKey())
	//fmt.Printf("DEBUG: State Private: %x\n", *(s.GetServerPrivateKey().Key))
	pub := s.GetServerPublicKey() // primitives.PubKeyFromString("0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a")
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
	auth.SigningKey = s.GetServerPublicKey() //primitives.PubKeyFromString("0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a")
	auth.MatryoshkaHash = primitives.NewZeroHash()
	auth.AuthorityChainID, _ = primitives.HexToHash("38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9") //s.IdentityChainID
	auth.ManagementChainID, _ = primitives.HexToHash("88888800000000000000000000000000")
	s.Authorities = append(s.Authorities, auth)
}
