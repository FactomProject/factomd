// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/boltdb"
	"math/rand"
	"testing"
)

var _ = hex.EncodeToString
var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Write
var _ = Prtln

func GetDatabase() IDatabase {
	var bucketList [][]byte

	bucketList = make([][]byte, 5, 5)

	bucketList[0] = []byte("factoidAddress_balances")
	bucketList[0] = []byte("factoidOrphans_balances")
	bucketList[0] = []byte("factomAddress_balances")

	db := new(BoltDB)

	db.Init(bucketList, "/tmp/fs_test.db")

	return db
}

func Test_updating_balances_FactoidState(test *testing.T) {
	fs := new(FactoidState)
	fs.database = GetDatabase()

}
