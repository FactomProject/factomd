package state_test

import (
	"math/rand"
	"testing"

	"github.com/FactomProject/factomd/testHelper"

	"github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/state"
)

func setNeedsAckToBouce() {
	constants.NeedsAck = func(t byte) bool {
		switch t {
		case constants.BOUNCE_MSG:
			return true
		}
		return false
	}
}

func TestMissingMessageReponseCache(t *testing.T) {
	setNeedsAckToBouce()

	s := testHelper.CreateEmptyTestState()

	// Check the expire expires n-2
	t.Run("ackmap expiration basic", func(t *testing.T) {
		m := state.NewMissingMessageReponseCache(s)
		ackmap := m.AckMessageCache
		// create acks for 1000 heights
		pairs := make([]*state.MsgPair, 1000)
		for i := 0; i < 1000; i++ {
			msg := randomTimedBounceMessage()
			pairs[i] = &state.MsgPair{Ack: ackBounce(i, 0, 0, msg), Msg: msg}
		}

		// Add all pairs and check for expired
		for i := 0; i < 1000; i++ {
			ackmap.AddMsgPair(pairs[i])

			if i > 3 {
				for j := 0; j < j; i++ {
					gPair := ackmap.Get(i-j, 0, 0)
					if j <= 2 {
						if gPair == nil {
							t.Errorf("Could not find message pair when expired")
						}
					} else {
						if gPair != nil {
							t.Errorf("Expected this pair to be expired. It was not. DBHeight now %d, looked for %d", i, i-j)
						}
					}

				}
			}
		}
	})

	// Checks some vm and pl height indexing
	t.Run("ackmap retrieval", func(t *testing.T) {
		numHeights := 30
		numVMs := 10
		numPLHeight := 25
		m := state.NewMissingMessageReponseCache(s)
		ackmap := m.AckMessageCache
		// create acks for 1000 heights
		count := 0
		pairs := make([]*state.MsgPair, numHeights*numVMs*numPLHeight)
		for i := 0; i < numHeights; i++ {
			for j := 0; j < numVMs; j++ {
				for k := 0; k < numPLHeight; k++ {
					msg := randomTimedBounceMessage()
					pairs[count] = &state.MsgPair{Ack: ackBounce(i, j, k, msg), Msg: msg}
					count++
				}
			}
		}

		// Add all pairs and check for expired
		for i := 0; i < len(pairs); i++ {
			pair := pairs[i]
			ackmap.AddMsgPair(pair)

			// Test retrieval for things below the pair
			//	spot check
			for j := 0; j < 5; j++ {
				ack := pair.Ack.(*messages.Ack)
				print(ack.DBHeight)
				rDBHeight := random.RandIntBetween(0, int(ack.DBHeight))
				rVM := random.RandIntBetween(0, int(ack.VMIndex))
				rPL := random.RandIntBetween(0, int(ack.Height))

				gPair := ackmap.Get(rDBHeight, rVM, rPL)
				if int(ack.DBHeight)-rDBHeight <= 2 {
					if gPair == nil {
						t.Errorf("Could not find message pair when expired")
					}
				} else {
					if gPair != nil {
						t.Errorf("Expected this pair to be expired. It was not. DBHeight now %d, looked for %d", i, i-j)
					}
				}
			}
		}
	})

	t.Run("ackmap out of order add", func(t *testing.T) {
		numHeights := 5000
		numMsgs := 5000
		m := state.NewMissingMessageReponseCache(s)
		ackmap := m.AckMessageCache
		// create acks for 1000 heights
		for i := 0; i < numMsgs; i++ {
			msg := randomTimedBounceMessage()
			pair := &state.MsgPair{
				Ack: ackBounce(random.RandIntBetween(0, numHeights),
					random.RandIntBetween(0, 10),
					random.RandIntBetween(0, 10), msg),
				Msg: msg}
			ackmap.AddMsgPair(pair)

			// Check max 3 heights exist
			if len(ackmap.MsgPairMap) > 3 {
				t.Errorf("Expiring did not function right, too many heights kept")
			}

			for dbheight, _ := range ackmap.MsgPairMap {
				if dbheight < ackmap.CurrentWorkingHeight-2 {
					t.Errorf("Working height %d, found %d. Should be expired", ackmap.CurrentWorkingHeight, dbheight)
				}
			}
		}
	})

	//m.Run()
}

func TestSortedMsgCache(t *testing.T) {
	totalLength := 4000
	m := state.NewMsgCache()
	setNeedsAckToBouce()

	// Testing inserting order
	for i := 0; i < totalLength; i++ {
		m.InsertMsg(randomTimedBounceMessage())
	}

	verifyMessageCacheOrder(m, t)

	deleteCount := 500
	// Test single deletes maintain order
	for i := 0; i < deleteCount; i++ {
		index := rand.Intn(totalLength - deleteCount)
		m.RemoveMsg(index)
		totalLength--

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

		if len(m.MessageSlice) != totalLength {
			t.Errorf("Exp length of %d after trim, found %d", totalLength, len(m.MessageSlice))
		}
	}

	// Test time trim
	trimI := totalLength / 2
	trimTime := m.MessageSlice[trimI].GetTimestamp()
	m.TrimOlderThan(trimTime)
	if len(m.MessageSlice) != totalLength-trimI {
		t.Errorf("Exp %d, found %d", totalLength-trimI, len(m.MessageSlice))
	}
	totalLength = totalLength - trimI

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

func verifyMessageCacheOrder(m *state.SortedMsgSlice, t *testing.T) {
	for i, msg := range m.MessageSlice {
		if i == 0 {
			continue
		}

		if msg.GetTimestamp().GetTimeSeconds() < m.MessageSlice[i-1].GetTimestamp().GetTimeSeconds() {
			t.Errorf("Message at index %d has an earlier timestamp than the prior", i)
		}
	}
}

func ackBounce(dbHeight, vmindex, plheight int, bounce interfaces.IMsg) *messages.Ack {
	ack := new(messages.Ack)
	ack.DBHeight = uint32(dbHeight)
	ack.VMIndex = vmindex
	ack.Height = uint32(plheight)

	ack.MessageHash = bounce.GetMsgHash()
	ack.Timestamp = bounce.GetTimestamp()
	return ack
}

func randomTimedBounceMessage() *messages.Bounce {
	b := new(messages.Bounce)
	b.Timestamp = primitives.NewTimestampFromSeconds(random.RandUInt32())
	return b
}
