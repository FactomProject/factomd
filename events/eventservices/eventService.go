package eventservices

import (
	"github.com/FactomProject/factomd/events/eventinput"
)

type EventService interface {
	Send(event eventinput.EventInput) error
}

type EventServiceControl interface {
	GetBroadcastContent() BroadcastContent
	Shutdown()
	IsSendStateChangeEvents() bool
}
