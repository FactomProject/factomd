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

	// The various publishers for various messages sorted by type
	groups []msgPub

	// The rest of the messages
	// pubList []pubsub.IPublisher
	// pubs    map[byte]pubsub.IPublisher
	rest pubsub.IPublisher

	replay *MsgReplay

	NodeName string
}

type msgPub struct {
	Name  string
	Types []byte
}

func NewBasicMessageValidator(nodeName string) *BasicMessageValidator {
	b := new(BasicMessageValidator)
	b.NodeName = nodeName
	b.msgs = pubsub.SubFactory.Channel(100)  //.Subscribe("path?")
	b.times = pubsub.SubFactory.Channel(100) //.Subscribe("path?")

	// b.groups = []msgPub{
	// 	// Each group is a publisher
	// 	{Name: "missing_messages", Types: []byte{constants.MISSING_MSG}},
	// 	{Name: "missing_dbstates", Types: []byte{constants.DBSTATE_MISSING_MSG}},
	// }
	//
	// // TODO: Remove this next line to keep the multiple publishers
	// b.groups = []msgPub{}

	// for _, g := range b.groups {
	// 	publisher := pubsub.PubFactory.Threaded(100).Publish(
	// 		pubsub.GetPath(b.NodeName, "bmv", g.Name), pubsub.PubMultiWrap())
	// 	for _, t := range g.Types {
	// 		b.pubs[t] = publisher
	// 	}
	// 	b.pubList = append(b.pubList, publisher)
	// }

	b.rest = pubsub.PubFactory.Threaded(100).Publish(pubsub.GetPath(b.NodeName, "bmv", "rest"), pubsub.PubMultiWrap())

	b.replay = NewMsgReplay(6)
	return b
}

func (b *BasicMessageValidator) Subscribe() {
	// TODO: Find actual paths
	b.msgs = b.msgs.Subscribe(pubsub.GetPath(b.NodeName, "msgs"))
	b.times = b.times.Subscribe(pubsub.GetPath(b.NodeName, "blocktime"))
}

func (b *BasicMessageValidator) ClosePublishing() {
	// for _, pub := range b.pubList {
	// 	_ = pub.Close()
	// }
	_ = b.rest.Close()
}

func (b *BasicMessageValidator) Run(ctx context.Context) {
	go b.rest.Start()
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
	// if p, ok := b.pubs[msg.Type()]; ok {
	// 	p.Write(msg)
	// 	return
	// }
	// Write to all pubs we are managing
	b.rest.Write(msg)
}
