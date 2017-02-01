// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	
	"fmt" // DEBUG
	"encoding/json" // DEBUG
)

var _ = fmt.Sprintln("DEBUG")
var _ = json.Marshal // "DEBUG"

func HandleHoldingQueue(
	state interfaces.IState,
	params interface{},
) (
	interface{},
	*primitives.JSONError,
) {
	type ret struct {
		Messages []interfaces.IMsg
	}
	r := new(ret)
	
	for _, v := range state.LoadHoldingMap() {
		r.Messages = append(r.Messages, v)
	}
	return r, nil
}

func HandleNodeStatus(
	state interfaces.IState,
	params interface{},
) (
	interface{},
	*primitives.JSONError,
) {
	type ret struct {
		Status []string
	}
	r := new(ret)
	r.Status = state.GetStatus()
	
	return r, nil
}