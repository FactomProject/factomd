package bmv

import (
	"context"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/pubsub"
)

type BasicMessageValidator struct {
	// msgs is where all the incoming messages com from.
	msgs  *pubsub.SubChannel
	times *pubsub.SubChannel

	pub pubsub.IPublisher

	replay *MsgReplay
}

func NewBasicMessageValidator() *BasicMessageValidator {
	b := new(BasicMessageValidator)
	b.msgs = pubsub.SubFactory.Channel(100)  //.Subscribe("path?")
	b.times = pubsub.SubFactory.Channel(100) //.Subscribe("path?")

	b.pub = pubsub.PubFactory.Threaded(100).Publish("/bmv", pubsub.PubMultiWrap())

	b.replay = NewMsgReplay(6)
	return b
}

func (b *BasicMessageValidator) Subscribe() {
	// TODO: Find actual paths
	b.msgs = b.msgs.Subscribe("/msgs")
	b.times = b.times.Subscribe("/blocktime")
}

func (b *BasicMessageValidator) Run(ctx context.Context) {
	go b.pub.Start()
	for {
		select {
		case <-ctx.Done():
			return
		case blockTime := <-b.times.Updates:
			b.replay.Recenter(blockTime.(time.Time))
		case data := <-b.msgs.Updates:
			msg, ok := data.(interfaces.IMsg)
			if !ok {
				continue
			}

			if b.replay.UpdateReplay(msg) < 0 {
				continue // Already seen
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
