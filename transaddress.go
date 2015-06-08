// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Transaction Address for a Simplecoin transaction.   contains an amount
// and the address.  Our inputs spec how much is going into a transaction
// and our outputs spec how much is going out of a transaction.  This
// avoids having to have extra outputs to deal with change.
//

package simplecoin

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ITransAddress interface {
	IBlock
	GetAmount() uint64
	SetAmount(uint64)
	GetAddress() IAddress
	SetAddress(IAddress)
	MarshalText2(*bytes.Buffer)
}

type TransAddress struct {
	ITransAddress
	amount  uint64
	address IAddress
}

var _ ITransAddress = (*TransAddress)(nil)

func (t *TransAddress) IsEqual(addr IBlock) []IBlock {
	a, ok := addr.(ITransAddress)
	if !ok || // Not the right kind of IBlock
		a.GetAmount() != t.GetAmount() {
            r := make([]IBlock,0,5)
            return append(r,t)
    }// Amount is different
    r := a.GetAddress().IsEqual(t.GetAddress()) // Address is different
    if r != nil {
        return append(r,t)
    }
    return nil
}



func (t *TransAddress) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	if len(data) < 36 {
		return nil, fmt.Errorf("Data source too short to UnmarshalBinary() an address: %d", len(data))
	}

	t.amount, data = binary.BigEndian.Uint64(data[0:8]), data[8:]
	t.address = new(Address)

	data, err = t.address.UnmarshalBinaryData(data)

	return data, err
}

// MarshalBinary.  'nuff said
func (a TransAddress) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	binary.Write(&out, binary.BigEndian, uint64(a.amount))
	data, err := a.address.MarshalBinary()
	out.Write(data)

	return out.Bytes(), err
}

// Accessor. Default to a zero length string.  This is a debug
// thing for looking out what we have built. Used by
// MarshalText
func (ta TransAddress) GetName() string {
	return ""
}

// Accessor.  Get the amount with this address.
func (ta TransAddress) GetAmount() uint64 {
	return ta.amount
}

// Accessor.  Get the amount with this address.
func (ta *TransAddress) SetAmount(amount uint64) {
	ta.amount = amount
}

// Accessor.  Get the raw address.  Could be an actual address,
// or a hash of an authorization block.  See authorization.go
func (ta TransAddress) GetAddress() IAddress {
	return ta.address
}

// Accessor.  Get the raw address.  Could be an actual address,
// or a hash of an authorization block.  See authorization.go
func (ta *TransAddress) SetAddress(address IAddress) {
	ta.address = address
}

// Make this into somewhat readable text.
func (ta TransAddress) MarshalText2(out *bytes.Buffer) {

	WriteNumber64(out, ta.amount)
	out.WriteString(" ")

	text, _ := ta.address.MarshalText()
	out.Write(text)

	return
}
