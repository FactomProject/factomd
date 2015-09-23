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
	"github.com/FactomProject/factomd/common/factoid/database"
	"github.com/FactomProject/factomd/common/factoid/state"
	"github.com/FactomProject/factomd/common/factoid/wallet"

	. "github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/factoid"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Printf

func NewFactoidState(filename string) IFactoidState {
	fs := new(state.FactoidState)
	wall := new(wallet.SCWallet)
	wall.Init()

	fs.SetWallet(wall)
	fs.GetWallet().GetDB().Init()

	// Use Bolt DB
	if true {
		fs.SetDB(new(database.MapDB))
		fs.GetDB().Init()
		db := GetDatabase(filename)
		fs.GetDB().SetPersist(db)
		fs.GetDB().SetBacker(db)
		fs.GetWallet().GetDB().SetPersist(db)
		fs.GetWallet().GetDB().SetBacker(db)

		fs.GetDB().DoNotPersist(DB_F_BALANCES)
		fs.GetDB().DoNotPersist(DB_EC_BALANCES)
		fs.GetDB().DoNotPersist(DB_BUILD_TRANS)
		fs.GetDB().DoNotCache(DB_FACTOID_BLOCKS)
		fs.GetDB().DoNotCache(DB_TRANSACTIONS)

	} else {
		fs.SetDB(GetDatabase(filename))
	}

	return fs
}

func GetDatabase(filename string) IFDatabase {

	var bucketList [][]byte
	var instances map[[ADDRESS_LENGTH]byte]IBlock

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

	instances = make(map[[ADDRESS_LENGTH]byte]IBlock)

	var addinstance = func(b IBlock) {
		key := new([32]byte)
		copy(key[:], b.GetDBHash().Bytes())
		instances[*key] = b
	}
	addinstance(new(database.ByteStore))
	addinstance(new(Address))
	addinstance(new(Hash))
	addinstance(new(InAddress))
	addinstance(new(OutAddress))
	addinstance(new(OutECAddress))
	addinstance(new(RCD_1))
	addinstance(new(RCD_2))
	addinstance(new(FactoidSignature))
	addinstance(new(Transaction))
	addinstance(new(block.FBlock))
	addinstance(new(state.FSbalance))
	addinstance(new(wallet.WalletEntry))

	db := new(database.BoltDB)
	db.Init(bucketList, instances, filename)
	return db
}
