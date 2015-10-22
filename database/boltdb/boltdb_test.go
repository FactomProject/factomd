// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package boltdb_test

import (
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/database/boltdb"
	"os"
	"testing"
)

type TestData struct {
	str string
}

func (t *TestData) MarshalBinary() ([]byte, error) {
	return []byte(t.str), nil
}
func (t *TestData) MarshalledSize() uint64 {
	return uint64(len(t.str))
}

func (t *TestData) UnmarshalBinaryData(data []byte) ([]byte, error) {
	t.str = string(data)
	return nil, nil
}

func (t *TestData) UnmarshalBinary(data []byte) (err error) {
	_, err = t.UnmarshalBinaryData(data)
	return
}

var _ interfaces.BinaryMarshallable = (*TestData)(nil)

var dbFilename string = "boltTest.db"

func TestPutGetDelete(t *testing.T) {
	m := new(BoltDB)
	m.Init(nil, dbFilename)
	defer CleanupTest(t, m)

	key := []byte("key")
	bucket := []byte("bucket")

	test := new(TestData)
	test.str = "testtest"

	err := m.Put(bucket, key, test)
	if err != nil {
		t.Errorf("%v", err)
	}

	resp, err := m.Get(bucket, key, new(TestData))
	if err != nil {
		t.Errorf("%v", err)
	}

	if resp == nil {
		t.Errorf("resp is nil")
	}

	if resp.(*TestData).str != test.str {
		t.Errorf("data mismatch")
	}

	err = m.Delete(bucket, key)
	if err != nil {
		t.Errorf("%v", err)
	}

	resp, err = m.Get(bucket, key, new(TestData))
	if err != nil {
		t.Errorf("%v", err)
	}
	if resp != nil {
		t.Errorf("resp is not nil while it should be")
	}
}

func CleanupTest(t *testing.T, b *BoltDB) {
	err := b.Close()
	if err != nil {
		t.Errorf("%v", err)
	}
	err = os.Remove(dbFilename)
	if err != nil {
		t.Errorf("%v", err)
	}
}

/*
import (
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"math/rand"
	"testing"
)

var _ = fmt.Printf
var _ = ed25519.Sign
var _ = rand.New
var _ = binary.Read

// This database stores and retrieves interfaces.IBlock instances.  To do that, it
// needs a list of buckets that the using function wants, so it can make sure
// all those buckets exist.  (Avoids checking and building buckets in every
// write).
//
// It also needs a map of a hash to a interfaces.IBlock instance.  To support this,
// every block needs to be able to give the database a Hash for its type.
// This has to match the reverse, where looking up the hash gives the
// database the type for the hash.  This way, the database can marshal
// and unmarshal interfaces.IBlocks for storage in the database.  And since the interfaces.IBlocks
// can provide the hash, we don't need two maps.  Just the Hash to the
// interfaces.IBlock.

func cp(a interfaces.IHash) [constants.ADDRESS_LENGTH]byte {
	r := new([constants.ADDRESS_LENGTH]byte)
	copy(r[:], a.Bytes())
	return *r
}

func Test_bolt_init(t *testing.T) {
	db := new(BoltDB)

	bucketList := make([][]byte, 5, 5)

	bucketList[0] = []byte("one")
	bucketList[1] = []byte("two")
	bucketList[2] = []byte("three")
	bucketList[3] = []byte("four")
	bucketList[4] = []byte("five")

	instances := make(map[[constants.ADDRESS_LENGTH]byte]interfaces.IBlock)
	{
		var a interfaces.IBlock
		a = new(Address)
		instances[cp(a.GetDBHash())] = a
		a = new(primitives.Hash)
		instances[cp(a.GetDBHash())] = a
		a = new(InAddress)
		instances[cp(a.GetDBHash())] = a
		a = new(OutAddress)
		instances[cp(a.GetDBHash())] = a
		a = new(OutECAddress)
		instances[cp(a.GetDBHash())] = a
		a = new(RCD_1)
		instances[cp(a.GetDBHash())] = a
		a = new(RCD_2)
		instances[cp(a.GetDBHash())] = a
		a = new(FactoidSignature)
		instances[cp(a.GetDBHash())] = a
		a = new(Transaction)
		instances[cp(a.GetDBHash())] = a
	}
	db.Init(bucketList, instances)
	a := new(Address)
	a.SetBytes(Sha([]byte("I came, I saw")).Bytes())
	db.Put("one", primitives.Sha([]byte("one")), a)
	r := db.Get("one", primitives.Sha([]byte("one")))

	if a.IsEqual(r) != nil {
		t.Fail()
	}

	db.DeleteKey([]byte("one"), primitives.Sha([]byte("one")).Bytes())
	r = db.Get("one", primitives.Sha([]byte("one")))

	if r != nil {
		t.Fail()
	}
}
*/
