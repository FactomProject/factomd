package pubsub

type PubWrapBase struct {
	base IPublisher
}

func (b *PubWrapBase) SetBase(sub IPublisher) {
	if w, ok := sub.(IPublisherWrapper); ok {
		b.base = w.Base()
	} else {
		b.base = sub
	}
}

func (b *PubWrapBase) Base() IPublisher {
	return b.base
}
