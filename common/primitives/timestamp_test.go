// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/common/primitives"
)

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
