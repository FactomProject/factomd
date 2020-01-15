package pubsub_test

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/pubsub"
	"testing"
	"time"
)

func TestSubMsgFilterWrap(t *testing.T) {
	ResetGlobalRegistry()

	p := PubFactory.Base().Publish("test")
	s := SubFactory.Value().Subscribe("test", SubMsgFilterWrap(constants.BOUNCE_MSG))

	p.Write(new(messages.MissingMsg))
	if s.Read() != nil {
		t.Error("Expected a nil")
	}

	p.Write(new(messages.Bounce))
	if v := s.Read(); v == nil {
		t.Error("Expected a msg")
	} else {
		msg, ok := v.(interfaces.IMsg)
		if !ok {
			t.Error("should be a msg")
		}
		if msg.Type() != constants.BOUNCE_MSG {
			t.Error("type is wrong")
		}
	}
}

func TestSubUnSub(t *testing.T) {
	ResetGlobalRegistry()
	nodeName := "FNode0"

	type Module struct {
		Path   string
		MsgOut IPublisher
		MsgIn  *SubChannel
		Count  int
	}

	p := &Module{Path: GetPath(nodeName, "test")}
	p.MsgOut = PubFactory.Threaded(1).Publish(p.Path)
	go p.MsgOut.Start()

	p.MsgIn = SubFactory.Channel(1)
	p.MsgIn.Subscribe(GetPath(nodeName, "test"))

	exit := make(chan interface{})

	drain := func() { // purge channel
		for {
			select {
			case <-p.MsgIn.Updates:
				// drain
			default:
				return
			}
		}
	}
	_ = drain

	go func() { // Reader
		timeOut := time.After(1 * time.Second)

		for {
			select {
			case <-timeOut:
				t.Log("Successfully unsubscribed")
				if p.Count != 5 {
					t.Errorf("Unexpected number of messages from subscriber %v", p.Count)
				}
				//  pass
				close(exit)
				return
			case v := <-p.MsgIn.Updates:
				m := v.(messages.Bounce)
				p.Count += 1
				t.Logf("%v: %v - %v", p.Count, m.Name, m.Number)

				if m.Number < 0 {
					p.MsgIn.Unsubscribe()
					t.Logf("%v: unsubscribed len: %v", p.Count, len(p.MsgIn.Updates))
					//drain() // purge remaining updates
				}

				if p.Count > 10 {
					t.Errorf("Failed to unsubscribe")
					close(exit)
					return

				}
			}
		}
	}()

	go func() { // writer
		var i int32 = 0
		m := messages.Bounce{Name: "Test"}

		for {
			select {
			case <-exit:
				return
			default:
				i += 1
				if i == 5 {
					m.Number = -1 // send disconnect
				} else {
					m.Number = i
				}
				p.MsgOut.Write(m)
				time.Sleep(time.Millisecond * 50)
			}
		}
	}()

	<-exit
}
