// +build all 

package securedb_test

import (
	"bytes"
	//"crypto/sha256"
	"testing"

	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/securedb"
	"github.com/FactomProject/factomd/testHelper"
)

func TestEncryptedMarshaler(t *testing.T) {
	var o interfaces.BinaryMarshallable
	var err error
	var ems []*EncryptedMarshaler
	var hashes []interfaces.BinaryMarshallable
	s := testHelper.CreateAndPopulateTestStateAndStartValidator()

	key := primitives.RandomHash().Bytes()
	// Test with IHash
	for i := 0; i < 10; i++ {
		o = primitives.RandomHash()
		m := NewEncryptedMarshaler(key, o)
		d := NewEncryptedMarshaler(key, new(primitives.Hash))
		testEM(m, d, o, t)
		ems = append(ems, m)
		hashes = append(hashes, o)
	}

	allData := []byte{}
	for _, e := range ems {
		data, err := e.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		allData = append(allData, data...)
	}

	for _, h := range hashes {
		e2 := NewEncryptedMarshaler(key, new(primitives.Hash))
		allData, err = e2.UnmarshalBinaryData(allData)
		if err != nil {
			t.Error(err)
		}

		d1, err := e2.Original.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		d2, err := h.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		if bytes.Compare(d1, d2) != 0 {
			t.Error("Byte stream unmarshal failed")
		}
	}

	// Test addresses
	for i := 0; i < 10; i++ {
		k := primitives.RandomHash().Bytes()
		o = factoid.RandomAddress()
		m := NewEncryptedMarshaler(k, o)
		d := NewEncryptedMarshaler(k, new(factoid.Address))
		testEM(m, d, o, t)
	}

	// Test a message
	for i := 0; i < 10; i++ {
		k := primitives.RandomHash().Bytes()
		o = messages.NewAddServerMsg(s, 0)
		m := NewEncryptedMarshaler(k, o)
		d := NewEncryptedMarshaler(k, new(messages.AddServerMsg))
		testEM(m, d, o, t)
	}

	// Test a timestamp
	for i := 0; i < 10; i++ {
		k := primitives.RandomHash().Bytes()
		o = primitives.NewTimestampNow()
		m := NewEncryptedMarshaler(k, o)
		d := NewEncryptedMarshaler(k, new(primitives.Timestamp))
		testEM(m, d, o, t)
	}

}

// Send marshaler with oringal and key, send d with blank original and key, send the original
func testEM(m *EncryptedMarshaler, d *EncryptedMarshaler, o interfaces.BinaryMarshallable, t *testing.T) {
	cipherData, err := m.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	plainData, err := d.UnmarshalBinaryData(cipherData)
	if err != nil {
		t.Error(err)
	}

	if len(plainData) > 0 {
		t.Error("Should have no bytes remaining")
	}

	oData, err := o.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	nData, err := d.Original.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(oData, nData) != 0 {
		t.Error("Did not properly marshal/unmarshal")
	}

	if bytes.Compare(oData, cipherData) == 0 {
		t.Error("Data was not encrypted")
	}
}
