// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// IncreaseBalance is an entry credit block entry type which increases the entry credit balance at an address
type IncreaseBalance struct {
	ECPubKey *primitives.ByteSlice32 `json:"ecpubkey"` // EC public key that will have balanced increased
	TXID     interfaces.IHash        `json:"txid"`     // The transaction id associated with this balance increase
	Index    uint64                  `json:"index"`    // The index into the transaction's purchase field for this balance increase
	NumEC    uint64                  `json:"numec"`    // The number of entry credits added to the address (based on current exchange rate)
}

var _ interfaces.Printable = (*IncreaseBalance)(nil)

var _ interfaces.BinaryMarshallable = (*IncreaseBalance)(nil)
var _ interfaces.ShortInterpretable = (*IncreaseBalance)(nil)
var _ interfaces.IECBlockEntry = (*IncreaseBalance)(nil)

// Init initializes all nil objects
func (e *IncreaseBalance) Init() {
	if e.ECPubKey == nil {
		e.ECPubKey = new(primitives.ByteSlice32)
	}
	if e.TXID == nil {
		e.TXID = primitives.NewZeroHash()
	}
}

// IsSameAs returns true iff the input object is identical to this object
func (e *IncreaseBalance) IsSameAs(b interfaces.IECBlockEntry) bool {
	if e == nil || b == nil {
		if e == nil && b == nil {
			return true
		}
		return false
	}
	if e.ECID() != b.ECID() {
		return false
	}

	bb, ok := b.(*IncreaseBalance)
	if ok == false {
		return false
	}

	if e.ECPubKey.IsSameAs(bb.ECPubKey) == false {
		return false
	}
	if e.TXID.IsSameAs(bb.TXID) == false {
		return false
	}
	if e.Index != bb.Index {
		return false
	}
	if e.NumEC != bb.NumEC {
		return false
	}

	return true
}

// String returns this object as a string
func (e *IncreaseBalance) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf(" %-20s\n", "IncreaseBalance"))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "ECPubKey", e.ECPubKey[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "TXID", e.TXID.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "Index", e.Index))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "NumEC", e.NumEC))

	return (string)(out.DeepCopyBytes())
}

// NewIncreaseBalance creates a newly initialized object
func NewIncreaseBalance() *IncreaseBalance {
	r := new(IncreaseBalance)
	r.Init()
	return r
}

// GetEntryHash always returns nil
func (e *IncreaseBalance) GetEntryHash() (rval interfaces.IHash) {
	// reenable if this function is implemented
	// defer func() { rval = primitives.CheckNil(rval, "IncreaseBalance.GetEntryHash") }()
	
	return nil
}

// Hash marshals the object and computes its sha
func (e *IncreaseBalance) Hash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "IncreaseBalance.Hash") }()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

// GetHash marshals the object and computes its sha
func (e *IncreaseBalance) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "IncreaseBalance.GetHash") }()

	return e.Hash()
}

// GetSigHash always returns nil
func (e *IncreaseBalance) GetSigHash() (rval interfaces.IHash) {
	// reenable if this function is implemented
	// defer func() { rval = primitives.CheckNil(rval, "IncreaseBalance.GetSigHash") }()

	return nil
}

// ECID returns the hard coded entry credit id ECIDBalanceIncrease
func (e *IncreaseBalance) ECID() byte {
	return constants.ECIDBalanceIncrease
}

// IsInterpretable always returns false
func (e *IncreaseBalance) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *IncreaseBalance) Interpret() string {
	return ""
}

// MarshalBinary marshals this object
func (e *IncreaseBalance) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "IncreaseBalance.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	buf := primitives.NewBuffer(nil)

	err = buf.PushBinaryMarshallable(e.ECPubKey)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.TXID)
	if err != nil {
		return nil, err
	}
	err = buf.PushVarInt(e.Index)
	if err != nil {
		return nil, err
	}
	err = buf.PushVarInt(e.NumEC)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (e *IncreaseBalance) UnmarshalBinaryData(data []byte) ([]byte, error) {
	e.Init()
	buf := primitives.NewBuffer(data)

	err := buf.PopBinaryMarshallable(e.ECPubKey)
	if err != nil {
		return nil, err
	}
	err = buf.PopBinaryMarshallable(e.TXID)
	if err != nil {
		return nil, err
	}
	e.Index, err = buf.PopVarInt()
	if err != nil {
		return nil, err
	}
	e.NumEC, err = buf.PopVarInt()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this object
func (e *IncreaseBalance) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *IncreaseBalance) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *IncreaseBalance) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

// GetTimestamp always returns nil
func (e *IncreaseBalance) GetTimestamp() interfaces.Timestamp {
	return nil
}
