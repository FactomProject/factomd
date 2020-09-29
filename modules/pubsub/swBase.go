package pubsub

type SubWrapBase struct {
	base IPubSubscriber
}

func (b *SubWrapBase) SetBase(sub IPubSubscriber) {
	if w, ok := sub.(ISubscriberWrapper); ok {
		b.base = w.Base()
	} else {
		b.base = sub
	}
}

func (b *SubWrapBase) Base() IPubSubscriber {
	return b.base
}
