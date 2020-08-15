package state_test

import (
	"testing"

	"github.com/PaulSnow/factom2d/testHelper"

	"github.com/PaulSnow/factom2d/common/interfaces"

	"github.com/PaulSnow/factom2d/common/messages"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/common/primitives/random"

	"github.com/PaulSnow/factom2d/state"
)

func TestMissingMessageReponseCache(t *testing.T) {
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
