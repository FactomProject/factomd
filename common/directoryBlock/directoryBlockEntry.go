// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryblock

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type DBEntry struct {
	ChainID interfaces.IHash
	KeyMR   interfaces.IHash // Different MR in EBlockHeader
}

var _ interfaces.Printable = (*DBEntry)(nil)
var _ interfaces.BinaryMarshallable = (*DBEntry)(nil)
var _ interfaces.IDBEntry = (*DBEntry)(nil)

func (c *DBEntry) GetChainID() interfaces.IHash {
	return c.ChainID
}
func (c *DBEntry) GetKeyMR() (interfaces.IHash, error) {
	return c.KeyMR, nil
}

func NewDBEntry(entry interfaces.IEntry) (*DBEntry, error) {
	e := new(DBEntry)

	e.ChainID = entry.GetChainID()
	var err error
	e.KeyMR, err = entry.GetKeyMR()
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (e *DBEntry) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, err = e.ChainID.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	data, err = e.KeyMR.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	return buf.Bytes(), nil
}

func (e *DBEntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	e.ChainID = new(primitives.Hash)
	newData, err = e.ChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	e.KeyMR = new(primitives.Hash)
	newData, err = e.KeyMR.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	return
}

func (e *DBEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *DBEntry) ShaHash() interfaces.IHash {
	byteArray, _ := e.MarshalBinary()
	return primitives.Sha(byteArray)
}

func (e *DBEntry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DBEntry) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *DBEntry) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *DBEntry) String() string {
	str, _ := e.JSONString()
	return str
}
