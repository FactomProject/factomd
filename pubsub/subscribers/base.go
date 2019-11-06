package subscribers

// Base implements some primitive subscriber behavior that is likely to
// be used by all subscribers.
type Base struct {
	unsub func()
}

func (b *Base) SetUnsubscribe(unsub func()) {
	b.unsub = unsub
}

// Unsubscribe will stop subscribing to the publisher
func (b *Base) Unsubscribe() {
	b.unsub()
}

// Done is a function that can be called by the publisher to tell
// the subscriber the publisher is done executing, and will be closed.
func (b *Base) Done() {
	// Noop by default
}
