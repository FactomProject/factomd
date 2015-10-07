// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/factoid/database"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
	"math/rand"
	"testing"
)

var _ = hex.EncodeToString
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write
var _ = Prtln

func GetDatabase() IFDatabase {

	var bucketList [][]byte
	var instances map[[ADDRESS_LENGTH]byte]IBlock
	var addinstance = func(b IBlock) {
		key := new([32]byte)
		copy(key[:], b.GetDBHash().Bytes())
		instances[*key] = b
	}

	bucketList = make([][]byte, 5, 5)

	bucketList[0] = []byte("factoidAddress_balances")
	bucketList[0] = []byte("factoidOrphans_balances")
	bucketList[0] = []byte("factomAddress_balances")

	instances = make(map[[ADDRESS_LENGTH]byte]IBlock)

	addinstance(new(Address))
	addinstance(new(Hash))
	addinstance(new(InAddress))
	addinstance(new(OutAddress))
	addinstance(new(OutECAddress))
	addinstance(new(RCD_1))
	addinstance(new(RCD_2))
	addinstance(new(FactoidSignature))
	addinstance(new(Transaction))
	addinstance(new(FSbalance))

	db := new(database.BoltDB)

	db.Init(bucketList, instances, "/tmp/fs_test.db")

	return db
}

func Test_updating_balances_FactoidState(test *testing.T) {
	fs := new(FactoidState)
	fs.database = GetDatabase()

}
