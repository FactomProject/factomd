package controlpanel

import (
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/controlPanel/pages"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/alexandrevicenzi/go-sse"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type controlPanel struct {
	MsgInputSubscription          *pubsub.SubChannel
	MsgOutputSubscription         *pubsub.SubChannel
	MovedToHeightSubscription     *pubsub.SubChannel
	BalanceChangedSubscription    *pubsub.SubChannel
	DBlockCreatedSubscription     *pubsub.SubChannel
	EomTickerSubscription         *pubsub.SubChannel
	ConnectionAddedSubscription   *pubsub.SubChannel
	ConnectionRemovedSubscription *pubsub.SubChannel
}

// New Control Panel.
// takes a follower name to subscribe on events from the pub sub
func New(config *Config) {
	go func() {
		router := mux.NewRouter()

		indexPage := pages.Index{
			FactomNodeName: config.FactomNodeName,
			BuildNumber:    config.BuildNumer,
			Version:        config.Version,
		}

		webHandler := NewWebHandler(indexPage)
		webHandler.RegisterRoutes(router)

		server := sse.NewServer(nil)

		eventHandler := &eventHandler{server: server}
		defer eventHandler.Shutdown()

		eventHandler.RegisterRoutes(router)
		eventHandler.RegisterChannel("channel-1", func() *sse.Message { return sse.SimpleMessage(time.Now().String()) }, 3*time.Second)

		controlPanel := controlPanel{
			MsgInputSubscription:          pubsub.SubFactory.Channel(100),
			MsgOutputSubscription:         pubsub.SubFactory.Channel(100),
			MovedToHeightSubscription:     pubsub.SubFactory.Channel(100),
			BalanceChangedSubscription:    pubsub.SubFactory.Channel(100),
			DBlockCreatedSubscription:     pubsub.SubFactory.Channel(100),
			EomTickerSubscription:         pubsub.SubFactory.Channel(100),
			ConnectionAddedSubscription:   pubsub.SubFactory.Channel(100),
			ConnectionRemovedSubscription: pubsub.SubFactory.Channel(100),
		}

		// leader output
		// controlPanel.MsgInputSubscription.Subscribe(pubsub.GetPath("FNode0", event.Path.LeaderMsgOut))

		// network inputs
		controlPanel.MsgInputSubscription.Subscribe(pubsub.GetPath(config.FactomNodeName, "bmv", "rest"))

		// internal events
		controlPanel.MovedToHeightSubscription.Subscribe(pubsub.GetPath(config.FactomNodeName, event.Path.Seq))
		controlPanel.BalanceChangedSubscription.Subscribe(pubsub.GetPath(config.FactomNodeName, event.Path.Bank))
		controlPanel.DBlockCreatedSubscription.Subscribe(pubsub.GetPath(config.FactomNodeName, event.Path.Directory))
		//controlPanel.EomTickerSubscription.Subscribe(pubsub.GetPath(config.FactomNodeName, event.Path.EOM))
		controlPanel.ConnectionAddedSubscription.Subscribe(pubsub.GetPath(event.Path.ConnectionAdded))
		controlPanel.ConnectionRemovedSubscription.Subscribe(pubsub.GetPath(event.Path.ConnectionRemoved))

		go controlPanel.pushEvents(server)

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
	}()
}

func (controlPanel *controlPanel) pushEvents(server *sse.Server) {
	for {
		select {
		case v := <-controlPanel.MsgInputSubscription.Updates:
			if msg, ok := v.(interfaces.IMsg); ok {
				data, err := json.Marshal(msg)
				if err != nil {
					message := sse.SimpleMessage(string(data))
					server.SendMessage(URL_PREFIX+"general-events", message)
				}
			}
		case v := <-controlPanel.MovedToHeightSubscription.Updates:
			if dbHeight, ok := v.(*event.DBHT); ok {
				data, err := json.Marshal(dbHeight)
				if err != nil {
					log.Println("failed to serialize push event: ", err)
				}
				message := sse.SimpleMessage(string(data))
				server.SendMessage(URL_PREFIX+"move-to-height", message)
			}
		case v := <-controlPanel.BalanceChangedSubscription.Updates:
			if balance, ok := v.(*event.Balance); ok {
				data, err := json.Marshal(balance)
				if err != nil {
					message := sse.SimpleMessage(string(data))
					server.SendMessage(URL_PREFIX+"general-events", message)
				}
			}
		case v := <-controlPanel.DBlockCreatedSubscription.Updates:
			if directory, ok := v.(*event.Directory); ok {
				data, err := json.Marshal(directory)
				if err != nil {
					message := sse.SimpleMessage(string(data))
					server.SendMessage(URL_PREFIX+"general-events", message)
				}
			}
		case v := <-controlPanel.EomTickerSubscription.Updates:
			if dbHeight, ok := v.(*event.DBHT); ok {
				data, err := json.Marshal(dbHeight)
				if err != nil {
					message := sse.SimpleMessage(string(data))
					server.SendMessage(URL_PREFIX+"general-events", message)
				}
			}
		case v := <-controlPanel.ConnectionAddedSubscription.Updates:
			if connectionInfo, ok := v.(*event.ConnectionAdded); ok {
				data, err := json.Marshal(connectionInfo)
				if err != nil {
					message := sse.SimpleMessage(string(data))
					server.SendMessage(URL_PREFIX+"general-events", message)
				}
			}
		case v := <-controlPanel.ConnectionRemovedSubscription.Updates:
			if connectionInfo, ok := v.(*event.ConnectionRemoved); ok {
				data, err := json.Marshal(connectionInfo)
				if err != nil {
					message := sse.SimpleMessage(string(data))
					server.SendMessage(URL_PREFIX+"general-events", message)
				}
			}
		}
	}
}
