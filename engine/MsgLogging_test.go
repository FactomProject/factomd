package engine_test

import (
	"testing"
	"time"

	"github.com/PaulSnow/factom2d/common/messages"
	"github.com/PaulSnow/factom2d/common/primitives"
	. "github.com/PaulSnow/factom2d/engine"
	"github.com/PaulSnow/factom2d/state"
)

func TestMessageLoging(t *testing.T) {
	msgLog := new(MsgLog)
	msgLog.Init(true, 0)

	fnode := new(FactomNode)
	fnode.State = new(state.State)
	s := fnode.State

	msg := new(messages.Bounce)
	msg.Name = "bob"
	msg.Timestamp = primitives.NewTimestampNow()
	msg.Data = []byte("here is some data")
	msg.Stamps = append(msg.Stamps, primitives.NewTimestampNow())

	msgLog.PrtMsgs(s)

	msgLog.Add2(fnode, true, "peer", "where", true, msg)
	msgLog.Add2(fnode, false, "peer", "where", true, msg)

	msgLog.Startp = primitives.NewTimestampFromMilliseconds(0)
	msgLog.Add2(fnode, false, "peer", "where", true, msg)

	if len(msgLog.MsgList) != 3 {
		t.Error("Should have three entries")
	}
	msgLog.PrtMsgs(s)

	msgLog.Last.SetTimeSeconds(msgLog.Last.GetTimeSeconds() - 6)
	time.Sleep(10 * time.Millisecond)
	msgLog.Add2(fnode, false, "peer", "where", true, msg)

	if len(msgLog.MsgList) != 0 {
		t.Error("Should have zero messages")
	}
	msgLog.PrtMsgs(s)
}
