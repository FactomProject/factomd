// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"fmt"
	"testing"

	. "github.com/PaulSnow/factom2d/common/primitives"
)

func TestUnmarshalNilTimestamp(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(Timestamp)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestTimestamp(t *testing.T) {
	ts := new(Timestamp)
	ts.SetTimeNow()
	fmt.Printf("ts: %d, milli: %d seconds %d", *ts, ts.GetTimeMilli(), ts.GetTimeSeconds())
	hex, err := ts.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	if len(hex) != 6 {
		t.Errorf("Wrong length of marshalled timestamp - %x", hex)
	}

	ts2 := new(Timestamp)
	rest, err := ts2.UnmarshalBinaryData(hex)
	if err != nil {
		t.Error(err)
	}

	if len(rest) > 0 {
		t.Errorf("Leftover data from unmarshalling - %x", rest)
	}

	if *ts-*ts2 != 0 {
		t.Errorf("Timestamps don't match up - %d vs %d", *ts, *ts2)
	}

	ts2 = new(Timestamp)
	rest, err = ts2.UnmarshalBinaryData(append(hex, 0x01))
	if err != nil {
		t.Error(err)
	}

	if len(rest) != 1 {
		t.Errorf("Leftover data from unmarshalling - %x", rest)
	}

	if *ts-*ts2 != 0 {
		t.Errorf("Timestamps don't match up - %d vs %d", *ts, *ts2)
	}
}

func TestTimestamp2(t *testing.T) {
	ts := new(Timestamp)
	ts.SetTime(0xFF22100122FF)
	hex, err := ts.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	if len(hex) != 6 {
		t.Errorf("Wrong length of marshalled timestamp - %x", hex)
	}

	ts2 := new(Timestamp)
	rest, err := ts2.UnmarshalBinaryData(hex)
	if err != nil {
		t.Error(err)
	}

	if len(rest) > 0 {
		t.Errorf("Leftover data from unmarshalling - %x", rest)
	}

	if *ts-*ts2 != 0 {
		t.Errorf("Timestamps don't match up - %d vs %d", *ts, *ts2)
	}

	ts2 = new(Timestamp)
	rest, err = ts2.UnmarshalBinaryData(append(hex, 0x01))
	if err != nil {
		t.Error(err)
	}

	if len(rest) != 1 {
		t.Errorf("Leftover data from unmarshalling - %x", rest)
	}

	if *ts-*ts2 != 0 {
		t.Errorf("Timestamps don't match up - %d vs %d", *ts, *ts2)
	}
}

func TestTimestamp3(t *testing.T) {
	ts := new(Timestamp)
	ts.SetTime(0x000001)
	hex, err := ts.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Marshalled - %x", hex)

	if len(hex) != 6 {
		t.Errorf("Wrong length of marshalled timestamp - %x", hex)
	}

	ts2 := new(Timestamp)
	rest, err := ts2.UnmarshalBinaryData(hex)
	if err != nil {
		t.Error(err)
	}

	if len(rest) > 0 {
		t.Errorf("Leftover data from unmarshalling - %x", rest)
	}

	if *ts-*ts2 != 0 {
		t.Errorf("Timestamps don't match up - %d vs %d", *ts, *ts2)
	}

	ts2 = new(Timestamp)
	rest, err = ts2.UnmarshalBinaryData(append(hex, 0x01))
	if err != nil {
		t.Error(err)
	}

	if len(rest) != 1 {
		t.Errorf("Leftover data from unmarshalling - %x", rest)
	}

	if *ts-*ts2 != 0 {
		t.Errorf("Timestamps don't match up - %d vs %d", *ts, *ts2)
	}
}

func TestTimestampMisc(t *testing.T) {
	for i := 0; i < 1000; i++ {
		ts := NewTimestampNow()
		ts2 := NewTimestampFromMilliseconds(ts.GetTimeMilliUInt64())
		ts3 := NewTimestampFromSeconds(uint32(ts.GetTimeSeconds()))
		ts4 := NewTimestampFromMinutes(ts.GetTimeMinutesUInt32())
		ts5 := NewTimestampFromMilliseconds(0)
		ts5.SetTimestamp(ts)

		if ts.String() != ts2.String() {
			t.Errorf("Timestamps are not identical")
		}
		if ts.String() != ts3.String() {
			t.Errorf("Timestamps are not identical")
		}
		if ts.String() != ts5.String() {
			t.Errorf("Timestamps are not identical")
		}

		if ts.GetTimeMilliUInt64() != ts2.GetTimeMilliUInt64() {
			t.Errorf("Timestamps are not identical")
		}
		if ts.GetTimeSeconds() != ts3.GetTimeSeconds() {
			t.Errorf("Timestamps are not identical")
		}
		if ts.GetTimeSecondsUInt32() != ts3.GetTimeSecondsUInt32() {
			t.Errorf("Timestamps are not identical")
		}
		if ts.GetTimeMinutesUInt32() != ts4.GetTimeMinutesUInt32() {
			t.Errorf("Timestamps are not identical")
		}
		if ts.GetTimeMilliUInt64() != ts5.GetTimeMilliUInt64() {
			t.Errorf("Timestamps are not identical")
		}
		if ts.UTCString() != ts5.UTCString() {
			t.Errorf("Timestamps are not identical")
		}
	}
}
