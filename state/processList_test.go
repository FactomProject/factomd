// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	. "github.com/FactomProject/factomd/state"
	"testing"
)

func TestGetLen(t *testing.T) {

}

func createProcessList() *ProcessList {
	p := NewProcessList(2, 10)
	return p
}
