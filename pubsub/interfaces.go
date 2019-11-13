package pubsub

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

// IPublisher is a routine that handles all publishes for a given publisher.
type IPublisher interface {
	IReadOnlyPublisher
	Start()

	// Publish will publish the event to all subscribers
	Write(o interface{})

	// Close should be called when all publishing events are done.
	// All subscribers can expect nothing new to ever be published.
	Close() error

	Subscribe(subscriber IPubSubscriber) bool
	Unsubscribe(subscriber IPubSubscriber) bool

	// Allow setting of the logger
	SetLogger(log interfaces.Log)

	// Informational Methods
	// only called by the registry
	setPath(path string)
}

type IReadOnlyPublisher interface {
	// NumberOfSubscribers is a method mainly used for debugging to
	// keep track of the size of a publisher.
	NumberOfSubscribers() int

	Path() string
	Logger() interfaces.Log
}

// TODO: Should we have some Quality of Service common params?
//		Like: Best Effort, buffering (might not want a buffer),
//		Data ownership (allow/disallow modification?)

type IPubSubscriber interface {
	// Publisher allows some access to the publisher for logging/debugging
	Publisher() IReadOnlyPublisher

	// setUnsubscribe is only called by a publisher
	setUnsubscribe(unsub func())
	// setPublisher is only called by a publisher
	setPublisher(pub IReadOnlyPublisher)

	// write is only called by a publisher
	write(o interface{})

	// Done is a function that can be called by the publisher to tell
	// the subscriber the publisher is done executing, and will be closed.
	// This means no new data will ever be received
	done()
}

type ISubscriber interface {
	// Should we have some common functions for subscribers?
}

type ISubscriberWrapper interface {
	// Base returns the underlying subscriber
	Base() IPubSubscriber
	Wrap(subscriber IPubSubscriber) IPubSubscriber
}

type IPublisherWrapper interface {
	IPublisher

	// Base returns the underlying publisher
	Base() IPublisher
	Wrap(subscriber IPublisher) IPublisherWrapper
	Publish(path string) IPublisherWrapper
}
