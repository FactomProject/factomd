// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Transaction Address for a factoid transaction.   contains an amount
// and the address.  Our inputs spec how much is going into a transaction
// and our outputs spec how much is going out of a transaction.  This
// avoids having to have extra outputs to deal with change.
//

package factoid

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

type TransAddress struct {
	Amount  uint64              `json:"amount"`
	Address interfaces.IAddress `json:"address"`
	// Not marshalled
	UserAddress string `json:"useraddress"`
}

var _ interfaces.ITransAddress = (*TransAddress)(nil)

func RandomTransAddress() interfaces.ITransAddress {
	ta := new(TransAddress)
	ta.Address = RandomAddress()
	ta.Amount = random.RandUInt64()
	ta.UserAddress = random.RandomString()
	return ta
}

func (t *TransAddress) SetUserAddress(v string) {
	t.UserAddress = v
}

func (t *TransAddress) GetUserAddress() string {
	return t.UserAddress
}

// Not useful on TransAddress objects
func (t *TransAddress) GetHash() interfaces.IHash {
	return nil
}

func (t *TransAddress) UnmarshalBinary(data []byte) error {
	_, err := t.UnmarshalBinaryData(data)
	return err
}

func (t *TransAddress) CustomMarshalText() ([]byte, error) {
	return nil, nil
}

func (e *TransAddress) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *TransAddress) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *TransAddress) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (t *TransAddress) String() string {
	txt, _ := t.CustomMarshalText()
	return (string(txt))
}

func (t *TransAddress) IsEqual(addr interfaces.IBlock) []interfaces.IBlock {
	a, ok := addr.(interfaces.ITransAddress)
	if !ok || // Not the right kind of interfaces.IBlock
		a.GetAmount() != t.GetAmount() {
		r := make([]interfaces.IBlock, 0, 5)
		return append(r, t)
	} // Amount is different
	r := a.GetAddress().IsEqual(t.GetAddress()) // Address is different
	if r != nil {
		return append(r, t)
	}
	return nil
}

func (t *TransAddress) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	if len(data) < 36 {
		return nil, fmt.Errorf("Data source too short to UnmarshalBinary() an address: %d", len(data))
	}

	t.Amount, data = primitives.DecodeVarInt(data)
	t.Address = new(Address)

	data, err = t.Address.UnmarshalBinaryData(data)

	return data, err
}

// MarshalBinary.  'nuff said
func (a TransAddress) MarshalBinary() ([]byte, error) {
	var out primitives.Buffer

	err := primitives.EncodeVarInt(&out, a.Amount)
	if err != nil {
		return nil, err
	}
	data, err := a.Address.MarshalBinary()
	out.Write(data)

	return out.DeepCopyBytes(), err
}

// Accessor. Default to a zero length string.  This is a debug
// thing for looking out what we have built. Used by
// CustomMarshalText
func (ta TransAddress) GetName() string {
	return ""
}

// Accessor.  Get the amount with this address.
func (ta TransAddress) GetAmount() uint64 {
	return ta.Amount
}

// Accessor.  Get the amount with this address.
func (ta *TransAddress) SetAmount(amount uint64) {
	ta.Amount = amount
}

// Accessor.  Get the raw address.  Could be an actual address,
// or a hash of an authorization block.  See authorization.go
func (ta TransAddress) GetAddress() interfaces.IAddress {
	return ta.Address
}

// Accessor.  Get the raw address.  Could be an actual address,
// or a hash of an authorization block.  See authorization.go
func (ta *TransAddress) SetAddress(address interfaces.IAddress) {
	ta.Address = address
}

// Make this into somewhat readable text.
func (ta TransAddress) CustomMarshalTextAll(fct bool, label string) ([]byte, error) {
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("   %8s:", label))
	v := primitives.ConvertDecimalToPaddedString(ta.Amount)
	fill := 8 - len(v) + strings.Index(v, ".") + 1
	fstr := fmt.Sprintf("%%%vs%%%vs ", 18-fill, fill)
	out.WriteString(fmt.Sprintf(fstr, v, ""))
	if fct {
		out.WriteString(primitives.ConvertFctAddressToUserStr(ta.Address))
	} else {
		out.WriteString(primitives.ConvertECAddressToUserStr(ta.Address))
	}
	str := fmt.Sprintf("\n                  %016x %038s\n\n", ta.Amount, string(hex.EncodeToString(ta.GetAddress().Bytes())))
	out.WriteString(str)
	return out.DeepCopyBytes(), nil
}

func (ta TransAddress) CustomMarshalText2(label string) ([]byte, error) {
	return ta.CustomMarshalTextAll(true, label)
}

func (ta TransAddress) CustomMarshalTextEC2(label string) ([]byte, error) {
	return ta.CustomMarshalTextAll(false, label)
}
