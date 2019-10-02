package events

import "github.com/FactomProject/factomd/events/allowcontent"

type EventService interface {
	Send(event EventInput) error
}

type EventServiceControl interface {
	GetAllowContent() allowcontent.AllowContent
	Shutdown()
	IsResendRegistrationsOnStateChange() bool
}
