// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

func TestGettingFromProcessLists(t *testing.T) {
	state := testHelper.CreateEmptyTestState()
	plls := NewProcessLists(state)

	pl := plls.Get(0)
	assertPllsLength(t, plls, 1)
	assertPlExists(t, pl)

	pl = plls.Get(1)
	assertPllsLength(t, plls, 2)
	assertPlExists(t, pl)

	pl = plls.Get(10)
	assertPllsLength(t, plls, 11)
	assertPlExists(t, pl)

	pl = plls.Get(199)
	assertPllsLength(t, plls, 200)
	assertPlExists(t, pl)

	pl = plls.Get(210)
	if pl != nil {
		t.Error("process lists are only created for height +200 max")
	}
}

func assertPlExists(t *testing.T, item *ProcessList) {
	if item == nil {
		t.Error(fmt.Sprintf("Item %v should exists, but got nil", item))
	}
}

func assertPlDoesNotExist(t *testing.T, item *ProcessList) {
	if item != nil {
		t.Error(fmt.Sprintf("Item should be nil, but got %v", item))
	}
}

func assertPllsLength(t *testing.T, plls *ProcessLists, expected int) {
	length := len(plls.Lists)
	if length != expected {
		t.Error(fmt.Sprintf("Length should be %v but got %v", expected, length))
	}
}
