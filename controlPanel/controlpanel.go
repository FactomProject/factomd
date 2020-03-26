package controlpanel

import (
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/controlPanel/pages"
	"github.com/FactomProject/factomd/modules/events"
	"github.com/FactomProject/factomd/modules/pubsub"
	"github.com/alexandrevicenzi/go-sse"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
)

type controlPanel struct {
	subscriptions
	DisplayState displayState
	DisplayDump  displayDump
}

type subscriptions struct {
	MsgInputSubscription          *pubsub.SubChannel
	MsgOutputSubscription         *pubsub.SubChannel
	MovedToHeightSubscription     *pubsub.SubChannel
	BalanceChangedSubscription    *pubsub.SubChannel
	DBlockCreatedSubscription     *pubsub.SubChannel
	EomTickerSubscription         *pubsub.SubChannel
	ConnectionMetricsSubscription *pubsub.SubChannel
	ProcessListInfo               *pubsub.SubChannel
	StateUpdate                   *pubsub.SubChannel
}

// displayState is the state which contain the information that changes in the UI.
type displayState struct {
	lock             sync.RWMutex
	NodeTime         string `json:"NodeTime"`
	CurrentHeight    uint32 `json:"CurrentHeight"`
	CurrentMinute    int    `json:"CurrentMinute"`
	LeaderHeight     uint32 `json:"LeaderHeight"`
	CompleteHeight   uint32 `json:"CompleteHeight"`
	FederatedServers string `json:"FederatedServers"`
	AuditServers     string `json:"AuditServers"`
	Connections      map[string]string
}

type displayDump struct {
	lock        sync.RWMutex
	Summary     string
	ProcessList string
	PrintMap    string
	Servers     serverDump
}

type serverDump struct {
	Authorities string `json:"Authorities"`
	Identities  string `json:"Identities"`
	Node        string `json:"Node"`
}

// New Control Panel.
// takes a follower name to subscribe on events from the pub sub
func New(config *Config) {
	router := mux.NewRouter()

	indexPage := pages.IndexContent{
		NodeName:    config.NodeName,
		BuildNumber: config.BuildNumer,
		Version:     config.Version,
	}

	webHandler := NewWebHandler(indexPage)
	webHandler.RegisterRoutes(router)

	eventHandler := NewEventHandler()
	defer eventHandler.Shutdown()

	eventHandler.RegisterRoutes(router)

	controlPanel := controlPanel{
		subscriptions: subscriptions{
			MsgInputSubscription:          pubsub.SubFactory.Channel(100),
			MsgOutputSubscription:         pubsub.SubFactory.Channel(100),
			MovedToHeightSubscription:     pubsub.SubFactory.Channel(100),
			BalanceChangedSubscription:    pubsub.SubFactory.Channel(100),
			DBlockCreatedSubscription:     pubsub.SubFactory.Channel(100),
			EomTickerSubscription:         pubsub.SubFactory.Channel(100),
			ConnectionMetricsSubscription: pubsub.SubFactory.Channel(100),
			ProcessListInfo:               pubsub.SubFactory.Channel(100),
			StateUpdate:                   pubsub.SubFactory.Channel(100),
		},
		DisplayState: displayState{
			CurrentHeight:  0,
			CurrentMinute:  0,
			LeaderHeight:   config.LeaderHeight,
			CompleteHeight: config.CompleteHeight,
		},
		DisplayDump: displayDump{
			Servers: serverDump{
				Node: nodeInfo(config.NodeName, config.IdentityChainID, config.PublicKey),
			},
		},
	}

	// leader output
	// controlPanel.MsgInputSubscription.Subscribe(pubsub.GetPath("FNode0", event.Path.LeaderMsgOut))

	// network inputs
	controlPanel.MsgInputSubscription.Subscribe(pubsub.GetPath(config.NodeName, "bmv", "rest"))

	// internal events
	controlPanel.MovedToHeightSubscription.Subscribe(pubsub.GetPath(config.NodeName, events.Path.Seq))
	controlPanel.BalanceChangedSubscription.Subscribe(pubsub.GetPath(config.NodeName, events.Path.Bank))
	controlPanel.DBlockCreatedSubscription.Subscribe(pubsub.GetPath(config.NodeName, events.Path.Directory))
	//controlPanel.EomTickerSubscription.Subscribe(pubsub.GetPath(config.NodeName, event.Path.EOM))
	controlPanel.ConnectionMetricsSubscription.Subscribe(pubsub.GetPath(config.NodeName, events.Path.ConnectionMetrics))

	// control panel details
	controlPanel.ProcessListInfo.Subscribe(pubsub.GetPath(config.NodeName, events.Path.ProcessListInfo))
	controlPanel.StateUpdate.Subscribe(pubsub.GetPath(config.NodeName, events.Path.StateUpdate))

	go controlPanel.handleEvents(eventHandler.Server())

	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	webserver := &http.Server{Addr: address, Handler: router}

	log.Printf("start control panel at: %s", address)
	if config.TLSEnabled {
		if err := webserver.ListenAndServeTLS(config.CertFile, config.KeyFile); err != http.ErrServerClosed {
			log.Fatalf("control panel failed: %v", err)
		}
	} else {
		if err := webserver.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("control panel failed: %v", err)
		}
	}
}

// handleEvents receives events from to which the control panel is subscribed on. It updates its
// state and pushes the state to the UI.
// TODO rewrite the function such that it handles all subscription events correctly.
func (controlPanel *controlPanel) handleEvents(server *sse.Server) {
	for {
		select {
		case v := <-controlPanel.MsgInputSubscription.Updates:
			if msg, ok := v.(interfaces.IMsg); ok {
				data, err := json.Marshal(msg)
				if err != nil {
					log.Printf("failed to serialize push event: %v", err)
					break
				}

				//log.Printf("msg input: %s", data)
				message := sse.SimpleMessage(string(data))
				server.SendMessage(URL_PREFIX+"general-events", message)
			}
		case v := <-controlPanel.MovedToHeightSubscription.Updates:
			if dbHeight, ok := v.(*events.DBHT); ok {
				controlPanel.updateHeight(dbHeight.DBHeight, dbHeight.Minute)
				controlPanel.pushUpdate(server)

				data, err := json.Marshal(dbHeight)
				if err != nil {
					log.Printf("failed to serialize push event: %v", err)
					break
				}

				message := sse.SimpleMessage(string(data))
				server.SendMessage(URL_PREFIX+"move-to-height", message)
			}
		case v := <-controlPanel.BalanceChangedSubscription.Updates:
			if balance, ok := v.(*events.Balance); ok {
				data, err := json.Marshal(balance)
				if err != nil {
					log.Printf("failed to serialize push event: %v", err)
					break
				}

				log.Printf("balance changed: %s", data)
				message := sse.SimpleMessage(string(data))
				server.SendMessage(URL_PREFIX+"general-events", message)
			}
		case v := <-controlPanel.DBlockCreatedSubscription.Updates:
			if directory, ok := v.(*events.Directory); ok {
				data, err := json.Marshal(directory)
				if err != nil {
					log.Printf("failed to serialize push event: %v", err)
					break
				}

				log.Printf("directory block created: %s", data)
				message := sse.SimpleMessage(string(data))
				server.SendMessage(URL_PREFIX+"general-events", message)
			}
		case v := <-controlPanel.EomTickerSubscription.Updates:
			if oem, ok := v.(*events.EOM); ok {
				data, err := json.Marshal(oem)
				if err != nil {
					log.Printf("failed to serialize push event: %v", err)
					break
				}

				log.Printf("end of minute ticker: %s", data)
				message := sse.SimpleMessage(string(data))
				server.SendMessage(URL_PREFIX+"general-events", message)
			}
		case v := <-controlPanel.ConnectionMetricsSubscription.Updates:
			data, err := json.Marshal(v)
			if err != nil {
				log.Printf("failed to serialize push event: %v", err)
				break
			}

			log.Printf("connection metric: %s", data)
			message := sse.SimpleMessage(string(data))
			server.SendMessage(URL_PREFIX+"general-events", message)

		case v := <-controlPanel.StateUpdate.Updates:
			if stateUpdate, ok := v.(*events.StateUpdate); ok {
				controlPanel.updateState(stateUpdate.Summary, stateUpdate.AuthoritiesDetails, stateUpdate.IdentitiesDetails)
				controlPanel.updateLeaderHeight(stateUpdate.LeaderHeight)
				controlPanel.pushStateUpdate(server)
			}
		case v := <-controlPanel.ProcessListInfo.Updates:
			if processList, ok := v.(*events.ProcessListInfo); ok {
				controlPanel.updateNodeTime(processList.ProcessTime)
				controlPanel.updateProcessList(processList.Dump, processList.PrintMap)
				controlPanel.pushProcessList(server)
			}
		}
	}
}

// updateHeight of the display state
func (controlPanel *controlPanel) updateHeight(currentHeight uint32, currentMinute int) {
	controlPanel.DisplayState.lock.Lock()
	defer controlPanel.DisplayState.lock.Unlock()

	controlPanel.DisplayState.CurrentHeight = currentHeight
	controlPanel.DisplayState.CurrentMinute = currentMinute
}

func (controlPanel *controlPanel) updateNodeTime(timestamp interfaces.Timestamp) {
	controlPanel.DisplayState.lock.Lock()
	defer controlPanel.DisplayState.lock.Unlock()
	controlPanel.DisplayState.NodeTime = timestamp.String()
}
func (controlPanel *controlPanel) updateLeaderHeight(height uint32) {
	controlPanel.DisplayState.lock.Lock()
	defer controlPanel.DisplayState.lock.Unlock()
	controlPanel.DisplayState.LeaderHeight = height
}

func (controlPanel *controlPanel) updateProcessList(dump string, printMap string) {
	controlPanel.DisplayDump.lock.Lock()
	defer controlPanel.DisplayDump.lock.Unlock()
	controlPanel.DisplayDump.ProcessList = dump
	controlPanel.DisplayDump.PrintMap = printMap
}

func (controlPanel *controlPanel) updateState(summary string, authorities string, identities string) {
	controlPanel.DisplayDump.lock.Lock()
	defer controlPanel.DisplayDump.lock.Unlock()
	controlPanel.DisplayDump.Summary = summary
	controlPanel.DisplayDump.Servers.Authorities = authorities
	controlPanel.DisplayDump.Servers.Identities = identities
}

// pushUpdate push an update of the state to all subscribed UI's
func (controlPanel *controlPanel) pushUpdate(server *sse.Server) {
	controlPanel.DisplayState.lock.RLock()
	defer controlPanel.DisplayState.lock.RUnlock()

	data, err := json.Marshal(controlPanel.DisplayState)
	if err != nil {
		log.Println("failed to serialize push event: ", err)
		return
	}
	message := sse.SimpleMessage(string(data))
	server.SendMessage(URL_PREFIX+"update", message)
}

func (controlPanel *controlPanel) pushStateUpdate(server *sse.Server) {
	controlPanel.DisplayDump.lock.RLock()
	defer controlPanel.DisplayDump.lock.RUnlock()

	summaryMessage := sse.SimpleMessage(controlPanel.DisplayDump.Summary)
	server.SendMessage(URL_PREFIX+"summary", summaryMessage)

	data, err := json.Marshal(controlPanel.DisplayDump.Servers)
	if err != nil {
		log.Println("failed to serialize push event: ", err)
		return
	}

	serversMessage := sse.SimpleMessage(string(data))
	server.SendMessage(URL_PREFIX+"servers", serversMessage)
}

func (controlPanel *controlPanel) pushProcessList(server *sse.Server) {
	controlPanel.DisplayDump.lock.RLock()
	defer controlPanel.DisplayDump.lock.RUnlock()

	// push process list dump
	processListMessage := sse.SimpleMessage(controlPanel.DisplayDump.ProcessList)
	server.SendMessage(URL_PREFIX+"processlist", processListMessage)

	// push print map
	printMapMessage := sse.SimpleMessage(controlPanel.DisplayDump.PrintMap)
	server.SendMessage(URL_PREFIX+"printmap", printMapMessage)
}
