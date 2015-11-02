// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
)

func NetworkProcessor(state interfaces.IState) {

	for {
		msg := <-state.NetworkInMsgQueue()
		log.Printf("%20s %s\n", "Network:", msg.String())
		state.InMsgQueue() <- msg
	}

}
