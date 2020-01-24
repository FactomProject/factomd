package pubsub_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestSubUnSub(t *testing.T) {
	ResetGlobalRegistry()
	nodeName := "FNode0"

	type Module struct {
		Path   string
		MsgOut IPublisher
		MsgIn0 *SubChannel
		MsgIn1 *SubChannel
	}

	p := &Module{Path: GetPath(nodeName, "test")}
	p.MsgOut = PubFactory.Threaded(10).Publish(p.Path)
	pub := p.MsgOut.(*PubThreaded)
	go pub.Start()

	p.MsgIn0 = SubFactory.Channel(10)
	p.MsgIn0.Subscribe(GetPath(nodeName, "test"))

	p.MsgIn1 = SubFactory.Channel(10)
	p.MsgIn1.Subscribe(GetPath(nodeName, "test"))

	replyChannel := make(chan interface{}, 10)

	getCount := func() {
		pub.CountSubscriberSync(replyChannel)
		for {
			select {
			case v := <-replyChannel:
				t.Logf("got count %v", v)
				if v.(int) == 2 {
					return
				}
			}
		}
	}
	getCount()

	// Write to both
	p.MsgOut.Write(messages.Bounce{Name: "foo"})

	<-p.MsgIn0.Updates
	assert.Len(t, p.MsgIn1.Updates, 1)
	<-p.MsgIn1.Updates

	// Unsubscribe
	pub.UnsubscribeSync(p.MsgIn1, replyChannel)
	v := <-replyChannel
	//t.Logf("UnsubscribeSync: %v\n", v)

	// Should only write to one queue
	p.MsgOut.Write(messages.Bounce{Name: "bar"})
	<-p.MsgIn0.Updates
	assert.Len(t, p.MsgIn1.Updates, 0)

	// resubscribe
	pub.SubscribeSync(p.MsgIn1, replyChannel)
	v = <-replyChannel
	//t.Logf("ResubscribeSync: %v\n", v)
	assert.Equal(t, 2, v.(int))

}
