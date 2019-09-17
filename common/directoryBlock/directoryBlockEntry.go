// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

// DBEntry is a struct containing the information for an arbitrary directory block entry. It includes a chain id (a hash) and a key
// Merkle root (another hash)
type DBEntry struct {
	ChainID interfaces.IHash `json:"chainid"`
	KeyMR   interfaces.IHash `json:"keymr"` // Different MR in EBlockHeader
}

var _ interfaces.Printable = (*DBEntry)(nil)
var _ interfaces.BinaryMarshallable = (*DBEntry)(nil)
var _ interfaces.IDBEntry = (*DBEntry)(nil)

// Init initializes the DBentry hashes to zero if they are nil
func (c *DBEntry) Init() {
	if c.ChainID == nil {
		c.ChainID = primitives.NewZeroHash()
	}
	if c.KeyMR == nil {
		c.KeyMR = primitives.NewZeroHash()
	}
}

// IsSameAs returns true iff the input DBEntry is identical to this DBEntry
func (c *DBEntry) IsSameAs(b interfaces.IDBEntry) bool {
	if c == nil || b == nil {
		if c == nil && b == nil {
			return true
		}
		return false
	}

	if c.ChainID.IsSameAs(b.GetChainID()) == false {
		return false
	}
	if c.KeyMR.IsSameAs(b.GetKeyMR()) == false {
		return false
	}
	return true
}

// GetChainID returns the chain id of the directory block entry
func (c *DBEntry) GetChainID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DBEntry.GetChainID() saw an interface that was nil")
		}
	}()

	return c.ChainID
}

// SetChainID sets the chain id to the input hash
func (c *DBEntry) SetChainID(chainID interfaces.IHash) {
	c.ChainID = chainID
}

// GetKeyMR returns the key Merkle root of the directory block entry
func (c *DBEntry) GetKeyMR() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DBEntry.GetKeyMR() saw an interface that was nil")
		}
	}()

	return c.KeyMR
}

// SetKeyMR sets the key Merkle root to the input hash
func (c *DBEntry) SetKeyMR(keyMR interfaces.IHash) {
	c.KeyMR = keyMR
}

// MarshalBinary marshals the directory block entry
func (c *DBEntry) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DBEntry.MarshalBinary err:%v", *pe)
		}
	}(&err)
	c.Init()
	buf := primitives.NewBuffer(nil)

	err = buf.PushBinaryMarshallable(c.ChainID)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(c.KeyMR)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input into the directory block entry
func (c *DBEntry) UnmarshalBinaryData(data []byte) ([]byte, error) {
	c.Init()
	newData := data
	var err error

	newData, err = c.ChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	newData, err = c.KeyMR.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	return newData, nil
}

// UnmarshalBinary unmarshals the input into the directory block entry
func (c *DBEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = c.UnmarshalBinaryData(data)
	return
}

// ShaHash marshals the directory block entry and returns its hash
func (c *DBEntry) ShaHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DBEntry.ShaHash() saw an interface that was nil")
		}
	}()

	byteArray, _ := c.MarshalBinary()
	return primitives.Sha(byteArray)
}

// JSONByte returns the json encoded byte array for the directory block entry
func (c *DBEntry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(c)
}

// JSONString returns the json encoded string for the directory block entry
func (c *DBEntry) JSONString() (string, error) {
	return primitives.EncodeJSONString(c)
}

// String returns the formatted string for the directory block entry
func (e *DBEntry) String() string {
	var out primitives.Buffer
	out.WriteString("chainid: " + e.GetChainID().String() + "\n")
	out.WriteString("      keymr:   " + e.GetKeyMR().String() + "\n")
	return (string)(out.DeepCopyBytes())
}
