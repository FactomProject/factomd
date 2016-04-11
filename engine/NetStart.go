// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
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

// Enum for node types
const ( // iota is reset to 0
	cSimStyle      = iota
	cTCPStyle      = iota
	cEthereumStyle = iota
)

var nodeStyle = cSimStyle

//
// Multi-node test, example parameters:
// factomd -count=5 -net=tree -folder=testdir1 -serve=9000 -connect="tcp://217.0.0.1:9500"
//

func NetStart(s *state.State) {

	listenToPtr := flag.Int("node", 0, "Node Number the simulator will set as the focus")
	cntPtr := flag.Int("count", 1, "The number of nodes to generate")
	netPtr := flag.String("net", "tree", "The default algorithm to build the network connections")
	journalPtr := flag.String("journal", "", "Rerun a Journal of messages")
	followerPtr := flag.Bool("follower", false, "If true, force node to be a follower.  Only used when replaying a journal.")
	stylePtr := flag.String("style", "sim", "sim, tcp, ether - chooses the node/network style.")
	dbPtr := flag.String("db", "", "Override the Database in the Config file and use this Database implementation")
	folderPtr := flag.String("folder", "m2", "Directory in .factom to store nodes. (eg: multiple nodes on one filesystem support)")
	servePtr := flag.String("serve", "", "Port to start a TCP server on.")
	connectPtr := flag.String("connect", "", "Address to connect into over TCP (eg: another factomd node)")
	portPtr := flag.Int("port", 8088, "Address to serve WSAPI on")

	flag.Parse()

	listenTo := *listenToPtr
	cnt := *cntPtr
	net := *netPtr
	journal := *journalPtr
	follower := *followerPtr
	db := *dbPtr
	style := *stylePtr
	folder := *folderPtr
	serve := *servePtr
	connect := *connectPtr
	port := *portPtr

	os.Stderr.WriteString(fmt.Sprintf("node     %d\n", listenTo))
	os.Stderr.WriteString(fmt.Sprintf("count    %d\n", cnt))
	os.Stderr.WriteString(fmt.Sprintf("net      \"%s\"\n", net))
	os.Stderr.WriteString(fmt.Sprintf("journal  \"%s\"\n", journal))
	os.Stderr.WriteString(fmt.Sprintf("follower \"%v\"\n", follower))
	os.Stderr.WriteString(fmt.Sprintf("folder   \"%s\"\n", folder))
	os.Stderr.WriteString(fmt.Sprintf("serve    \"%s\"\n", serve))
	os.Stderr.WriteString(fmt.Sprintf("connect  \"%s\"\n", connect))
	os.Stderr.WriteString(fmt.Sprintf("port     \"%s\"\n", port))

	switch style {
	case "sim":
		nodeStyle = cSimStyle
		os.Stderr.WriteString(fmt.Sprintf("style \"Sim Style\"\n"))

	case "tcp":
		nodeStyle = cTCPStyle
		os.Stderr.WriteString(fmt.Sprintf("style \"TCP Style\"\n"))
	case "ether":
		nodeStyle = cEthereumStyle
		os.Stderr.WriteString(fmt.Sprintf("style \"Ether Style\"\n"))
	default:
		nodeStyle = cSimStyle
		os.Stderr.WriteString(fmt.Sprintf("style \"Sim Style\"\n"))
	}

	if journal != "" {
		cnt = 1
	}

	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Net Sim Start!!!!!")
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
		fmt.Print("Waiting...\r\n")
		time.Sleep(3 * time.Second)
		os.Exit(0)
	})

	FactomConfigFilename := util.GetConfigFilename("m2")

	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))

	s.LoadConfig(FactomConfigFilename, folder)
	if journal != "" {
		s.DBType = "Map"
	}

	if follower {
		s.NodeMode = "FULL"
		s.IdentityChainID = primitives.Sha([]byte(time.Now().String()))
	} else {
		s.NodeMode = "SERVER"
	}

	if len(db) > 0 {
		s.DBType = db
	}
	s.SetOut(false)
	s.PortNumber = port
	s.Init()

	mLog.init(cnt)

	//************************************************
	// Actually setup the Network
	//************************************************

	// Make cnt Factom nodes
	for i := 0; i < cnt; i++ {
		makeServer(s) // We clone s to make all of our servers
	}

	switch net {
	case "long":
		fmt.Println("Using long Network")
		for i := 1; i < cnt; i++ {
			AddPeer(nodeStyle, fnodes, i-1, i)
		}
	case "loops":
		fmt.Println("Using loops Network")
		for i := 1; i < cnt; i++ {
			AddPeer(nodeStyle, fnodes, i-1, i)
		}
		for i := 0; (i+17)*2 < cnt; i += 17 {
			AddPeer(nodeStyle, fnodes, i%cnt, (i+5)%cnt)
		}
		for i := 0; (i+13)*2 < cnt; i += 13 {
			AddPeer(nodeStyle, fnodes, i%cnt, (i+7)%cnt)
		}
	case "tree":
		index := 0
		row := 1
	treeloop:
		for i := 0; true; i++ {
			for j := 0; j <= i; j++ {
				AddPeer(nodeStyle, fnodes, index, row)
				AddPeer(nodeStyle, fnodes, index, row+1)
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
			AddPeer(nodeStyle, fnodes, index, index+circleSize-1)
			for i := index; i < index+circleSize-1; i++ {
				AddPeer(nodeStyle, fnodes, i, i+1)
			}
			index += circleSize

			AddPeer(nodeStyle, fnodes, index, index-circleSize/3)
			AddPeer(nodeStyle, fnodes, index+2, index-circleSize-circleSize*2/3-1)
			AddPeer(nodeStyle, fnodes, index+3, index-(2*circleSize)-circleSize*2/3)
			AddPeer(nodeStyle, fnodes, index+5, index-(3*circleSize)-circleSize*2/3+1)

			if index >= len(fnodes) {
				break
			}
		}
	default:
		fmt.Println("Didn't understand network type. Known types: mesh, long, circles, tree, loops.  Using a Long Network")
		for i := 1; i < cnt; i++ {
			AddPeer(nodeStyle, fnodes, i-1, i)
		}

	}
	if journal != "" {
		go LoadJournal(s, journal)
		startServers(false)
	} else {
		startServers(true)
	}

	// Right before we hand off to sim control, lets set up our network connections to other nodes
	// for the test point-to-point network.
	if 0 != len(serve) { // Start serving on the given port.
		port, _ := strconv.Atoi(serve)
		RemoteServeOnPort(fnodes, port)
	}

	if 0 != len(connect) { // Connect to the remote server.
		success := false
		for attempts := 0; attempts < 5 && !success; attempts++ {
			err := RemoteConnect(fnodes, connect)
			if nil == err {
				success = true
			} else {
				fmt.Println("Failed to connect, sleeping for 10 seconds and trying again.")
				time.Sleep(10 * time.Second)
			}
		}
		fmt.Println("Unable to connect to remote peer after 5 attempts!")

	}
	go wsapi.Start(fnodes[0].State)

	SimControl(listenTo)

}

// AddPeer adds a peer of the indicated type. There's probably a better
// way to do  this using a closure or maybe a superclass function (but go isn't
// "OO" so this isn't obvious to me.  This hack works for now.)
func AddPeer(nodeStyle int, fnodes []*FactomNode, i1 int, i2 int) {
	switch nodeStyle {
	case cSimStyle:
		AddSimPeer(fnodes, i1, i2)
	case cTCPStyle:
		AddNetPeer(fnodes, i1, i2)
	case cEthereumStyle:
		AddNetPeer(fnodes, i1, i2)
	default:
		AddSimPeer(fnodes, i1, i2)
	}
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
