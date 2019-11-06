package pubregistry

// IPublisher is a routine that handles all publishes for a given publisher.
type IPublisher interface {
	// NumberOfSubscribers is a method mainly used for debugging to
	// keep track of the size of a publisher.
	NumberOfSubscribers() int

	// Publish will publish the event to all subscribers
	Write(o interface{})

	// Close should be called when all publishing events are done.
	// All subscribers can expect nothing new to ever be published.
	Close() error

	Subscribe(subscriber IPubSubscriber) bool
	Unsubscribe(subscriber IPubSubscriber) bool

	// Informational Methods
	SetPath(name string)
	Path() string
}

// TODO: Should we have some Quality of Service common params?
//		Like: Best Effort, buffering (might not want a buffer),
//		Data ownership (allow/disallow modification?)

type IPubSubscriber interface {
	SetUnsubscribe(unsub func())
	Write(o interface{})

	// Done is a function that can be called by the publisher to tell
	// the subscriber the publisher is done executing, and will be closed.
	// This means no new data will ever be received
	Done()
}

type ISubscriber interface {
	// Should we have some common functions for subscribers?
}
