package state_test

import (
	"math/rand"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/state"
)

func TestSortedMsgCache(t *testing.T) {
	totalLength := 4000
	m := state.NewMsgCache()
	constants.NeedsAck = func(t byte) bool {
		switch t {
		case constants.BOUNCE_MSG:
			return true
		}
		return false
	}

	// Testing inserting order
	for i := 0; i < totalLength; i++ {
		m.AddMsg(randomTimedBounceMessage())
	}

	verifyMessageCacheOrder(m, t)

	deleteCount := 500
	// Test single deletes maintain order
	for i := 0; i < deleteCount; i++ {
		index := rand.Intn(totalLength - deleteCount)
		msg := m.MessageSlice[index]
		m.RemoveMsg(index)
		totalLength--
		if _, ok := m.MessageMap[msg.GetMsgHash().Fixed()]; ok {
			t.Errorf("The removed message still exists in the map")
		}
		if len(m.MessageSlice) != totalLength {
			t.Errorf("Exp length of %d after delete, found %d", totalLength-i-1, len(m.MessageSlice))
		}
		verifyMessageCacheOrder(m, t)
	}

	// Test the trim
	trimCount := 10
	trimAmount := 100
	for i := 0; i < trimCount; i++ {
		deleted := make([]interfaces.IMsg, trimAmount)
		copy(deleted, m.MessageSlice[:trimAmount])
		m.TrimTo(trimAmount)
		totalLength -= trimAmount

		for _, msg := range deleted {
			if _, ok := m.MessageMap[msg.GetMsgHash().Fixed()]; ok {
				t.Errorf("The removed message still exists in the map")
			}
		}
		if len(m.MessageSlice) != totalLength {
			t.Errorf("Exp length of %d after trim, found %d", totalLength, len(m.MessageSlice))
		}
	}

	// Test the 0 trim
	m.TrimTo(0)
	if len(m.MessageSlice) != totalLength {
		t.Errorf("Exp length of %d after trim, found %d", totalLength, len(m.MessageSlice))
	}

	// Test the full trim
	m.TrimTo(len(m.MessageSlice))
	if len(m.MessageSlice) != 0 {
		t.Errorf("Exp length of %d after trim, found %d", totalLength, len(m.MessageSlice))
	}
}

func verifyMessageCacheOrder(m *state.MsgCache, t *testing.T) {
	for i, msg := range m.MessageSlice {
		if i == 0 {
			continue
		}

		if msg.GetTimestamp().GetTimeSeconds() < m.MessageSlice[i-1].GetTimestamp().GetTimeSeconds() {
			t.Errorf("Message at index %d has an earlier timestamp than the prior", i)
		}
	}
}

func randomTimedBounceMessage() *messages.Bounce {
	b := new(messages.Bounce)
	b.Timestamp = primitives.NewTimestampFromSeconds(random.RandUInt32())
	return b
}
