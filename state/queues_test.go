package state_test

import (
	"fmt"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/state"
)

var _ = fmt.Println

func TestQueues(t *testing.T) {
	RegisterPrometheus()
	channel := make(chan interfaces.IMsg, 1000)
	general := GeneralMSGQueue(channel)
	inmsg := InMsgMSGQueue{GeneralMSGQueue: general}

	if !checkLensAndCap(channel, general, inmsg) {
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
		if !checkLensAndCap(channel, general, inmsg) {
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
		if !checkLensAndCap(channel, general, inmsg) {
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
	general.Dequeue()
	inmsg.Dequeue()

	// Trip prometheus, unfortunately, we cannot actually check the values
	inmsg.Enqueue(new(messages.EOM))
	inmsg.Enqueue(new(messages.Ack))
	inmsg.Enqueue(new(messages.AuditServerFault))
	inmsg.Enqueue(new(messages.ServerFault))
	inmsg.Enqueue(new(messages.FullServerFault))
	inmsg.Enqueue(new(messages.CommitChainMsg))
	inmsg.Enqueue(new(messages.CommitEntryMsg))
	inmsg.Enqueue(new(messages.DirectoryBlockSignature))
	inmsg.Enqueue(new(messages.EOMTimeout))
	inmsg.Enqueue(new(messages.Heartbeat))
	inmsg.Enqueue(new(messages.InvalidDirectoryBlock))
	inmsg.Enqueue(new(messages.MissingMsg))
	inmsg.Enqueue(new(messages.MissingMsgResponse))
	inmsg.Enqueue(new(messages.MissingData))
	inmsg.Enqueue(new(messages.RevealEntryMsg))
	inmsg.Enqueue(new(messages.DBStateMsg))
	inmsg.Enqueue(new(messages.DBStateMissing))
	inmsg.Enqueue(new(messages.Bounce))
	inmsg.Enqueue(new(messages.BounceReply))
	inmsg.Enqueue(new(messages.SignatureTimeout))

	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()
	inmsg.Dequeue()

	if len(channel) != 0 {
		t.Errorf("Channel should be 0, found %d", len(channel))
	}
	if !checkLensAndCap(channel, general, inmsg) {
		t.Error("Error: Lengths/Cap does not match")
	}
}

func checkLensAndCap(channel chan interfaces.IMsg, gen GeneralMSGQueue, in InMsgMSGQueue) bool {
	if len(channel) != gen.Length() {
		return false
	}

	if len(channel) != in.Length() {
		return false
	}

	if cap(channel) != gen.Cap() {
		return false
	}

	if cap(channel) != in.Cap() {
		return false
	}
	return true
}
