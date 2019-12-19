// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

/**************************
 * RCD_1 Simple Signature
 **************************/

// RCD_1 is a Redeem Condition Datastructure (RCD) of type 1, which is simply validating one address to ensure it signed
// this transaction.
type RCD_1 struct {
	PublicKey [constants.ADDRESS_LENGTH]byte
	validSig  bool
}

var _ interfaces.IRCD = (*RCD_1)(nil)

/***************************************
 *       Methods
 ***************************************/

// IsSameAs checks whether the incoming RCD is the same as this RCD
func (r RCD_1) IsSameAs(rcd interfaces.IRCD) bool {
	return r.String() == rcd.String()
}

// UnmarshalBinary unmarshals the input data into this object
func (r RCD_1) UnmarshalBinary(data []byte) error {
	_, err := r.UnmarshalBinaryData(data)
	return err
}

// JSONByte returns the json encoded byte array
func (r *RCD_1) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(r)
}

// JSONString returns the json encoded string
func (r *RCD_1) JSONString() (string, error) {
	return primitives.EncodeJSONString(r)
}

// MarshalJSON will return the json encoding of the string "<rcd_type><public_key>"
func (r *RCD_1) MarshalJSON() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "RCD_1.MarshalJSON err:%v", *pe)
		}
	}(&err)
	return json.Marshal(fmt.Sprintf("%x", append([]byte{0x01}, r.PublicKey[:]...)))
}

// String returns this object as a string
func (r RCD_1) String() string {
	txt, err := r.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

// MarshalText returns the public key as a byte array
func (r *RCD_1) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "RCD_1.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(hex.EncodeToString(r.PublicKey[:])), nil
}

// CheckSig returns validSig if its already been found to be true, otherwise sets and returns the internal validSig bool
// by verifying transaction using the public key has been properly signed by the input signature
func (r *RCD_1) CheckSig(trans interfaces.ITransaction, sigblk interfaces.ISignatureBlock) bool {
	if r.validSig {
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

	r.validSig = ed25519.VerifyCanonical(&r.PublicKey, data, cryptosig)

	return r.validSig
}

// Clone creates a copy of this RCD
func (r RCD_1) Clone() interfaces.IRCD {
	c := new(RCD_1)
	copy(c.PublicKey[:], r.PublicKey[:])
	return c
}

// GetAddress returns a new address created by sha256(sha256(public key))
func (r RCD_1) GetAddress() (interfaces.IAddress, error) {
	data := []byte{1}
	data = append(data, r.PublicKey[:]...)
	return CreateAddress(primitives.Shad(data)), nil
}

// GetPublicKey returns the public key of the RCD
func (r RCD_1) GetPublicKey() []byte {
	return r.PublicKey[:]
}

// NumberOfSignatures returns the number of signatures (a hardcoded 1 for RCD type 1 objects)
func (r RCD_1) NumberOfSignatures() int {
	return 1
}

// UnmarshalBinaryData unmarshals the input data into this object
func (r *RCD_1) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	if data == nil || len(data) < 1+constants.ADDRESS_LENGTH {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	typ := int8(data[0])
	data = data[1:]

	if typ != 1 {
		return nil, fmt.Errorf("Bad type byte: %d", typ)
	}

	if len(data) < constants.ADDRESS_LENGTH {
		return nil, fmt.Errorf("Data source too short to unmarshal an address: %d", len(data))
	}

	copy(r.PublicKey[:], data[:constants.ADDRESS_LENGTH])
	data = data[constants.ADDRESS_LENGTH:]

	return data, nil
}

// MarshalBinary marshals the object
func (r RCD_1) MarshalBinary() ([]byte, error) {
	var out primitives.Buffer
	out.WriteByte(byte(1)) // The First Authorization method
	out.Write(r.PublicKey[:])

	return out.DeepCopyBytes(), nil
}

// CustomMarshalText writes this RCD as a string "RCD 1: 1 <hexidecimal_public_key>\n" and returns its byte array
func (r RCD_1) CustomMarshalText() (text []byte, err error) {
	var out primitives.Buffer
	out.WriteString("RCD 1: ")
	primitives.WriteNumber8(&out, uint8(1)) // Type Zero Authorization
	out.WriteString(" ")
	out.WriteString(hex.EncodeToString(r.PublicKey[:]))
	out.WriteString("\n")

	return out.DeepCopyBytes(), nil
}
