package bmv_test

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"

	. "github.com/FactomProject/factomd/modules/bmv"
)

func TestMsgReplay_UpdateReplay(t *testing.T) {
	initR := func() *MsgReplay {
		r := NewMsgReplay(1)
		// Add the 2 windows
		r.Recenter(time.Now().Add(-60 * time.Second))
		r.Recenter(time.Now())
		return r
	}

	t.Run("test timeout", func(t *testing.T) {
		r := initR()
		v := r.UpdateReplay(bounce(time.Now().Add(-120 * time.Second)))
		if v != TimestampExpired {
			t.Errorf("expected expired, found %d", v)
		}
	})

	t.Run("test exists", func(t *testing.T) {
		r := initR()
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

	t.Run("test recenter", func(t *testing.T) {
		for i := 10; i < 20; i++ {
			r := initR()
			msg := bounce(time.Now().Add(time.Minute * 11))
			v := r.UpdateReplay(msg)
			if v != MsgAdded {
				t.Errorf("expected added, found %d", v)
			}

			// Check the replay is still caught for future
			v = r.UpdateReplay(msg)
			if v != ReplayMsg {
				t.Errorf("expected replay, found %d", v)
			}

			// Update the latest window to put the msg into the current block
			r.Recenter(time.Now().Add(time.Minute * time.Duration(i)))
			// Ensure the msg was shifted
			v = r.UpdateReplay(msg)
			if v != ReplayMsg {
				t.Errorf("%d: expected replay, found %d", i, v)
			}
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
