// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// This package exists to isolate the fact that we need to reference
// almost all the objects within the Transaction package inorder to
// properly configure the database.   This code can be called by
// system level code (which is isolated from the other Transaction
// packages) without causing import loops.
package stateinit

import (
	"fmt"
	"github.com/FactomProject/factomd/common/factoid/state"
	"github.com/FactomProject/factomd/common/factoid/wallet"
	"github.com/FactomProject/factomd/database/hybridDB"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

var _ = fmt.Printf

func NewFactoidState(filename string) IFactoidState {
	fs := new(state.FactoidState)
	wall := new(wallet.SCWallet)
	wall.Init()

	fs.SetWallet(wall)

	// Use Bolt DB
	fs.SetDB(GetDatabase(filename))

	return fs
}

func GetDatabase(filename string) interfaces.IDatabase {
	var bucketList [][]byte

	bucketList = make([][]byte, 0, 5)

	bucketList = append(bucketList, []byte(DB_FACTOID_BLOCKS))
	bucketList = append(bucketList, []byte(DB_BAD_TRANS))
	bucketList = append(bucketList, []byte(DB_F_BALANCES))
	bucketList = append(bucketList, []byte(DB_EC_BALANCES))

	bucketList = append(bucketList, []byte(DB_BUILD_TRANS))
	bucketList = append(bucketList, []byte(DB_TRANSACTIONS))
	bucketList = append(bucketList, []byte(W_RCD_ADDRESS_HASH))
	bucketList = append(bucketList, []byte(W_ADDRESS_PUB_KEY))
	bucketList = append(bucketList, []byte(W_NAME))
	bucketList = append(bucketList, []byte(W_SEEDS))
	bucketList = append(bucketList, []byte(W_SEED_HEADS))

	db := hybridDB.NewBoltMapHybridDB(bucketList, filename)
	return db
}
