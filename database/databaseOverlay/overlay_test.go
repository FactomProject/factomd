// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	"bytes"
	"encoding/gob"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/mapdb"
	"testing"
)

func TestInsertFetch(t *testing.T) {
	dbo := createOverlay()
	defer dbo.Close()
	b := NewDBTestObject()
	b.Data = []byte{0x00, 0x01, 0x02, 0x03}

	err := dbo.Insert(TestBucket, b)
	if err != nil {
		t.Error(err)
	}

	index := b.DatabasePrimaryIndex()

	b2 := NewDBTestObject()
	resp, err := dbo.FetchBlock(TestBucket, index, b2)
	if err != nil {
		t.Error(err)
	}

	if resp == nil {
		t.Error("Response is nil while it shouldn't be.")
	}

	bResp := resp.(*DBTestObject)

	bytes1 := b.Data
	bytes2 := bResp.Data

	if AreBytesEqual(bytes1, bytes2) == false {
		t.Errorf("Bytes are not equal - %x vs %x", bytes1, bytes2)
	}

}

func createOverlay() *Overlay {
	return NewOverlay(new(mapdb.MapDB))
}

var TestBucket []byte = []byte{0x01}

func AreBytesEqual(b1, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i := range b1 {
		if b1[i] != b2[i] {
			return false
		}
	}
	return true
}

type bareDBTestObject struct {
	Data           []byte
	DatabaseHeight uint32
	PrimaryIndex   interfaces.IHash
	SecondaryIndex interfaces.IHash
	ChainID        interfaces.IHash
}

func NewBareDBTestObject() *bareDBTestObject {
	d := new(bareDBTestObject)
	d.Data = []byte{}
	d.DatabaseHeight = 0
	d.PrimaryIndex = new(primitives.Hash)
	d.SecondaryIndex = new(primitives.Hash)
	d.ChainID = new(primitives.Hash)
	return d
}

type DBTestObject struct {
	Data           []byte
	DatabaseHeight uint32
	PrimaryIndex   interfaces.IHash
	SecondaryIndex interfaces.IHash
	ChainID        interfaces.IHash
}

func NewDBTestObject() *DBTestObject {
	d := new(DBTestObject)
	d.Data = []byte{}
	d.DatabaseHeight = 0
	d.PrimaryIndex = new(primitives.Hash)
	d.SecondaryIndex = new(primitives.Hash)
	d.ChainID = new(primitives.Hash)
	return d
}

var _ interfaces.DatabaseBatchable = (*DBTestObject)(nil)

func (d *DBTestObject) GetDatabaseHeight() uint32 {
	return d.DatabaseHeight
}

func (d *DBTestObject) DatabasePrimaryIndex() interfaces.IHash {
	return d.PrimaryIndex
}

func (d *DBTestObject) DatabaseSecondaryIndex() interfaces.IHash {
	return d.SecondaryIndex
}

func (d *DBTestObject) GetChainID() []byte {
	return d.ChainID.Bytes()
}

func (d *DBTestObject) New() interfaces.BinaryMarshallableAndCopyable {
	return NewDBTestObject()
}

func (d *DBTestObject) UnmarshalBinaryData(data []byte) ([]byte, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(data))

	tmp := NewBareDBTestObject()

	err := dec.Decode(tmp)
	if err != nil {
		return nil, err
	}

	d.Data = tmp.Data
	d.DatabaseHeight = tmp.DatabaseHeight
	d.PrimaryIndex = tmp.PrimaryIndex
	d.SecondaryIndex = tmp.SecondaryIndex
	d.ChainID = tmp.ChainID

	return nil, nil
}

func (d *DBTestObject) UnmarshalBinary(data []byte) error {
	_, err := d.UnmarshalBinaryData(data)
	return err
}

func (d *DBTestObject) MarshalBinary() ([]byte, error) {
	var data bytes.Buffer

	enc := gob.NewEncoder(&data)

	tmp := new(bareDBTestObject)

	tmp.Data = d.Data
	tmp.DatabaseHeight = d.DatabaseHeight
	tmp.PrimaryIndex = d.PrimaryIndex
	tmp.SecondaryIndex = d.SecondaryIndex
	tmp.ChainID = d.ChainID

	err := enc.Encode(tmp)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}
