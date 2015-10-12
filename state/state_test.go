// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"testing"
)

func TestInit(t *testing.T) {
	state := new(State)
	state.Init()
	fmt.Println(state)
}

