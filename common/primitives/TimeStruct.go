// Decode a variable integer from the given data buffer.
// Returns the uint64 bit value and a data slice positioned
// after the variable integer

package primitives

// This structure watches the time, and tosses stuff that hasn't settled within the
// time limits (with some room for error).  Also

import ()

// We only allow timestamps that are +/- 12 hours.  But followers need to keep
// structures anyway.  So we are conservative, and allow +/- 13 hours
const (
	period = 13
)

type hash2obj map[[32]byte]interface{}

type TimeStruct struct {
	values [period * 2]*hash2obj
	time   int64
}

func (ts *TimeStruct) Init(time int64) {
	for i := 0; i < period*2; i++ {
		ts.values[i] = new(hash2obj)
	}
}
