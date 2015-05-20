// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Output object for a Simplecoin transaction.   contains an amount
// and the destination address.

package simplecoin

import (
	"bytes"
	"encoding/binary"
)

type ITransAddress interface {
	IBlock
	GetName() string
	GetAmount() uint64
	GetAddress() IAddress
}

type TransAddress struct {
	ITransAddress
	amount  uint64
	address IAddress
}

func (a TransAddress) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	binary.Write(&out, binary.BigEndian, uint64(a.amount))
	data, err := a.address.MarshalBinary()
	out.Write(data)

	return out.Bytes(), err
}

// Default to a zero length string
func (ta TransAddress) GetName() string {
	return ""
}

func (ta TransAddress) GetAmount() uint64 {
	return ta.amount
}

func (ta TransAddress) GetAddress() IAddress {
	return ta.address
}

func (ta TransAddress) MarshalText() ([]byte, error) {
	var out bytes.Buffer

	out.WriteString(ta.GetName())
	out.WriteString("  ")
	WriteNumber64(&out, ta.amount)
	out.WriteString(" ")

	text, _ := ta.address.MarshalText()
	out.Write(text)

	return out.Bytes(), nil
}
