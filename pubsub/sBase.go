package pubsub

// SubBase implements some primitive subscriber behavior that is likely to
// be used by all subscribers.
type SubBase struct {
	unsub func()
}

func (b *SubBase) setUnsubscribe(unsub func()) {
	b.unsub = unsub
}

// Unsubscribe will stop subscribing to the publisher
func (b *SubBase) Unsubscribe() {
	b.unsub()
}

// Done is a function that can be called by the publisher to tell
// the subscriber the publisher is done executing, and will be closed.
func (b *SubBase) done() {
	// Noop by default
}

func (b *SubBase) write(o interface{}) {
	// Noop by default
}

func (b *SubBase) Subscribe(path string, wrappers ...ISubscriberWrapper) *SubBase {
	globalSubscribeWith(path, b, wrappers...)
	return b
}
