// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// This package doesn't pretend to be Factom.  What it does
// is provide data so I can validate Transaction Processing.

package factomwire

import (
    "github.com/FactomProject/simplecoin"
    "github.com/FactomProject/simplecoin/database"
    "cointainer/list"
)

type IFactomState interface {
    Init()
    GetDatabase() database.ISCDatabase
}

type FactomState struct {
    IFactomState
    db              database.ISCDatabase    
    processList     *list.List
}

func (f FactomState)Init {
    processList = list.New()
    db = new(database.SCDatabase)
    