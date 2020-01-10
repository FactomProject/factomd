package controlpanel

import (
	"fmt"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/alexandrevicenzi/go-sse"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	host       = ""
	port       = 3001
	tlsEnabled = false
	keyFile    = ""
	certFile   = ""
)

func ControlPanel() {

	go func() {
		router := mux.NewRouter()

		webHandler := NewWebHandler()
		webHandler.RegisterRoutes(router)

		server := sse.NewServer(nil)

		eventHandler := &eventHandler{server: server}
		defer eventHandler.Shutdown()

		eventHandler.RegisterRoutes(router)
		eventHandler.RegisterChannel("channel-1", func() *sse.Message { return sse.SimpleMessage(time.Now().String()) }, 3*time.Second)

		// this causes panic if the path not exists
		channel := pubsub.SubFactory.Channel(5)
		subscription := channel.Subscribe("test")
		go func() {
			for data := range subscription.Channel() {
				message := sse.SimpleMessage(fmt.Sprintf("%v", data))
				server.SendMessage(URL_PREFIX+"channel-2", message)
			}
		}()

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

/*
 * stateful event handler
 */
type event struct {
	i int
}

func (event *event) message() func() *sse.Message {
	return func() *sse.Message {
		event.i++
		// return sse.SimpleMessage(strconv.Itoa(event.i))
		return sse.NewMessage(strconv.Itoa(event.i), strconv.Itoa(event.i), "")
	}
}
