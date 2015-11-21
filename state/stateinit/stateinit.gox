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
	"github.com/FactomProject/factomd/common/factoid/wallet"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
	"github.com/FactomProject/factomd/state"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

var _ = fmt.Printf

func NewFactoidState(filename string) interfaces.IFactoidState {
	fs := new(state.FactoidState)
	wall := new(wallet.SCWallet)
	wall.Init()

	fs.SetWallet(wall)

	// Use Bolt DB
	fs.SetDB(databaseOverlay.NewOverlay(GetDatabase(filename)))

	return fs
}

func GetDatabase(filename string) interfaces.IDatabase {
	var bucketList [][]byte

	bucketList = make([][]byte, 0, 5)

	bucketList = append(bucketList, []byte(constants.DB_FACTOID_BLOCKS))
	bucketList = append(bucketList, []byte(constants.DB_BAD_TRANS))
	bucketList = append(bucketList, []byte(constants.DB_F_BALANCES))
	bucketList = append(bucketList, []byte(constants.DB_EC_BALANCES))

	bucketList = append(bucketList, []byte(constants.DB_BUILD_TRANS))
	bucketList = append(bucketList, []byte(constants.DB_TRANSACTIONS))
	bucketList = append(bucketList, []byte(constants.W_RCD_ADDRESS_HASH))
	bucketList = append(bucketList, []byte(constants.W_ADDRESS_PUB_KEY))
	bucketList = append(bucketList, []byte(constants.W_NAME))
	bucketList = append(bucketList, []byte(constants.W_SEEDS))
	bucketList = append(bucketList, []byte(constants.W_SEED_HEADS))

	db := hybridDB.NewBoltMapHybridDB(bucketList, filename)
	return db
}
