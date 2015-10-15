// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

// Entry Credit Addresses are the same as Addresses in factoid
// They just get printed out differently when we output them in
// human readable form.
//
// Entry Credit Addresses are always outputs.

import (
	"bytes"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)

type OutECAddress struct {
	TransAddress
}

var _ IOutECAddress = (*OutECAddress)(nil)

func (e *OutECAddress) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *OutECAddress) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (e *OutECAddress) JSONBuffer(b *bytes.Buffer) error {
	return EncodeJSONToBuffer(e, b)
}

func (b OutECAddress) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (w1 OutECAddress) GetNewInstance() IBlock {
	return new(OutECAddress)
}

func (oa OutECAddress) GetName() string {
	return "outEC"
}

func (a OutECAddress) CustomMarshalText() (text []byte, err error) {
	return a.CustomMarshalTextEC2("ecoutput")
}

/******************************
 * Helper functions
 ******************************/

func NewOutECAddress(address IAddress, amount uint64) IOutAddress {
	oa := new(OutECAddress)
	oa.Amount = amount
	oa.Address = address
	return oa
}
