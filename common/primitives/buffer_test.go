// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"math"
	"testing"

	. "github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

func TestPushPopInt64(t *testing.T) {
	b := NewBuffer(nil)

	var i int64
	err := b.PushInt64(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err := b.PopInt64()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	i = math.MaxInt64
	err = b.PushInt64(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err = b.PopInt64()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	i = math.MinInt64
	err = b.PushInt64(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err = b.PopInt64()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	for j := 0; j < 1000; j++ {
		i = random.RandInt64()
		err = b.PushInt64(i)
		if err != nil {
			t.Errorf("%v", err)
		}
		r, err = b.PopInt64()
		if err != nil {
			t.Errorf("%v", err)
		}
		if i != r {
			t.Errorf("Received wrong number - %v vs %v", i, r)
		}
	}
}

func TestPushPopUInt64(t *testing.T) {
	b := NewBuffer(nil)

	var i uint64
	err := b.PushUInt64(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err := b.PopUInt64()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	i = math.MaxInt64
	err = b.PushUInt64(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err = b.PopUInt64()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	for j := 0; j < 1000; j++ {
		i = random.RandUInt64()
		err = b.PushUInt64(i)
		if err != nil {
			t.Errorf("%v", err)
		}
		r, err = b.PopUInt64()
		if err != nil {
			t.Errorf("%v", err)
		}
		if i != r {
			t.Errorf("Received wrong number - %v vs %v", i, r)
		}
	}
}

func TestPushPopVArInt(t *testing.T) {
	b := NewBuffer(nil)

	var i uint64
	err := b.PushVarInt(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err := b.PopVarInt()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	i = math.MaxInt64
	err = b.PushVarInt(i)
	if err != nil {
		t.Errorf("%v", err)
	}
	r, err = b.PopVarInt()
	if err != nil {
		t.Errorf("%v", err)
	}
	if i != r {
		t.Errorf("Received wrong number - %v vs %v", i, r)
	}

	for j := 0; j < 1000; j++ {
		i = random.RandUInt64()
		err = b.PushVarInt(i)
		if err != nil {
			t.Errorf("%v", err)
		}
		r, err = b.PopVarInt()
		if err != nil {
			t.Errorf("%v", err)
		}
		if i != r {
			t.Errorf("Received wrong number - %v vs %v", i, r)
		}
	}
}
