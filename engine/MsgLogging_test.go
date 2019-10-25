package engine_test

import (
	"github.com/FactomProject/factomd/fnode"
	"testing"
	"time"

	"github.com/FactomProject/factomd/fnode"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
)

func TestMessageLoging(t *testing.T) {
	msgLog := new(MsgLog)
	msgLog.Init(true, 0)

	node := new(fnode.FactomNode)
	node.State = new(state.State)
	s := node.State

	msg := new(messages.Bounce)
	msg.Name = "bob"
	msg.Timestamp = primitives.NewTimestampNow()
	msg.Data = []byte("here is some data")
	msg.Stamps = append(msg.Stamps, primitives.NewTimestampNow())

	msgLog.PrtMsgs(s)

	msgLog.Add2(node, true, "peer", "where", true, msg)
	msgLog.Add2(node, false, "peer", "where", true, msg)

	msgLog.Startp = primitives.NewTimestampFromMilliseconds(0)
	msgLog.Add2(node, false, "peer", "where", true, msg)

	if len(msgLog.MsgList) != 3 {
		t.Error("Should have three entries")
	}
	msgLog.PrtMsgs(s)

	msgLog.Last.SetTimeSeconds(msgLog.Last.GetTimeSeconds() - 6)
	time.Sleep(10 * time.Millisecond)
	msgLog.Add2(node, false, "peer", "where", true, msg)

	if len(msgLog.MsgList) != 0 {
		t.Error("Should have zero messages")
	}
	msgLog.PrtMsgs(s)
}
