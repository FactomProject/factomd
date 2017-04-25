package state_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/state"
)

var _ = fmt.Println

func TestQueues(t *testing.T) {
	RegisterPrometheus()
	RegisterPrometheus()
	channel := make(chan interfaces.IMsg, 1000)
	general := GeneralMSGQueue(channel)
	inmsg := InMsgMSGQueue(channel)
	netOut := NetOutMsgQueue(channel)

	if !checkLensAndCap(channel, []interfaces.IQueue{general, inmsg, netOut}) {
		t.Error("Error: Lengths/Cap does not match")
	}

	c := 0
	for i := 0; i < 100; i++ {
		switch c {
		case 0:
			channel <- nil
		case 1:
			general.Enqueue(new(messages.DBStateMsg))
		case 2:
			inmsg.Enqueue(nil)
		}
		c++
		if c == 3 {
			c = 0
		}
		if !checkLensAndCap(channel, []interfaces.IQueue{general, inmsg, netOut}) {
			t.Error("Error: Lengths/Cap does not match")
		}

	}

	for i := 0; i < 100; i++ {
		switch c {
		case 0:
			<-channel
		case 1:
			general.Dequeue()
		case 2:
			inmsg.Dequeue()
		}
		c++
		if c == 3 {
			c = 0
		}
		if !checkLensAndCap(channel, []interfaces.IQueue{general, inmsg, netOut}) {
			t.Error("Error: Lengths/Cap does not match")
		}
	}

	if len(channel) != 0 {
		t.Errorf("Channel should be 0, found %d", len(channel))
	}

	// Check for blocking
	select {
	case <-channel:
	default:
	}
	go func() {
		time.Sleep(1100 * time.Millisecond)
		general.Enqueue(nil)
		inmsg.Enqueue(nil)
		netOut.Enqueue(nil)
	}()

	b := time.Now().Unix()
	general.BlockingDequeue()
	if time.Now().Unix()-b < 1 {
		t.Error("Did not properly block")
	}

	inmsg.BlockingDequeue()
	if time.Now().Unix()-b < 1 {
		t.Error("Did not properly block")
	}

	netOut.BlockingDequeue()
	if time.Now().Unix()-b < 1 {
		t.Error("Did not properly block")
	}

	// Trip prometheus, unfortunately, we cannot actually check the values
	tripAllMessages(inmsg)
	tripAllMessages(general)
	tripAllMessages(netOut)

	if len(channel) != 0 {
		t.Errorf("Channel should be 0, found %d", len(channel))
	}
	if !checkLensAndCap(channel, []interfaces.IQueue{general, inmsg, netOut}) {
		t.Error("Error: Lengths/Cap does not match")
	}
}

func tripAllMessages(q interfaces.IQueue) {
	q.Enqueue(new(messages.EOM))
	q.Enqueue(new(messages.Ack))
	q.Enqueue(new(messages.AuditServerFault))
	q.Enqueue(new(messages.ServerFault))
	q.Enqueue(new(messages.FullServerFault))
	q.Enqueue(new(messages.CommitChainMsg))
	q.Enqueue(new(messages.CommitEntryMsg))
	q.Enqueue(new(messages.DirectoryBlockSignature))
	q.Enqueue(new(messages.EOMTimeout))
	q.Enqueue(new(messages.Heartbeat))
	q.Enqueue(new(messages.InvalidDirectoryBlock))
	q.Enqueue(new(messages.MissingMsg))
	q.Enqueue(new(messages.MissingMsgResponse))
	q.Enqueue(new(messages.MissingData))
	q.Enqueue(new(messages.RevealEntryMsg))
	q.Enqueue(new(messages.DBStateMsg))
	q.Enqueue(new(messages.DBStateMissing))
	q.Enqueue(new(messages.Bounce))
	q.Enqueue(new(messages.BounceReply))
	q.Enqueue(new(messages.SignatureTimeout))

	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
	q.Dequeue()
}

func checkLensAndCap(channel chan interfaces.IMsg, qs []interfaces.IQueue) bool {
	for _, q := range qs {
		if len(channel) != q.Length() {
			return false
		}
	}
	return true
}
