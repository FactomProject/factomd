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
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

// TransAddress contains an address associated with a transaction and a transaction amount
type TransAddress struct {
	Amount  uint64              `json:"amount"`  // The amount in factoshis
	Address interfaces.IAddress `json:"address"` // The address for this transaction
	// Not marshalled
	UserAddress string `json:"useraddress"` // Human readable address for this transaction
}

var _ interfaces.ITransAddress = (*TransAddress)(nil)

// RandomTransAddress returns a new random TransAddress with a random amount
func RandomTransAddress() interfaces.ITransAddress {
	ta := new(TransAddress)
	ta.Address = RandomAddress()
	ta.Amount = random.RandUInt64()
	ta.UserAddress = primitives.ConvertFctAddressToUserStr(ta.Address)
	return ta
}

// SetUserAddress sets the user address to the input
func (ta *TransAddress) SetUserAddress(v string) {
	ta.UserAddress = v
}

// GetUserAddress returns the user address
func (ta *TransAddress) GetUserAddress() string {
	return ta.UserAddress
}

// UnmarshalBinary unmarshals the input data into this object
func (ta *TransAddress) UnmarshalBinary(data []byte) error {
	_, err := ta.UnmarshalBinaryData(data)
	return err
}

// JSONByte returns the json encoded byte array
func (ta *TransAddress) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(ta)
}

// JSONString returns the json encoded string
func (ta *TransAddress) JSONString() (string, error) {
	return primitives.EncodeJSONString(ta)
}

// String returns this object as a string
func (ta *TransAddress) String() string {
	str, _ := ta.JSONString()
	return str
}

// IsSameAs returns true iff the input object is identical to this object
func (ta *TransAddress) IsSameAs(add interfaces.ITransAddress) bool {
	if ta.GetAmount() != add.GetAmount() {
		return false
	}
	if ta.GetAddress().IsSameAs(add.GetAddress()) == false {
		return false
	}
	return true
}

// UnmarshalBinaryData unmrashals the input data into this object
func (ta *TransAddress) UnmarshalBinaryData(data []byte) ([]byte, error) {
	//
	if len(data) < 33 { // leading varint has to be one or more bytes and the address is 32
		return nil, fmt.Errorf("Data source too short in TransAddress.UnmarshalBinaryData() an address: %d", len(data))
	}
	buf := primitives.NewBuffer(data)
	var err error

	ta.Amount, err = buf.PopVarInt()
	if len(data) < 32 { // after the varint there have to be 32 bytes left but there may be some other struct following so longer is fine
		return nil, fmt.Errorf("Data source too short in TransAddress.UnmarshalBinaryData() an address: %d", len(data))
	}
	if err != nil {
		return nil, err
	}

	ta.Address = new(Address)
	err = buf.PopBinaryMarshallable(ta.Address)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// MarshalBinary marshals the object
func (ta TransAddress) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)
	err := buf.PushVarInt(ta.Amount)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(ta.Address)
	if err != nil {
		return nil, err
	}
	return buf.DeepCopyBytes(), nil
}

// GetName defaults to a zero length string.  This is a debug
// thing for looking out what we have built. Used by
// CustomMarshalText
func (ta TransAddress) GetName() string {
	return ""
}

// GetAmount returns the transaction amount in factoshis for this address.
func (ta TransAddress) GetAmount() uint64 {
	return ta.Amount
}

// SetAmount sets the transaction amount in factoshis for this address.
func (ta *TransAddress) SetAmount(amount uint64) {
	ta.Amount = amount
}

// GetAddress gets the raw address.  Could be an actual address,
// or a hash of an authorization block.  See authorization.go
func (ta TransAddress) GetAddress() interfaces.IAddress {
	return ta.Address
}

// SetAddress sets the raw address.  Could be an actual address,
// or a hash of an authorization block.  See authorization.go
func (ta *TransAddress) SetAddress(address interfaces.IAddress) {
	ta.Address = address
}

// CustomMarshalTextAll marshals the object into somewhat readable text. Input 'fct' bool tells the function
// whether to interpret the address as an FCT address or an EC address. Input string 'label' is a user
// defined debug string to help differentiate other TransAddress in string form
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

// CustomMarshalText2 marshals the object as an FCT address with the input label
func (ta TransAddress) CustomMarshalText2(label string) ([]byte, error) {
	return ta.CustomMarshalTextAll(true, label)
}

// CustomMarshalTextEC2 marshals this object as an EC address with the input label
func (ta TransAddress) CustomMarshalTextEC2(label string) ([]byte, error) {
	return ta.CustomMarshalTextAll(false, label)
}

// CustomMarshalTextInput marshals the object as an FCT address with label 'input'
func (ta TransAddress) CustomMarshalTextInput() ([]byte, error) {
	return ta.CustomMarshalText2("input")
}

// StringInput marshals the object to a string as an FCT address with label 'input'
func (ta TransAddress) StringInput() string {
	b, _ := ta.CustomMarshalTextInput()
	return string(b)
}

// CustomMarshalTextOutput marshals the object to a string as an FCT address with label 'output'
func (ta TransAddress) CustomMarshalTextOutput() ([]byte, error) {
	return ta.CustomMarshalText2("output")
}

// StringOutput marshals the object to a string as an FCT address with label 'output'
func (ta TransAddress) StringOutput() string {
	b, _ := ta.CustomMarshalTextOutput()
	return string(b)
}

// CustomMarshalTextECOutput marshals this object as an EC address with label 'ecoutput'
func (ta TransAddress) CustomMarshalTextECOutput() ([]byte, error) {
	return ta.CustomMarshalTextEC2("ecoutput")
}

// StringECOutput marshals the object to a string as an EC address with label 'ecoutput'
func (ta TransAddress) StringECOutput() string {
	b, _ := ta.CustomMarshalTextECOutput()
	return string(b)
}

/******************************
 * Helper functions
 ******************************/

// NewOutECAddress creates a new TransAddress with input EC address and amount
func NewOutECAddress(address interfaces.IAddress, amount uint64) interfaces.ITransAddress {
	ta := new(TransAddress)
	ta.Amount = amount
	ta.Address = address
	ta.UserAddress = primitives.ConvertECAddressToUserStr(address)
	return ta
}

// NewOutAddress creates a new TransAddress with input FCT address and amount
func NewOutAddress(address interfaces.IAddress, amount uint64) interfaces.ITransAddress {
	ta := new(TransAddress)
	ta.Amount = amount
	ta.Address = address
	ta.UserAddress = primitives.ConvertFctAddressToUserStr(address)
	return ta
}

// NewInAddress creates a new TransAddress with input FCT address and amount
func NewInAddress(address interfaces.IAddress, amount uint64) interfaces.ITransAddress {
	ta := new(TransAddress)
	ta.Amount = amount
	ta.Address = address
	//  at this point we know this address is an EC address.
	//  so fill useraddress with a factoid formatted human readable address
	ta.UserAddress = primitives.ConvertFctAddressToUserStr(address)
	return ta
}
