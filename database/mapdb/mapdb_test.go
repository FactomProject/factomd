package mapdb_test

import (
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/database/mapdb"
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

func TestPutGetDelete(t *testing.T) {
	m := new(MapDB)

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
