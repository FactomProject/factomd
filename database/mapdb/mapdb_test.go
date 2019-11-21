package mapdb_test

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	. "github.com/FactomProject/factomd/database/mapdb"
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

func TestPutGetDelete(t *testing.T) {
	m := new(MapDB)

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
		t.Errorf("resp is not nil while it should be")
	}
}

func TestMultiValue(t *testing.T) {
	m := new(MapDB)

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

func TestParallelAccess(t *testing.T) {
	threads := 100
	m := new(MapDB)
	c := make(chan int)
	closed := make(chan int, threads)
	for i := 0; i < threads; i++ {
		go func() {
			for {
				select {
				case <-c:
					closed <- 1
					return
				default:
					str := new(TestData)
					str.Str = fmt.Sprintf("%x", RandomHex(32))
					err := m.Put(RandomHex(5), RandomHex(5), str)
					if err != nil {
						t.Errorf("Got error - %v", err)
					}
					_, err = m.Get(RandomHex(5), RandomHex(5), str)
					if err != nil {
						t.Errorf("Got error - %v", err)
					}
				}
			}
		}()
	}
	time.Sleep(10 * time.Second)
	close(c)
	time.Sleep(1 * time.Second)
	for i := 0; i < threads; i++ {
		<-closed
	}
}

func RandomHex(length int) []byte {
	if length <= 0 {
		return nil
	}
	answer := make([]byte, length)
	_, err := rand.Read(answer)
	if err != nil {
		return nil
	}
	return answer
}

func TestDoesKeyExist(t *testing.T) {
	m := new(MapDB)
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

func TestGetAll(t *testing.T) {
	m := new(MapDB)

	dbo := databaseOverlay.NewOverlay(m, nil)
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
