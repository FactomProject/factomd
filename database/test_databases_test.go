// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package database_test

import (
	"fmt"
	"os"
	"testing"

	"reflect"

	"time"

	"sync"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/database/boltdb"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/leveldb"
	"github.com/FactomProject/factomd/database/mapdb"
	"github.com/FactomProject/factomd/database/securedb"
	"github.com/FactomProject/factomd/testHelper"
)

type TestData struct {
	Str string
}

func (t *TestData) New() interfaces.BinaryMarshallableAndCopyable {
	return new(TestData)
}

func (t *TestData) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "TestData.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return []byte(t.Str), nil
}

func (t *TestData) UnmarshalBinaryData(data []byte) ([]byte, error) {
	t.Str = string(data)
	return nil, nil
}

func (t *TestData) UnmarshalBinary(data []byte) (err error) {
	_, err = t.UnmarshalBinaryData(data)
	return
}

var _ interfaces.BinaryMarshallable = (*TestData)(nil)

var dbFilename = "testdb"

func TestAllDatabase(t *testing.T) {
	// Level
	m, err := leveldb.NewLevelDB(dbFilename, true)
	if err != nil {
		t.Error(err)
	}
	testNilRetreive(t, m)
	CleanupTest(t, m)
}

func TestAllDatabases(t *testing.T) {
	totalTests := 5

	// Secure Bolt
	for i := 0; i < totalTests; i++ {
		m, err := securedb.NewEncryptedDB(dbFilename, "Bolt", random.RandomString())
		if err != nil {
			t.Error(err)
		}
		testDB(t, m, i)
		CleanupTest(t, m)
	}

	// Secure LDB
	for i := 0; i < totalTests; i++ {
		m, err := securedb.NewEncryptedDB(dbFilename, "LDB", random.RandomString())
		if err != nil {
			t.Error(err)
		}
		testDB(t, m, i)
		CleanupTest(t, m)
	}

	// Secure Map
	for i := 0; i < totalTests; i++ {
		m, err := securedb.NewEncryptedDB(dbFilename, "Map", random.RandomString())
		if err != nil {
			t.Error(err)
		}
		testDB(t, m, i)
		CleanupTest(t, m)
	}

	// Bolt
	for i := 0; i < totalTests; i++ {
		m := boltdb.NewBoltDB(nil, dbFilename)
		testDB(t, m, i)
		CleanupTest(t, m)
	}

	// Level
	for i := 0; i < totalTests; i++ {
		m, err := leveldb.NewLevelDB(dbFilename, true)
		if err != nil {
			t.Error(err)
		}
		testDB(t, m, i)
		CleanupTest(t, m)
	}

	// Map
	for i := 0; i < totalTests; i++ {
		m := new(mapdb.MapDB)
		testDB(t, m, i)
		CleanupTest(t, m)
	}
}

func testDB(t *testing.T, m interfaces.IDatabase, i int) {
	switch i {
	case 0:
		testPutGetDelete(t, m)
	case 1:
		testMultiValue(t, m)
	case 2:
		testDoesKeyExist(t, m)
	case 3:
		testGetAll(t, m)
	case 4:
		testNilRetreive(t, m)
	}
}

func testPutGetDelete(t *testing.T, m interfaces.IDatabase) {
	defer CleanupTest(t, m)

	key := []byte("key")
	bucket := []byte("bucket")

	test := new(TestData)
	test.Str = "testtest"

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

	if resp.(*TestData).Str != test.Str {
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
		fmt.Println(resp)
		t.Errorf("resp is not nil while it should be")
	}
}

func testMultiValue(t *testing.T, m interfaces.IDatabase) {
	defer CleanupTest(t, m)

	bucket := []byte("bucket")
	batch := []interfaces.Record{}
	for i := 0; i < 10; i++ {
		r := interfaces.Record{}
		r.Key = []byte(fmt.Sprintf("%v", i))
		r.Bucket = bucket
		td := new(TestData)
		td.Str = fmt.Sprintf("Data %v", i)
		r.Data = td
		batch = append(batch, r)
	}

	err := m.PutInBatch(batch)
	if err != nil {
		t.Error(err)
	}

	keys, err := m.ListAllKeys(bucket)
	if err != nil {
		t.Error(err)
	}
	if len(keys) != 10 {
		t.Error("Invalid length of keys")
	}
	for i := range keys {
		if string(keys[i]) != fmt.Sprintf("%v", i) {
			t.Error("Wrong key returned")
		}
	}

	all, _, err := m.GetAll(bucket, new(TestData))
	if err != nil {
		t.Error(err)
	}
	if len(all) != 10 {
		t.Error("Invalid length of keys")
	}
	for i := range all {
		v := all[i].(*TestData)
		if v.Str != fmt.Sprintf("Data %v", i) {
			t.Error("Wrong data returned")
		}
	}
	err = m.Clear(bucket)
	if err != nil {
		t.Error(err)
	}

	keys, err = m.ListAllKeys(bucket)
	if err != nil {
		t.Error(err)
	}
	if len(keys) != 0 {
		t.Error("Keys not cleared from database properly")
	}
}

func CleanupTest(t *testing.T, m interfaces.IDatabase) {
	m.Close()
	os.RemoveAll(dbFilename)
	os.Remove(dbFilename)
}

func testDoesKeyExist(t *testing.T, m interfaces.IDatabase) {
	defer CleanupTest(t, m)

	for i := 0; i < 1000; i++ {
		key := random.RandNonEmptyByteSlice()
		bucket := random.RandNonEmptyByteSlice()

		test := new(TestData)
		test.Str = "testtest"

		err := m.Put(bucket, key, test)
		if err != nil {
			t.Errorf("%v", err)
		}

		exists, err := m.DoesKeyExist(bucket, key)
		if err != nil {
			t.Errorf("%v", err)
		}

		if exists == false {
			t.Errorf("Key does not exist")
		}

		key = random.RandNonEmptyByteSlice()
		bucket = random.RandNonEmptyByteSlice()

		exists, err = m.DoesKeyExist(bucket, key)
		if err != nil {
			t.Errorf("%v", err)
		}

		if exists == true {
			t.Errorf("Key does exist while it shouldn't")
		}
	}
}

func testGetAll(t *testing.T, m interfaces.IDatabase) {
	defer CleanupTest(t, m)

	dbo := databaseOverlay.NewOverlay(m)
	testHelper.PopulateTestDatabaseOverlay(dbo)

	_, keys, err := dbo.GetAll(databaseOverlay.INCLUDED_IN, primitives.NewZeroHash())
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(keys) != 150 {
		t.Errorf("Invalid amount of keys returned - expected 150, got %v", len(keys))
	}
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			if primitives.AreBytesEqual(keys[i], keys[j]) {
				t.Errorf("Key %v is equal to key %v - %x", i, j, keys[i])
			}
		}
		if len(keys[i]) != 32 {
			t.Errorf("Wrong key length at index %v - %v", i, len(keys[i]))
		}
	}
}

func testNilRetreive(t *testing.T, m interfaces.IDatabase) {
	o := databaseOverlay.NewOverlay(m)
	//totalEntries := 10000

	g := sync.WaitGroup{}

	writer := func(s, l int) { // Writes
		g.Add(1)
		for k, _ := range filledMap(s, l) {
			err := o.InsertEntry(entryBlock.DeterministicEntry(k))
			if err != nil {
				t.Errorf("%s", err.Error())
			}
			time.Sleep(5 * time.Millisecond)
		}
		g.Done()
	}

	reader := func(s, l int) { // Reads
		g.Add(1)
		for k, _ := range filledMap(s, l) {
			f_e, err := o.FetchEntry(entryBlock.DeterministicEntry(k).GetHash())
			if err != nil {
				t.Errorf("%s", err.Error())
			}
			if f_e != nil && reflect.ValueOf(f_e).IsNil() {
				t.Errorf("Expected a nil, got %v", f_e)
			}
			time.Sleep(5 * time.Millisecond)
		}
		g.Done()
	}

	for i := 0; i < 3; i++ {
		go writer(0, 1000)
		go writer(0, 10000)
		go writer(0, 10000)

		// Add contention on 0-1k
		go reader(0, 1000)
		go reader(0, 1000)
		go reader(0, 1000)
		go reader(0, 10000)
		go reader(0, 10000)
	}
	// Kinda kulgy, but each goroutine adds itself to wait group.
	// Give them a chance to add themselves
	time.Sleep(10 * time.Millisecond)

	g.Wait()

	e := entryBlock.RandomEntry()
	f_e, err := o.FetchEntry(e.GetHash())
	if f_e != nil && reflect.ValueOf(f_e).IsNil() {
		t.Errorf("Expected a nil, got %v", f_e)
	}

	err = o.InsertEntry(e)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	f_e, err = o.FetchEntry(e.GetHash())
	if f_e != nil && reflect.ValueOf(f_e).IsNil() {
		t.Errorf("Expected a nil, got %v", f_e)
	}

}

func filledMap(start, length int) map[int]struct{} {
	avail := make(map[int]struct{})
	for i := start; i < start+length; i++ {
		avail[i] = struct{}{}
	}
	return avail
}
