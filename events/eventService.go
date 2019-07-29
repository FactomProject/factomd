package events

type EventService interface {
	Send(event EventInput) error
	HasQueuedMessages() bool
	WaitForQueuedMessages()
}
