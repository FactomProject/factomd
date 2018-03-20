// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"time"

	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/elections"
	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

type FactomNode struct {
	Index int
	State *state.State
	Peers []interfaces.IPeer
	MLog  *MsgLog
}

var fnodes []*FactomNode

var networkpattern string
var mLog = new(MsgLog)
var p2pProxy *P2PProxy
var p2pNetwork *p2p.Controller
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
	s.AddPrefix(p.prefix)
	FactomConfigFilename := util.GetConfigFilename("m2")
	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))
	s.LoadConfig(FactomConfigFilename, p.NetworkName)
	s.OneLeader = p.rotate
	s.TimeOffset = primitives.NewTimestampFromMilliseconds(uint64(p.timeOffset))
	s.StartDelayLimit = p.StartDelay * 1000
	s.Journaling = p.Journaling
	s.FactomdVersion = FactomdVersion

	log.SetOutput(os.Stdout)
	switch p.loglvl {
	case "none":
		log.SetOutput(ioutil.Discard)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	}

	if p.logjson {
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

	if p.fast == false {
		s.StateSaverStruct.FastBoot = false
	}
	if p.fastLocation != "" {
		s.StateSaverStruct.FastBootLocation = p.fastLocation
	}

	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Net Sim Start!")
	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Listening to Node", p.ListenTo)
	fmt.Println(">>>>>>>>>>>>>>>>")

	AddInterruptHandler(func() {
		fmt.Print("<Break>\n")
		fmt.Print("Gracefully shutting down the server...\n")
		for _, fnode := range fnodes {
			fmt.Print("Shutting Down: ", fnode.State.FactomNodeName, "\r\n")
			fnode.State.ShutdownChan <- 0
		}
		if p.EnableNet {
			p2pNetwork.NetworkStop()
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

	s.KeepMismatch = p.keepMismatch

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

	s.UseLogstash = p.useLogstash
	s.LogstashURL = p.logstashURL

	go StartProfiler(p.memProfileRate, p.exposeProfiling)

	s.AddPrefix(p.prefix)
	s.SetOut(false)
	s.Init()
	s.SetDropRate(p.DropRate)

	if p.Sync2 >= 0 {
		s.EntryDBHeightComplete = uint32(p.Sync2)
	} else {
		height, err := s.DB.FetchDatabaseEntryHeight()
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("ERROR: %v", err))
		} else {
			s.EntryDBHeightComplete = height
		}
	}

	mLog.Init(p.RuntimeLog, p.Cnt)

	setupFirstAuthority(s)

	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "Build", Build))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "balancehash", messages.AckBalanceHash))
	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "FNode 0 Salt", s.Salt.String()[:16]))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "enablenet", p.EnableNet))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "waitentries", p.WaitEntries))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "node", p.ListenTo))
	os.Stderr.WriteString(fmt.Sprintf("%20s %s\n", "prefix", p.prefix))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "node count", p.Cnt))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "net spec", pnet))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "Msgs droped", p.DropRate))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "journal", p.Journal))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database", p.Db))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "database for clones", p.CloneDB))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "peers", p.Peers))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%t\"\n", "exclusive", p.Exclusive))
	os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "block time", p.BlkTime))
	//os.Stderr.WriteString(fmt.Sprintf("%20s %d\n", "faultTimeout", p.FaultTimeout)) // TODO old fault timeout mechanism to be removed
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "runtimeLog", p.RuntimeLog))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "rotate", p.rotate))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "timeOffset", p.timeOffset))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "keepMismatch", p.keepMismatch))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "startDelay", p.StartDelay))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "Network", s.Network))
	os.Stderr.WriteString(fmt.Sprintf("%20s %x\n", "customnet", p.customNet))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "deadline (ms)", p.deadline))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "tls", s.FactomdTLSEnable))
	os.Stderr.WriteString(fmt.Sprintf("%20s %v\n", "selfaddr", s.FactomdLocations))
	os.Stderr.WriteString(fmt.Sprintf("%20s \"%s\"\n", "rpcuser", s.RpcUser))
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
	// Modify Identities of new nodes
	if len(fnodes) > 1 && len(s.Prefix) == 0 {
		modifyLoadIdentities() // We clone s to make all of our servers
	}

	// Setup the Skeleton Identity
	for i := range fnodes {
		fnodes[i].State.IntiateNetworkSkeletonIdentity()
	}

	// Start the P2P network
	var networkID p2p.NetworkID
	var seedURL, networkPort, specialPeers string
	switch s.Network {
	case "MAIN", "main":
		networkID = p2p.MainNet
		seedURL = s.MainSeedURL
		networkPort = s.MainNetworkPort
		specialPeers = s.MainSpecialPeers
		s.DirectoryBlockInSeconds = 600
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
		seedURL = s.CustomSeedURL
		networkPort = s.CustomNetworkPort
		specialPeers = s.CustomSpecialPeers
	default:
		panic("Invalid Network choice in Config File or command line. Choose MAIN, TEST, LOCAL, or CUSTOM")
	}

	connectionMetricsChannel := make(chan interface{}, p2p.StandardChannelSize)
	p2p.NetworkDeadline = time.Duration(p.deadline) * time.Millisecond

	if p.EnableNet {
		nodeName := fnodes[0].State.FactomNodeName
		if 0 < p.NetworkPortOverride {
			networkPort = fmt.Sprintf("%d", p.NetworkPortOverride)
		}

		ci := p2p.ControllerInit{
			NodeName:                 nodeName,
			Port:                     networkPort,
			PeersFile:                s.PeersFile,
			Network:                  networkID,
			Exclusive:                p.Exclusive,
			SeedURL:                  seedURL,
			SpecialPeers:             specialPeers,
			ConnectionMetricsChannel: connectionMetricsChannel,
		}
		p2pNetwork = new(p2p.Controller).Init(ci)
		fnodes[0].State.NetworkControler = p2pNetwork
		p2pNetwork.StartNetwork()
		p2pProxy = new(P2PProxy).Init(nodeName, "P2P Network").(*P2PProxy)
		p2pProxy.FromNetwork = p2pNetwork.FromNetwork
		p2pProxy.ToNetwork = p2pNetwork.ToNetwork

		fnodes[0].Peers = append(fnodes[0].Peers, p2pProxy)
		p2pProxy.StartProxy()
		// Command line peers lets us manually set special peers
		p2pNetwork.DialSpecialPeersString(p.Peers)

		go networkHousekeeping() // This goroutine executes once a second to keep the proxy apprised of the network status.
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

	// Initate dbstate plugin if enabled. Only does so for first node,
	// any more nodes on sim control will use default method
	fnodes[0].State.SetTorrentUploader(p.torUpload)
	if p.torManage {
		fnodes[0].State.SetUseTorrent(true)
		manager, err := LaunchDBStateManagePlugin(p.pluginPath, fnodes[0].State.InMsgQueue(), fnodes[0].State, fnodes[0].State.GetServerPrivateKey(), p.memProfileRate)
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

	// Start the webserver
	wsapi.Start(fnodes[0].State)

	// Start prometheus on port
	launchPrometheus(9876)
	// Start Package's prometheus
	state.RegisterPrometheus()
	p2p.RegisterPrometheus()
	leveldb.RegisterPrometheus()
	RegisterPrometheus()

	go controlPanel.ServeControlPanel(fnodes[0].State.ControlPanelChannel, fnodes[0].State, connectionMetricsChannel, p2pNetwork, Build)

	SimControl(p.ListenTo, listenToStdin)

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
		fnode.State.EFactory = new(electionMsgs.ElectionsFactory)
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
		go elections.Run(fnode.State)
	}
}

func setupFirstAuthority(s *state.State) {
	var id identity.Identity
	if len(s.Authorities) > 0 {
		//Don't initialize first authority if we are loading during fast boot
		//And there are already authorities present
		return
	}

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

	var auth identity.Authority
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
