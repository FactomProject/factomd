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
