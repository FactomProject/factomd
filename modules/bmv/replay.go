package bmv

import (
	"sync"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
)

const (
	MsgValid           = iota * -1 // +0 |
	TimestampExpired               // -1 |
	TimestampTooFuture             // -2 | We reject things too far in the future
	ReplayMsg                      // -3 | Msg already seen

	// If the msg is added to the filter
	MsgAdded = 1
)

type MsgReplay struct {
	bucketsMtx sync.RWMutex // protects buckets and oldest
	buckets    []*bucket
	oldest     time.Time

	// helper indices for readability
	current int
	future  int

	blocktime time.Duration // Minimum block time

}

// MsgReplay is divided into blocks. The window is the number of valid blocks.
//
// If the window is 6, then there will be 8 buckets. The first 6 are the window of valid
// blocks, anything before the 0th bucket timestamp is expired. The 6th index is the current block.
// The window corresponds to the number of blocks before the current. The 7th index is all future messages
// that fall outside the current block.
func NewMsgReplay(window int, blocktime time.Duration) *MsgReplay {
	m := new(MsgReplay)
	m.buckets = make([]*bucket, window+2, window+2)
	for i := range m.buckets {
		m.buckets[i] = newBucket()
	}
	m.current = window
	m.future = window + 1
	m.blocktime = blocktime

	return m
}

// Recenter sets the new center for the window to be valid around
func (m *MsgReplay) Recenter(stamp time.Time) {
	m.bucketsMtx.Lock()
	defer m.bucketsMtx.Unlock()

	if stamp.Before(m.oldest) {
		return // We can't go backwards in time
	}

	// [0:window] == past
	// [window]   == current
	// [window+1] == future messages to be re-evaluated

	copy(m.buckets, m.buckets[1:])
	m.buckets[m.current].SetTime(stamp)
	m.oldest = m.buckets[0].Time()

	m.buckets[m.future] = newBucket()

	estimatedEnd := stamp.Add(m.blocktime)
	for key, ts := range m.Buckets[m.windows+1].data {
		if ts.Before(m.BlockTimes[m.windows]) {
			// The future msg falls behind the current block time.
			// So shift it back into the prior block's.
			m.Buckets[m.windows-1].data[key] = ts
			delete(m.Buckets[m.windows+1].data, key)
		} else if ts.Before(estimatedEnd) {
			// The future msg falls into the current block
			m.Buckets[m.windows].data[key] = ts
			delete(m.Buckets[m.windows+1].data, key)
		}
	}
}

// UpdateReplay given a message will return 1 if the message is new, and add it
// to the replay filter. If the message is rejected, it will return < 0.
// The reject codes are constants. A return of 0 should never happen.
func (m *MsgReplay) UpdateReplay(msg interfaces.IMsg) int {
	return m.checkReplay(msg, true)
}

func (m *MsgReplay) checkReplay(msg interfaces.IMsg, update bool) int {
	m.bucketsMtx.RLock()
	defer m.bucketsMtx.RUnlock()
	var index int = -1 // Index of the bucket to check against

	// TODO: .GetTime() might be expensive? We should switch to time.Time in the msg so
	//		this conversion is free.
	mTime := msg.GetTimestamp().GetTime()

	// First see if the this msg is from the past
	for i := range m.BlockTimes {
		if mTime.Before(m.BlockTimes[i]) {
			if i == 0 { // Too far in the past
				return TimestampExpired
			}
			// Place the msg into the correct bucket
			index = i - 1 // Bucket[i-1] is the right bucket
			break
		}
	}

	if index == -1 {
		// Msg is from the future or the current
		if mTime.Before(m.BlockTimes[m.windows].Add(m.blocktime)) {
			// Current
			index = m.windows
		} else {
			// Future
			// TODO: Handle future messages
			index = m.windows + 1
			// return TimestampTooFuture
		}
	}

	if index != -1 {
		_, ok := m.Buckets[index].data[msg.GetRepeatHash().Fixed()]
		if ok {
			return ReplayMsg
		}
		if update {
			m.Buckets[index].data[msg.GetRepeatHash().Fixed()] = mTime
			return MsgAdded // Added
		}
	}

	return MsgValid // Found, but not updated
}

// IsReplay returns the same error codes as UpdateReplay, but will return a 0 if the
// message is not found in the filter and is valid.
func (m *MsgReplay) IsReplay(msg interfaces.IMsg) int {
	return m.checkReplay(msg, false)
}
