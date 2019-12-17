// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// FactoidSignature is an object for holding a single signature, currently at 64 bytes length.
// The default FactoidSignature doesn't care about indexing.  We will extend this
// FactoidSignature for multisig
type FactoidSignature struct {
	Signature [constants.SIGNATURE_LENGTH]byte `json:"signature"` // The FactoidSignature
}

var _ interfaces.ISignature = (*FactoidSignature)(nil)

// IsSameAs returns true iff the input object is identical to this object
func (s *FactoidSignature) IsSameAs(sig interfaces.ISignature) bool {
	return primitives.AreBytesEqual(s.Bytes(), sig.Bytes())
}

// Verify returns a hard coded true, with a print to the screen saying this function is broken
func (s *FactoidSignature) Verify([]byte) bool {
	fmt.Println("Verify is Broken")
	return true
}

// Bytes returns the signature
func (s *FactoidSignature) Bytes() []byte {
	return s.Signature[:]
}

// GetKey returns the public key part of the signature (the last 32 bytes)
func (s *FactoidSignature) GetKey() []byte {
	return s.Signature[32:]
}

// MarshalText encodes the signature to a hexidecimal string, then returns its bytes
func (s *FactoidSignature) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "FactoidSignature.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(hex.EncodeToString(s.Signature[:])), nil
}

// JSONByte returns the json encoded byte array
func (s *FactoidSignature) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(s)
}

// JSONString returns the json encoded string
func (s *FactoidSignature) JSONString() (string, error) {
	return primitives.EncodeJSONString(s)
}

// String returns this object as a string
func (s FactoidSignature) String() string {
	txt, err := s.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

// SetSignature sets the signature to the input. We only have one FactoidSignature
func (s *FactoidSignature) SetSignature(sig []byte) error {
	if len(sig) != constants.SIGNATURE_LENGTH {
		return fmt.Errorf("Bad FactoidSignature.  Should not happen")
	}
	copy(s.Signature[:], sig)
	return nil
}

// GetSignature returns thte signature pointer
func (s *FactoidSignature) GetSignature() *[constants.SIGNATURE_LENGTH]byte {
	return &s.Signature
}

// MarshalBinary marshals this object
func (s FactoidSignature) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(s.Signature[:])
	return buf.DeepCopyBytes(), nil
}

// CustomMarshalText returns a custom string " FactoidSignature: <hexidecimal_sig>\n"
func (s FactoidSignature) CustomMarshalText() ([]byte, error) {
	var out primitives.Buffer

	out.WriteString(" FactoidSignature: ")
	out.WriteString(hex.EncodeToString(s.Signature[:]))
	out.WriteString("\n")

	return out.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (s *FactoidSignature) UnmarshalBinaryData(data []byte) ([]byte, error) {
	if data == nil || len(data) < constants.SIGNATURE_LENGTH {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	copy(s.Signature[:], data[:constants.SIGNATURE_LENGTH])
	return data[constants.SIGNATURE_LENGTH:], nil
}

// UnmarshalBinary unmarshals the input data into this object
func (s *FactoidSignature) UnmarshalBinary(data []byte) error {
	_, err := s.UnmarshalBinaryData(data)
	return err
}

// NewED25519Signature returns a new signature for the input private key and data
func NewED25519Signature(priv, data []byte) *FactoidSignature {
	sig := primitives.Sign(priv, data)
	fs := new(FactoidSignature)
	copy(fs.Signature[:], sig[:constants.SIGNATURE_LENGTH])
	return fs
}
