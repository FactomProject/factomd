// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"encoding/json" // DEBUG
	"fmt"           // DEBUG
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

func HandleNetworkInfo(
	state interfaces.IState,
	params interface{},
) (
	interface{},
	*primitives.JSONError,
) {
	type ret struct {
		NetworkNumber int
		NetworkName   string
		NetworkID     uint32
	}
	r := new(ret)
	r.NetworkNumber = state.GetNetworkNumber()
	r.NetworkName = state.GetNetworkName()
	r.NetworkID = state.GetNetworkID()
	return r, nil
}

func HandlePredictiveFER(
	state interfaces.IState,
	params interface{},
) (
	interface{},
	*primitives.JSONError,
) {
	type ret struct {
		PredictiveFER uint64
	}
	r := new(ret)
	r.PredictiveFER = state.GetPredictiveFER()
	return r, nil
}