package bmv_test

import (
	"crypto/rand"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/modules/bmv"
)

func TestMsgReplay_UpdateReplay(t *testing.T) {
	r := NewMsgReplay(1)
	// Add the 2 windows
	r.Recenter(time.Now().Add(-60 * time.Second))
	r.Recenter(time.Now())

	t.Run("test timeout", func(t *testing.T) {
		v := r.UpdateReplay(bounce(time.Now().Add(-120 * time.Second)))
		if v != TimestampExpired {
			t.Errorf("expected expired, found %d", v)
		}
	})

	t.Run("test exists", func(t *testing.T) {
		msg := bounce(time.Now())
		v := r.UpdateReplay(msg)
		if v != MsgAdded {
			t.Errorf("expected added, found %d", v)
		}

		v = r.UpdateReplay(msg)
		if v != ReplayMsg {
			t.Errorf("expected replay, found %d", v)
		}
	})
}

func bounce(t time.Time) *messages.Bounce {
	b := new(messages.Bounce)
	b.Data = make([]byte, 100)
	_, _ = rand.Read(b.Data)
	b.Timestamp = primitives.NewTimestampFromSeconds(uint32(t.Unix()))

	return b
}
