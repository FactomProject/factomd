// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (	
	"github.com/FactomProject/factomd/common/interfaces"
)

type FactomPeer struct {
	
	BroadcastOut      chan interfaces.IMsg
	BroadcastIn       chan interfaces.IMsg
	PrivateOut        chan interfaces.IMsg
	PrivateIn         chan interfaces.IMsg
	
}
