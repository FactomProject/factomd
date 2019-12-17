package pubsub_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
	. "github.com/FactomProject/factomd/pubsub"
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
