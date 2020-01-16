package controlpanel

import (
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/alexandrevicenzi/go-sse"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

var (
	host       = ""
	port       = 3001
	tlsEnabled = false
	keyFile    = ""
	certFile   = ""
)

type controlPanel struct {
	MsgInputSubscription       *pubsub.SubChannel
	MovedToHeightSubscription  *pubsub.SubChannel
	BalanceChangedSubscription *pubsub.SubChannel
	DBlockCreatedSubscription  *pubsub.SubChannel
	EomTickerSubscription      *pubsub.SubChannel
}

func New(factomNodeName string) {
	go func() {
		router := mux.NewRouter()

		webHandler := NewWebHandler()
		webHandler.RegisterRoutes(router)

		server := sse.NewServer(nil)

		eventHandler := &eventHandler{server: server}
		defer eventHandler.Shutdown()

		eventHandler.RegisterRoutes(router)
		eventHandler.RegisterChannel("channel-1", func() *sse.Message { return sse.SimpleMessage(time.Now().String()) }, 3*time.Second)

		controlPanel := controlPanel{
			MsgInputSubscription:       pubsub.SubFactory.Channel(100),
			MovedToHeightSubscription:  pubsub.SubFactory.Channel(100),
			BalanceChangedSubscription: pubsub.SubFactory.Channel(100),
			DBlockCreatedSubscription:  pubsub.SubFactory.Channel(100),
			EomTickerSubscription:      pubsub.SubFactory.Channel(100),
		}

		controlPanel.MovedToHeightSubscription.Subscribe(pubsub.GetPath(factomNodeName, event.Path.Seq))
		go controlPanel.pushEvents(server)

		address := fmt.Sprintf("%s:%d", host, port)
		webserver := &http.Server{Addr: address, Handler: router}

		if tlsEnabled {
			if err := webserver.ListenAndServeTLS(certFile, keyFile); err != http.ErrServerClosed {
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
				message := sse.SimpleMessage(fmt.Sprintf("%v", msg))
				server.SendMessage(URL_PREFIX+"general-events", message)
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
			if event, ok := v.(*event.Balance); ok {
				message := sse.SimpleMessage(fmt.Sprintf("%v", event))
				server.SendMessage(URL_PREFIX+"general-events", message)
			}
		case v := <-controlPanel.DBlockCreatedSubscription.Updates:
			if event, ok := v.(*event.Directory); ok {
				message := sse.SimpleMessage(fmt.Sprintf("%v", event))
				server.SendMessage(URL_PREFIX+"general-events", message)
			}
		case v := <-controlPanel.EomTickerSubscription.Updates:
			if event, ok := v.(*event.DBHT); ok {
				message := sse.SimpleMessage(fmt.Sprintf("%v", event))
				server.SendMessage(URL_PREFIX+"general-events", message)
			}
		}
	}
}
