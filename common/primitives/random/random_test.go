// +build all

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package random_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/primitives/random"
)

func TestRandUInt64Between(t *testing.T) {
	var min uint64 = 100
	var max uint64 = 200

	for i := 0; i < 1000; i++ {
		r := RandUInt64Between(min, max)
		if r < min {
			t.Errorf("Returned value smaller than min - %v < %v", r, min)
		}
		if r > max {
			t.Errorf("Returned value greater than max - %v > %v", r, max)
		}

		if RandUInt64Between(10, 0) != 0 {
			t.Errorf("Returned a wrong value for invalid input")
		}
	}
}

func TestRandInt64Between(t *testing.T) {
	var min int64 = -100
	var max int64 = 200

	for i := 0; i < 1000; i++ {
		r := RandInt64Between(min, max)
		if r < min {
			t.Errorf("Returned value smaller than min - %v < %v", r, min)
		}
		if r > max {
			t.Errorf("Returned value greater than max - %v > %v", r, max)
		}

		if RandInt64Between(10, 0) != 0 {
			t.Errorf("Returned a wrong value for invalid input")
		}
	}
}

func TestRandIntBetween(t *testing.T) {
	var min int = -100
	var max int = 200

	for i := 0; i < 1000; i++ {
		r := RandIntBetween(min, max)
		if r < min {
			t.Errorf("Returned value smaller than min - %v < %v", r, min)
		}
		if r > max {
			t.Errorf("Returned value greater than max - %v > %v", r, max)
		}

		if RandIntBetween(10, 0) != 0 {
			t.Errorf("Returned a wrong value for invalid input")
		}
	}
}

func TestRandByteSlice(t *testing.T) {
	for i := 0; i < 1000; i++ {
		r := RandByteSlice()

		if len(r) > 64 {
			t.Errorf("Returned a wrong size")
		}
	}
}

func TestRandNonEmptyByteSlice(t *testing.T) {
	for i := 0; i < 1000; i++ {
		r := RandNonEmptyByteSlice()

		if len(r) > 64 || len(r) == 0 {
			t.Errorf("Returned a wrong size %v", len(r))
		}
	}
}

func TestRandByteSliceOfLen(t *testing.T) {
	for i := 0; i < 1000; i++ {
		r := RandByteSliceOfLen(100)

		if len(r) != 100 {
			t.Errorf("Returned a wrong size %v", len(r))
		}
	}
}

func TestRandomString(t *testing.T) {
	for i := 0; i < 100; i++ {
		r := RandomString()

		if len(r) > 128 {
			t.Errorf("Returned a wrong size %v", len(r))
		}
	}
}

func TestUintNumbers(t *testing.T) {
	for i := 0; i < 1000; i++ {
		r := RandUInt64()

		if r < 0 {
			t.Errorf("Returned a negative %v", r)
		}
	}
	for i := 0; i < 1000; i++ {
		s := RandUInt32()

		if s < 0 {
			t.Errorf("Returned a negative %v", s)
		}
	}
	for i := 0; i < 1000; i++ {
		u := RandUInt8()

		if u < 0 {
			t.Errorf("Returned a negative %v", u)
		}
	}
	for i := 0; i < 1000; i++ {
		v := RandByte()

		if v < 0 {
			t.Errorf("Returned a negative %v", v)
		}
	}
}
