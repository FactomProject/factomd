// This

package main

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/pubsub"
)

func logf(f string, a ...interface{}) {
	f = f + "\n"
	fmt.Printf(f, a...)
}

func main() {
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

	go func() { // Reader
		timeOut := time.After(1 * time.Second)

		for {
			select {
			case <-timeOut:
				logf("Successfully unsubscribed")
				if p.Count != 5 {
					logf("Unexpected number of messages from subscriber %v", p.Count)
				}
				close(exit)
				return
			case v := <-p.MsgIn.Updates:
				m := v.(messages.Bounce)
				p.Count += 1
				logf("%v: %v - %v", p.Count, m.Name, m.Number)

				if m.Number < 0 {
					// have subscriber disconnect himself
					p.MsgIn.Unsubscribe()
					logf("%v: unsubscribed len: %v", p.Count, len(p.MsgIn.Updates))
				}

				if p.Count > 10 {
					logf("Failed to unsubscribe")
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
					m.Number = -1 // signal client you are bout to disconnect
				} else {
					m.Number = i
				}

				p.MsgOut.Write(m)

				if m.Number < 0 {
					p.MsgOut.Unsubscribe(p.MsgIn) // unsubscribe from this end
				}

				time.Sleep(time.Millisecond * 50)
			}
		}
	}()

	<-exit
}
