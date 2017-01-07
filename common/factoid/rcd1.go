// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

/**************************
 * RCD_1 Simple Signature
 **************************/

// In this case, we are simply validating one address to ensure it signed
// this transaction.
type RCD_1 struct {
	PublicKey [constants.ADDRESS_LENGTH]byte
	validSig  bool
}

var _ interfaces.IRCD = (*RCD_1)(nil)

/*************************************
 *       Stubs
 *************************************/

func (b RCD_1) GetHash() interfaces.IHash {
	return nil
}

/***************************************
 *       Methods
 ***************************************/

func (b RCD_1) UnmarshalBinary(data []byte) error {
	_, err := b.UnmarshalBinaryData(data)
	return err
}

func (e *RCD_1) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *RCD_1) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *RCD_1) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (b RCD_1) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (r *RCD_1) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(r.PublicKey[:])), nil
}

func (w RCD_1) CheckSig(trans interfaces.ITransaction, sigblk interfaces.ISignatureBlock) bool {
	if w.validSig {
		return true
	}
	if sigblk == nil {
		return false
	}
	data, err := trans.MarshalBinarySig()
	if err != nil {
		return false
	}
	signature := sigblk.GetSignature(0)
	if signature == nil {
		return false
	}
	cryptosig := signature.GetSignature()
	if cryptosig == nil {
		return false
	}

	w.validSig = ed25519.VerifyCanonical(&w.PublicKey, data, cryptosig)

	return w.validSig
}

func (w RCD_1) Clone() interfaces.IRCD {
	c := new(RCD_1)
	copy(c.PublicKey[:], w.PublicKey[:])
	return c
}

func (w RCD_1) GetAddress() (interfaces.IAddress, error) {
	data := []byte{1}
	data = append(data, w.PublicKey[:]...)
	return CreateAddress(primitives.Shad(data)), nil
}

func (a RCD_1) GetPublicKey() []byte {
	return a.PublicKey[:]
}

func (w1 RCD_1) NumberOfSignatures() int {
	return 1
}

func (a1 *RCD_1) IsEqual(addr interfaces.IBlock) []interfaces.IBlock {
	a2, ok := addr.(*RCD_1)

	if !ok || a1.PublicKey != a2.PublicKey { // Not the right object or sigature
		r := make([]interfaces.IBlock, 0, 5)
		return append(r, a1)
	}

	return nil
}

func (t *RCD_1) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	typ := int8(data[0])
	data = data[1:]

	if typ != 1 {
		return nil, fmt.Errorf("Bad type byte: %d", typ)
	}

	if len(data) < constants.ADDRESS_LENGTH {
		return nil, fmt.Errorf("Data source too short to unmarshal an address: %d", len(data))
	}

	copy(t.PublicKey[:], data[:constants.ADDRESS_LENGTH])
	data = data[constants.ADDRESS_LENGTH:]

	return data, nil
}

func (a RCD_1) MarshalBinary() ([]byte, error) {
	var out primitives.Buffer
	out.WriteByte(byte(1)) // The First Authorization method
	out.Write(a.PublicKey[:])

	return out.DeepCopyBytes(), nil
}

func (a RCD_1) CustomMarshalText() (text []byte, err error) {
	var out primitives.Buffer
	out.WriteString("RCD 1: ")
	primitives.WriteNumber8(&out, uint8(1)) // Type Zero Authorization
	out.WriteString(" ")
	out.WriteString(hex.EncodeToString(a.PublicKey[:]))
	out.WriteString("\n")

	return out.DeepCopyBytes(), nil
}
