// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

type DBEntry struct {
	ChainID interfaces.IHash `json:"chainid"`
	KeyMR   interfaces.IHash `json:"keymr"` // Different MR in EBlockHeader
}

var _ interfaces.Printable = (*DBEntry)(nil)
var _ interfaces.BinaryMarshallable = (*DBEntry)(nil)
var _ interfaces.IDBEntry = (*DBEntry)(nil)

func (c *DBEntry) Init() {
	if c.ChainID == nil {
		c.ChainID = primitives.NewZeroHash()
	}
	if c.KeyMR == nil {
		c.KeyMR = primitives.NewZeroHash()
	}
}

func (a *DBEntry) IsSameAs(b interfaces.IDBEntry) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}

	if a.ChainID.IsSameAs(b.GetChainID()) == false {
		return false
	}
	if a.KeyMR.IsSameAs(b.GetKeyMR()) == false {
		return false
	}
	return true
}

func (c *DBEntry) GetChainID() interfaces.IHash {
	return c.ChainID
}

func (c *DBEntry) SetChainID(chainID interfaces.IHash) {
	c.ChainID = chainID
}

func (c *DBEntry) GetKeyMR() interfaces.IHash {
	return c.KeyMR
}

func (c *DBEntry) SetKeyMR(keyMR interfaces.IHash) {
	c.KeyMR = keyMR
}

func (e *DBEntry) MarshalBinary() ([]byte, error) {
	e.Init()
	buf := primitives.NewBuffer(nil)

	err := buf.PushBinaryMarshallable(e.ChainID)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(e.KeyMR)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *DBEntry) UnmarshalBinaryData(data []byte) ([]byte, error) {
	e.Init()
	newData := data
	var err error

	newData, err = e.ChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	newData, err = e.KeyMR.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	return newData, nil
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

func (e *DBEntry) String() string {
	var out primitives.Buffer
	out.WriteString("chainid: " + e.GetChainID().String() + "\n")
	out.WriteString("      keymr:   " + e.GetKeyMR().String() + "\n")
	return (string)(out.DeepCopyBytes())
}
