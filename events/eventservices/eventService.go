package eventservices

import "github.com/FactomProject/factomd/events"

type EventService interface {
	Send(event events.EventInput) error
}

type EventServiceControl interface {
	GetBroadcastContent() BroadcastContent
	Shutdown()
	IsSendStateChangeEvents() bool
}
