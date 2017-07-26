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
)

func TestEncryptedMarshaler(t *testing.T) {
	var o interfaces.BinaryMarshallable

	// Test with IHash
	for i := 0; i < 10; i++ {
		k := primitives.RandomHash().Bytes()
		o = primitives.RandomHash()
		m := NewEncryptedMarshaler(k, o)
		d := NewEncryptedMarshaler(k, new(primitives.Hash))
		testEM(m, d, o, t)
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
		o = messages.NewServerFault(primitives.RandomHash(), primitives.RandomHash(), 0, 0, 0, 0, primitives.NewTimestampNow())
		m := NewEncryptedMarshaler(k, o)
		d := NewEncryptedMarshaler(k, new(messages.ServerFault))
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
