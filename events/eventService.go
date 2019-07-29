package events

type EventService interface {
	Send(event EventInput) error
}

type EventServiceControl interface {
	HasQueuedMessages() bool
	WaitForQueuedMessages()
}
