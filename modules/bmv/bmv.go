package bmv

import (
	"context"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/pubsub"
)

type BasicMessageValidator struct {
	// sub is where all the incoming messages com from.
	sub *pubsub.SubChannel
	pub *pubsub.PubSimpleMulti

	// TODO: All well formed should be written to a publisher
}

func NewBasicMessageValidator() *BasicMessageValidator {
	b := new(BasicMessageValidator)
	b.sub = pubsub.SubFactory.Channel(100) //.Subscribe("path?")
	b.pub = pubsub.PubFactory.Multi(100).Publish("bmv")

	return b
}

func (b *BasicMessageValidator) Run(ctx context.Context) {
	go b.pub.Run()
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-b.sub.Updates:
			msg, ok := data.(interfaces.IMsg)
			if !ok {
				continue
			}

			if msg.WellFormed() {
				b.Write(msg)
			}
		}
	}
}

func (b *BasicMessageValidator) Write(msg interfaces.IMsg) {
	// Write to all pubs we are managing
	b.pub.Write(msg)
}
