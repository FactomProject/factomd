package controlpanel

import (
	"github.com/alexandrevicenzi/go-sse"
	"github.com/gorilla/mux"
	"log"
	"time"
)

const URL_PREFIX = "/events/"

// EventHandler handles the subscriptions for the UI.
type EventHandler interface {
	Shutdown()
	RegisterRoutes(router *mux.Router)
	RegisterChannel(channel string, message func() *sse.Message, interval time.Duration)
	Server() *sse.Server
}

type eventHandler struct {
	server *sse.Server
}

// NewEventHandler create new event handler to allow and handle subscriptions from the UI
func NewEventHandler() EventHandler {
	return &eventHandler{
		server: sse.NewServer(&sse.Options{Logger: nil}), // NOTE: this disables// very verbose sse messages
	}
}

func (handler *eventHandler) Server() *sse.Server {
	return handler.server
}

// RegisterRoutes registers the routes with where channels will be available
func (handler *eventHandler) RegisterRoutes(router *mux.Router) {
	router.PathPrefix(URL_PREFIX).Handler(handler.server)
}

// RegisterChannel setup dispatcher for messages to channel registered
func (handler *eventHandler) RegisterChannel(channel string, message func() *sse.Message, interval time.Duration) {
	go func() {
		for {
			message := message()
			handler.server.SendMessage(URL_PREFIX+channel, message)
			time.Sleep(interval)
		}
	}()
}

func (handler *eventHandler) Shutdown() {
	log.Println("event service shutdown")
	handler.server.Shutdown()
}
