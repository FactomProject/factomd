// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factomwire

import (
    "github.com/FactomProject/simplecoin"
)

struct IFactomMessage interface {
    simplecoin.IBlock 
    
    Validate(IFactomState) int
    Leader(IFactomState)
    Follower(IFactomState)
}