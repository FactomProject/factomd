package events

import (
	"github.com/FactomProject/factomd/events/contentfiltermode"
)

type EventService interface {
	Send(event EventInput) error
}

type EventServiceControl interface {
	GetContentFilterMode() contentfiltermode.ContentFilterMode
	Shutdown()
	IsResendRegistrationsOnStateChange() bool
}
