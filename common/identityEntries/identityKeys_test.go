// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identityEntries_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/identityEntries"
	"github.com/FactomProject/factomd/common/primitives/random"
)

func TestCheckExternalIDsLength(t *testing.T) {
	for i := 0; i < 1000; i++ {
		l := random.RandIntBetween(0, 100)
		extIDs := [][]byte{}
		lens := []int{}
		for j := 0; j < l; j++ {
			b := random.RandByteSlice()
			extIDs = append(extIDs, b)
			lens = append(lens, len(b))
		}
		if CheckExternalIDsLength(extIDs, lens) == false {
			t.Errorf("Wrong CheckExternalIDsLength response")
		}

		if len(lens) > 0 {
			lens[0]++
			if CheckExternalIDsLength(extIDs, lens) == true {
				t.Errorf("Wrong CheckExternalIDsLength response")
			}

			lens = lens[1:]
			if CheckExternalIDsLength(extIDs, lens) == true {
				t.Errorf("Wrong CheckExternalIDsLength response")
			}

			extIDs = extIDs[1:]
			if CheckExternalIDsLength(extIDs, lens) == false {
				t.Errorf("Wrong CheckExternalIDsLength response")
			}
		}
	}

	extIDs := [][]byte{
		{0x00, 0x00, 0x00, 0x00, 0x00},                               // 5
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // 10
		{0x00, 0x00, 0x00},                                           // 3
		{0x00},                                                       // 1
		{},                                                           // 0
	}
	lengths := []int{5, 10, 3, 1, 0}
	lengthsBad := []int{5, 10, 3, 1, 1}
	if CheckExternalIDsLength(extIDs, lengthsBad) {
		t.Error("1: CheckExternalIDsLength check failed")
	}

	lengthsBad = []int{}
	if CheckExternalIDsLength(extIDs, lengthsBad) {
		t.Error("2: CheckExternalIDsLength check failed")
	}

	lengthsBad = []int{5, 10, 3, 1}
	if CheckExternalIDsLength(extIDs, lengthsBad) {
		t.Error("3: CheckExternalIDsLength check failed")
	}

	if !CheckExternalIDsLength(extIDs, lengths) {
		t.Error("4: CheckExternalIDsLength check failed")
	}
}
