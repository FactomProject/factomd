package bmv

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"time"
)

const (
	MsgAdded = 1
	MsgValid = iota * -1
	TimestampExpired
	TimestampTooFuture // We reject things too far in the future
	ReplayMsg          // Msg already seen

)

type MsgReplay struct {
	Buckets    []map[[32]byte]int
	BlockTimes []time.Time
	window     int
}

// MsgReplay is divided into blocks. The windows is the number of valid blocks.
//
// If the window is 6, then there will be 8 buckets. The first 6 are the window of valid
// blocks, anything before the 0th bucket timestamp is expired. The 6th index is the current block.
// The window corresponds to the number of blocks before the current. The 7th index is all future messages
// that fall outside the current block.
func NewMsgReplay(window int) *MsgReplay {
	m := new(MsgReplay)
	m.Buckets = make([]map[[32]byte]int, window+2, window+2)
	// The future block time is unknown
	m.BlockTimes = make([]time.Time, window+1, window+1)
	m.window = window

	return m
}

// Recenter sets the new center for the window to be valid around
func (m *MsgReplay) Recenter(stamp time.Time) {
	if stamp.Before(m.BlockTimes[m.window]) {
		return // We can't go backwards in time
	}

	// [0:window] == past
	// [window]   == previous current
	// [window+1] == future messages to be re-evaluated
	copy(m.BlockTimes, m.BlockTimes[1:m.window+1])
	m.BlockTimes[m.window] = stamp

	copy(m.Buckets, m.Buckets[1:m.window])
	m.Buckets[m.window] = make(map[[32]byte]int)

	// TODO: Re-eval the future bucket
}

// UpdateReplay given a message will return 1 if the message is new, and add it
// to the replay filter. If the message is rejected, it will return < 0.
// The reject codes are constants. A return of 0 should never happen.
func (m *MsgReplay) UpdateReplay(msg interfaces.IMsg) int {
	return m.checkReplay(msg, true)
}

func (m *MsgReplay) checkReplay(msg interfaces.IMsg, update bool) int {
	for i := range m.BlockTimes {
		// TODO: This might be expensive? We should switch to time.Time in the msg so
		//		this conversion is free.
		if msg.GetTimestamp().GetTime().Before(m.BlockTimes[i]) {
			if i == 0 { // Too far in the past
				return TimestampExpired
			}
			// Place the msg into the correct bucket
			_, ok := m.Buckets[i][msg.GetRepeatHash().Fixed()]
			if ok {
				return ReplayMsg
			}
			if update {
				m.Buckets[i][msg.GetRepeatHash().Fixed()] += 1
				return MsgAdded // Added
			}
			return MsgValid // Found, but not updated
		}
	}

	// Msg is from the future
	// TODO: Handle future messages
	return TimestampTooFuture
}

// IsReplay returns the same error codes as UpdateReplay, but will return a 0 if the
// message is not found in the filter and is valid.
func (m *MsgReplay) IsReplay(msg interfaces.IMsg) int {
	return m.checkReplay(msg, false)
}
