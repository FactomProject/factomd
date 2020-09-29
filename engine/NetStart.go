// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/FactomProject/factomd/controlPanel"
	"github.com/FactomProject/factomd/modules/events/eventmessages/generated/eventmessages"
	"github.com/FactomProject/factomd/modules/pubsub"
	"github.com/FactomProject/factomd/p2p"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/fnode"
	llog "github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/modules/debugsettings"
	"github.com/FactomProject/factomd/modules/leader"
	"github.com/FactomProject/factomd/modules/registry"
	"github.com/FactomProject/factomd/modules/worker"
	"github.com/FactomProject/factomd/simulation"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
	"github.com/FactomProject/logrustash"
	log "github.com/sirupsen/logrus"
)

var connectionMetricsChannel = make(chan map[string]p2p.PeerMetrics, 5000)
var mLog = new(MsgLog)
var p2pProxy *P2PProxy
var network *p2p.Network
var logPort string

func init() {
	messages.General = new(msgsupport.GeneralFactory)
	primitives.General = messages.General

	// Globally Enable/Disable Modules
	state.EnableLeaderThread = false // use new threaded leader
}

func echo(s string, more ...interface{}) {
	_, _ = os.Stderr.WriteString(fmt.Sprintf(s, more...))
}

func echoConfig(s *state.State, p *globals.FactomParams) {

	fmt.Println()
	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Factom parameters")
	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Listening to Node", p.ListenTo)
	fmt.Println(">>>>>>>>>>>>>>>>")

	pnet := p.Net
	if len(p.Fnet) > 0 {
		pnet = p.Fnet
		p.Net = "file"
	}

	echo("%20s %s\n", "Build", Build)
	echo("%20s %s\n", "Node name", p.NodeName)
	echo("%20s %v\n", "balancehash", messages.AckBalanceHash)
	echo("%20s %s\n", fmt.Sprintf("%s Salt", s.GetFactomNodeName()), s.Salt.String()[:16])
	echo("%20s %v\n", "enablenet", p.EnableNet)
	//echo("%20s %v\n", "net incoming", p2p.MaxNumberIncomingConnections)
	//echo("%20s %v\n", "net outgoing", p2p.NumberPeersToConnect)
	echo("%20s %v\n", "waitentries", p.WaitEntries)
	echo("%20s %d\n", "node", p.ListenTo)
	echo("%20s %s\n", "prefix", p.Prefix)
	echo("%20s %d\n", "node count", p.Cnt)
	echo("%20s %d\n", "FastSaveRate", p.FastSaveRate)
	echo("%20s \"%s\"\n", "net spec", pnet)
	echo("%20s %d\n", "Msgs droped", p.DropRate)
	echo("%20s \"%s\"\n", "database", p.Db)
	echo("%20s \"%s\"\n", "database for clones", p.CloneDB)
	echo("%20s \"%s\"\n", "peers", p.Peers)
	echo("%20s \"%t\"\n", "exclusive", p.Exclusive)
	echo("%20s \"%t\"\n", "exclusive_in", p.ExclusiveIn)
	echo("%20s %d\n", "block time", p.BlkTime)
	echo("%20s %v\n", "runtimeLog", p.RuntimeLog)
	echo("%20s %v\n", "rotate", p.Rotate)
	echo("%20s %v\n", "timeOffset", p.TimeOffset)
	echo("%20s %v\n", "keepMismatch", p.KeepMismatch)
	echo("%20s %v\n", "startDelay", p.StartDelay)
	echo("%20s %v\n", "Network", s.Network)
	echo("%20s %x (%s)\n", "customnet", p.CustomNet, p.CustomNetName)
	echo("%20s %v\n", "deadline (ms)", p.Deadline)
	echo("%20s %v\n", "tls", s.FactomdTLSEnable)
	echo("%20s %v\n", "selfaddr", s.FactomdLocations)
	echo("%20s \"%s\"\n", "rpcuser", s.RpcUser)
	echo("%20s \"%s\"\n", "corsdomains", s.CorsDomains)
	echo("%20s %d\n", "Start 2nd Sync at height", s.EntryBlockDBHeightComplete)

	echo(fmt.Sprintf("%20s %d\n", "faultTimeout", elections.FaultTimeout))

	if "" == s.RpcPass {
		echo(fmt.Sprintf("%20s %s\n", "rpcpass", "is blank"))
	} else {
		echo(fmt.Sprintf("%20s %s\n", "rpcpass", "is set"))
	}
	echo("%20s \"%d\"\n", "TCP port", s.PortNumber)
	echo("%20s \"%s\"\n", "pprof port", logPort)
	echo("%20s \"%d\"\n", "Control Panel port", s.ControlPanelPort)
	echo("\n\n")
}

// shutdown factomd
func interruptHandler() {
	fmt.Print("<Break>\n")
	fmt.Print("Gracefully shutting down the server...\n")
	for _, node := range fnode.GetFnodes() {
		node.State.ShutdownNode(0)
	}
	network.Stop()
	fmt.Print("Waiting...\r\n")
	time.Sleep(3 * time.Second)
	os.Exit(0)
}

func initEntryHeight(s *state.State, target int) {
	if target >= 0 {
		s.EntryBlockDBHeightComplete = uint32(target)
		s.LogPrintf("EntrySync", "Force with Sync2 NetStart EntryBlockDBHeightComplete = %d", s.EntryBlockDBHeightComplete)
	} else {
		height, err := s.DB.FetchDatabaseEntryHeight()
		if err != nil {
			s.LogPrintf("EntrySync", "Error reading EntryBlockDBHeightComplete NetStart EntryBlockDBHeightComplete = %d", s.EntryBlockDBHeightComplete)
			_, _ = os.Stderr.WriteString(fmt.Sprintf("ERROR reading Entry DBHeight Complete: %v\n", err))
		} else {
			s.EntryBlockDBHeightComplete = height
			s.LogPrintf("EntrySync", "NetStart EntryBlockDBHeightComplete = %d", s.EntryBlockDBHeightComplete)
		}
	}
}

func NetStart(w *worker.Thread, p *globals.FactomParams, listenToStdin bool) {
	initEngine(w, p)
	for i := 0; i < p.Cnt; i++ {
		fnode.Factory(w)
	}
	startNetwork(w, p)
	startFnodes(w)
	startLiveFeed(w, p)
	startWebserver(w)
	startControlPanel(w)
	simulation.StartSimControl(w, p.ListenTo, listenToStdin)
}

// initialize package-level vars
func initEngine(w *worker.Thread, p *globals.FactomParams) {
	messages.AckBalanceHash = p.AckbalanceHash
	w.RegisterInterruptHandler(interruptHandler)

	// add these to the name substitution table in logs so election dumps of the authority set look better
	globals.FnodeNames["Fed"] = "erated "
	globals.FnodeNames["Aud"] = "id     "
	// nodes can spawn with a different thread lifecycle
	fnode.Factory = func(w *worker.Thread) {
		makeServer(w, p)
	}
}

// Anchoring related configurations
func initAnchors(s *state.State, reparse bool) {
	config := s.Cfg.(*util.FactomdConfig)
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
	if reparse {
		fmt.Println("Reparsing anchor chains...")
		err := s.GetDB().(*databaseOverlay.Overlay).ReparseAnchorChains()
		if err != nil {
			panic("Encountered an error while trying to re-parse anchor chains: " + err.Error())
		}
	}
}

// start live feed service (for fnode 0 only)
func startLiveFeed(w *worker.Thread, p *globals.FactomParams) {
	state0 := fnode.Get(0).State
	if state0.EventService == nil {
		return
	}
	config := state0.Cfg.(*util.FactomdConfig)

	if config.LiveFeedAPI.EnableLiveFeedAPI || p.EnableLiveFeedAPI {
		state0.EventService.ConfigService(state0, config, p)
		if p.UseLogstash {
			eventForward(w)
		}
	}
}

func startWebserver(w *worker.Thread) {
	state0 := fnode.Get(0).State
	wsapi.Start(w, state0)
	if state0.DebugExec() && state0.CheckFileName("graphData.txt") {
		go printGraphData("graphData.txt", 30)
	}

	// Start prometheus on port
	launchPrometheus(9876)
}

func startControlPanel(w *worker.Thread) {
	state0 := fnode.Get(0).State
	w.Run("controlpanel", func() {
		controlPanel.ServeControlPanel(state0.ControlPanelChannel, state0, connectionMetricsChannel, network, Build, state0.FactomNodeName)

		/*		controlPanelConfig := &controlpanel.Config{
					Port:       state0.ControlPanelPort,
					TLSEnabled: state0.FactomdTLSEnable,
					CertFile:   state0.FactomdTLSCertFile,
					KeyFile:    state0.FactomdTLSKeyFile,

					NodeName:   state0.FactomNodeName,
					BuildNumer: Build,
					Version:    FactomdVersion,

					CompleteHeight: state0.EntryBlockDBHeightComplete,
					LeaderHeight:   state0.LLeaderHeight,

					IdentityChainID: state0.GetIdentityChainID().String(),
					PublicKey:       state0.GetServerPublicKeyString(),
				}
				controlpanel.New(controlPanelConfig)*/
	})
}

func startNetwork(w *worker.Thread, p *globals.FactomParams) {
	node := fnode.Get(0)
	s := node.State

	// Modify Identities of simulated nodes
	if fnode.Len() > 1 && len(s.Prefix) == 0 {
		simulation.ModifySimulatorIdentities() // set proper chain id & keys
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
		networkID = p2p.NetworkID(binary.BigEndian.Uint32(p.CustomNet))
		for _, node := range fnode.GetFnodes() {
			node.State.CustomNetworkID = p.CustomNet
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

	// FIXME: should this be relocated?
	p2pconf := p2p.DefaultP2PConfiguration()
	p2pconf.Network = networkID
	p2pconf.SeedURL = seedURL
	p2pconf.ListenPort = networkPort
	p2pconf.Special = configPeers
	p2pconf.ReadDeadline = time.Minute * 5
	p2pconf.WriteDeadline = time.Minute * 5

	simulation.BuildNetTopology(p)

	/*// start a worker that publishes the connection metrics
	connectionMetricsPublisher := p2p.NewMetricPublisher(s.FactomNodeName, connectionMetricsChannel)
	connectionMetricsPublisher.Start(w)*/

	if p.EnableNet {
		nodeName := s.FactomNodeName
		if 0 < p.NetworkPortOverride {
			networkPort = fmt.Sprintf("%d", p.NetworkPortOverride)
		}

		var err error
		if network, err = p2p.NewNetwork(p2pconf); err != nil {
			fmt.Println(err)
			panic("Unable to start p2p network")
		} else {
			network.SetMetricsHook(func(pm map[string]p2p.PeerMetrics) {
				select {
				case connectionMetricsChannel <- pm:
				default:
				}
			})
			network.Run()
			s.NetworkController = network

			p2pProxy = new(P2PProxy).Initialize(nodeName, "P2P Network").(*P2PProxy)
			p2pProxy.Network = network
			p2pProxy.NameInit(s, "p2pProxy", reflect.TypeOf(p2pProxy).String())
			p2pProxy.Network = network
			p2pProxy.StartProxy(w)
			node.Peers = append(node.Peers, p2pProxy)
		}
	}
}

func printGraphData(filename string, period int) {
	downscale := int64(1)
	llog.LogPrintf(filename, "%10s%10s%10s%10s%10s%10s", "Dbh-:-min", "Node", "ProcessCnt", "ListPCnt", "UpdateState", "SleepCnt")
	for {
		for _, f := range fnode.GetFnodes() {
			s := f.State
			llog.LogPrintf(filename, "%10s%10s%10d%10d%10d%10d", fmt.Sprintf("%d-:-%d", s.LLeaderHeight, s.CurrentMinute), s.FactomNodeName, s.StateProcessCnt/downscale, s.ProcessListProcessCnt/downscale, s.StateUpdateState/downscale, s.ValidatorLoopSleepCnt/downscale)
		}
		time.Sleep(time.Duration(period) * time.Second)
	} // for ever ...
}

var state0Init sync.Once // we do some extra init for the first state

//**********************************************************************
// Functions that access variables in this method to set up Factom Nodes
// and start the servers.
//**********************************************************************
func makeServer(w *worker.Thread, p *globals.FactomParams) (node *fnode.FactomNode) {
	i := fnode.Len()

	if i == 0 {
		node = fnode.New(state.NewState(p, FactomdVersion))
	} else {
		node = fnode.New(state.Clone(fnode.Get(0).State, i).(*state.State))
	}

	// Election factory was created and passed int to avoid import loop
	//node.State.BindPublishers() // REVIEW: has this been fully refactored?
	node.State.NameInit(node, node.State.GetFactomNodeName()+"STATE", reflect.TypeOf(node.State).String())
	node.State.BuildPubRegistry()
	node.State.Initialize(w, new(electionMsgs.ElectionsFactory))

	state0Init.Do(func() {
		logPort = p.LogPort
		setupFirstAuthority(node.State)
		initEntryHeight(node.State, p.Sync2)
		initAnchors(node.State, p.ReparseAnchorChains)
		echoConfig(node.State, p) // print the config only once
		// Init settings
	})

	if state.EnableLeaderThread {
		l := leader.New(node.State)
		l.Start(w)
	}

	// TODO: Init any settings from the config
	debugsettings.NewNode(node.State.GetFactomNodeName())

	node.State.Logger = log.WithFields(log.Fields{"node-name": node.State.GetFactomNodeName(), "identity": node.State.GetIdentityChainID().String()})
	if p.UseLogstash {
		hookLogstash(node.State, p.LogstashURL)
	}

	return node
}

func startFnodes(w *worker.Thread) {
	state.CheckGrants() // check the grants table hard coded into the build is well formed.
	for i := range fnode.GetFnodes() {
		node := fnode.Get(i)
		w.Spawn(node.GetName()+"Thread", func(w *worker.Thread) {
			startServer(w, node)
		})
	}
}

func startServer(w *worker.Thread, node *fnode.FactomNode) {
	NetworkProcessorNet(w, node)
	s := node.State
	w.Run("MsgSort", s.MsgSort)

	w.Run("MsgExecute", s.MsgExecute)

	elections.Run(w, s)
	s.StartMMR(w)

	w.Run("DBStateCatchup", s.DBStates.Catchup)
	w.Run("LoadDatabase", s.LoadDatabase)

	entrySync := state.NewEntrySync(s)
	s.EntrySync = entrySync
	w.Run("SyncEntries", s.EntrySync.DelayedStart)
	w.Run("WriteEntries", s.WriteEntries)

	if !state.EnableLeaderThread {
		w.Run("EOMTicker", func() { Timer(node.State) })
	}
	w.Run("MMResponseHandler", s.MissingMessageResponseHandler.Run)
}

func setupFirstAuthority(s *state.State) {
	if len(s.IdentityControl.Authorities) > 0 {
		//Don't initialize first authority if we are loading during fast boot
		//And there are already authorities present
		return
	}

	_ = s.IdentityControl.SetBootstrapIdentity(s.GetNetworkBootStrapIdentity(), s.GetNetworkBootStrapKey())
}

// create a new simulated fnode
func AddNode() {
	p := registry.New()
	p.Register(func(w *worker.Thread) {
		i := fnode.Len()
		fnode.Factory(w)
		simulation.ModifySimulatorIdentity(i)
		simulation.AddSimPeer(fnode.GetFnodes(), i, i-1) // KLUDGE peer w/ only last node
		n := fnode.Get(i)
		startServer(w, n)
	})
	go p.Run()
	p.WaitForRunning()
}

// enable output to logstash
func hookLogstash(s *state.State, logStashURL string) error {
	hook, err := logrustash.NewAsyncHook("tcp", logStashURL, "factomdLogs")
	if err != nil {
		fmt.Printf("Failed to connect to logstash %v", err)
		return err
	}

	hook.ReconnectBaseDelay = time.Second // Wait for one second before first reconnect.
	hook.ReconnectDelayMultiplier = 2
	hook.MaxReconnectRetries = 10

	s.Logger.Logger.Hooks.Add(hook)
	return nil
}

// forward live feed events to logstash
// NOTE set env var 'EVENTLOG' to enable
func eventForward(w *worker.Thread) {
	if _, enabled := os.LookupEnv("EVENTLOG"); !enabled {
		return
	}
	w.Spawn("LiveFeed Logs", func(w *worker.Thread) {
		threadLogger := log.WithFields(log.Fields{"thread": w.ID, "process": w.PID})
		var feed *pubsub.SubChannel
		w.OnReady(func() {
			feed = pubsub.SubFactory.Channel(5000).Subscribe("/live-feed")
		})
		w.OnRun(func() {
			for {
				select {
				case v, ok := <-feed.Updates:
					if !ok {
						return
					}
					evt := v.(*eventmessages.FactomEvent)
					data, err := json.Marshal(evt.Event)
					if err != nil {
						panic(err)
					}
					threadLogger.WithFields(
						log.Fields{"Event": string(data)},
					).Info(evt.EventSource)
				}
			}
		})
	})
}
