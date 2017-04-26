// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package meta_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/meta"
)

func TestCheckExternalIDsLength(t *testing.T) {
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
