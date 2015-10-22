// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
)

func Follower(state interfaces.IState) {
	
	for {
		msg := <- state.FollowerInMsgQueue() 
		fmt.Printf("%20s %s\n","Follower:", msg.String())
		msg.Leader(state)
	}
	
}