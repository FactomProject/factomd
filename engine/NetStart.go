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
	addressPtr := flag.String("p2pPort", "8108", "Address & port to listen for peers on.")
	peersPtr := flag.String("peers", "", "Array of peer addresses. ")
	blkTimePtr := flag.Int("blktime", 0, "Seconds per block.  Production is 600.")
	runtimeLogPtr := flag.Bool("runtimeLog", true, "If true, maintain runtime logs of messages passed.")
	netdebugPtr := flag.Int("netdebug", 0, "0-5: 0 = quiet, >0 = increasing levels of logging")
	heartbeatPtr := flag.Bool("heartbeat", false, "If true, network just sends heartbeats.")
	prefixNodePtr := flag.String("prefix", "", "Prefix the Factom Node Names with this value; used to create leaderless networks.")
	profilePtr := flag.String("profile", "", "If true, turn on the go Profiler to profile execution of Factomd")

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
	address := *addressPtr
	peers := *peersPtr
	blkTime := *blkTimePtr
	runtimeLog := *runtimeLogPtr
	netdebug := *netdebugPtr
	heartbeat := *heartbeatPtr
	prefix := *prefixNodePtr
	profile := *profilePtr

	// Must add the prefix before loading the configuration.
	s.AddPrefix(prefix)
	FactomConfigFilename := util.GetConfigFilename("m2")
	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))
	s.LoadConfig(FactomConfigFilename, folder)

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

	if profile == "true" {
		go StartProfiler()
	} else {
		profile = "false"
	}

	os.Stderr.WriteString(fmt.Sprintf("node        %d\n", listenTo))
	os.Stderr.WriteString(fmt.Sprintf("prefix      %s\n", prefix))
	os.Stderr.WriteString(fmt.Sprintf("count       %d\n", cnt))
	os.Stderr.WriteString(fmt.Sprintf("net         \"%s\"\n", net))
	os.Stderr.WriteString(fmt.Sprintf("drop        %d\n", droprate))
	os.Stderr.WriteString(fmt.Sprintf("journal     \"%s\"\n", journal))
	os.Stderr.WriteString(fmt.Sprintf("db          \"%s\"\n", db))
	os.Stderr.WriteString(fmt.Sprintf("clonedb     \"%s\"\n", cloneDB))
	os.Stderr.WriteString(fmt.Sprintf("folder      \"%s\"\n", folder))
	os.Stderr.WriteString(fmt.Sprintf("port        \"%d\"\n", s.PortNumber))
	os.Stderr.WriteString(fmt.Sprintf("address     \"%s\"\n", address))
	os.Stderr.WriteString(fmt.Sprintf("peers       \"%s\"\n", peers))
	os.Stderr.WriteString(fmt.Sprintf("blkTime     %d\n", blkTime))
	os.Stderr.WriteString(fmt.Sprintf("runtimeLog  %v\n", runtimeLog))
	os.Stderr.WriteString(fmt.Sprintf("profile     %v\n", profile))

	s.AddPrefix(prefix)
	s.SetOut(false)
	s.Init()
	s.SetDropRate(droprate)

	mLog.init(runtimeLog, cnt)

	//************************************************
	// Actually setup the Network
	//************************************************

	// Make cnt Factom nodes
	for i := 0; i < cnt; i++ {
		makeServer(s) // We clone s to make all of our servers
	}

	// Start the P2P netowrk

	// BUGBUG Get peers file from config
	p2p := new(p2p.Controller).Init(address, "~/.factom/peers.json")
	network = *p2p
	network.StartNetwork(false) //BUGBUG This should be command line flag? Talk to Brian
	// Setup the proxy (Which translates from network parcels to factom messages, handling addressing for directed messages)
	p2pProxy := new(P2PProxy).Init(fnodes[0].State.FactomNodeName, "P2P Network").(*P2PProxy)
	p2pProxy.FromNetwork = network.FromNetwork
	p2pProxy.ToNetwork = network.ToNetwork
	fnodes[0].Peers = append(fnodes[0].Peers, p2pProxy)
	p2pProxy.SetDebugMode(netdebug)
	p2pProxy.SetTestMode(heartbeat)
	if 0 < netdebug {
		go PeriodicStatusReport(fnodes)
		go p2pProxy.ProxyStatusReport()
		network.StartLogging(uint8(netdebug))
	} else {
		network.StartLogging(uint8(0))
	}
	p2pProxy.startProxy()
	// Bootstrap peers (will be obsolete when discovery is finished)
	// Parse the peers into an array.
	parseFunc := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
	}
	peerAddresses := strings.FieldsFunc(peers, parseFunc)
	for _, peer := range peerAddresses {
		fmt.Println("Dialing Peer: ", peer)
		network.DialPeer(peer)
	}

	switch net {
	case "long":
		fmt.Println("Using long Network")
		for i := 1; i < cnt; i++ {
			AddSimPeer(fnodes, i-1, i)
		}
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
			AddSimPeer(fnodes, i, (i+3)%n)
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
