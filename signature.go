// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

/**************************************
 * ISign
 *
 * Interface for RCB Signatures
 *
 * Data structure to support signatures, including multisig.
 **************************************/

type ISignature interface {
	IBlock
	SetSignature(sig []byte) error // Set or update the signature
	GetSignature() *[SIGNATURE_LENGTH]byte
}

// The default signature doesn't care about indexing.  We will extend this
// signature for multisig
type Signature struct {
	signature [SIGNATURE_LENGTH]byte // The signature
}

var _ ISignature = (*Signature)(nil)

func (t *Signature) GetHash() IHash {
    return nil
}

func (b Signature) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (Signature) GetDBHash() IHash {
	return Sha([]byte("Signature"))
}

func (w1 Signature) GetNewInstance() IBlock {
	return new(Signature)
}

// Checks that the signatures are the same.
func (s1 *Signature) IsEqual(sig IBlock) []IBlock {
	s2, ok := sig.(*Signature)
	if !ok || // Not the right kind of IBlock
		s1.signature != s2.signature { // Not the right rcd
		r := make([]IBlock, 0, 5)
		return append(r, s1)
	}
	return nil
}

// Index is ignored.  We only have one signature
func (s *Signature) SetSignature(sig []byte) error {
	if len(sig) != SIGNATURE_LENGTH {
		return fmt.Errorf("Bad signature.  Should not happen")
	}
	copy(s.signature[:], sig)
	return nil
}

func (s *Signature) GetSignature() *[SIGNATURE_LENGTH]byte {
	return &s.signature
}

func (s Signature) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	out.Write(s.signature[:])

	return out.Bytes(), nil
}

func (s Signature) CustomMarshalText() ([]byte, error) {
	var out bytes.Buffer

	out.WriteString(" signature: ")
	out.WriteString(hex.EncodeToString(s.signature[:]))
	out.WriteString("\n")

	return out.Bytes(), nil
}

func (s *Signature) UnmarshalBinaryData(data []byte) ([]byte, error) {
	copy(s.signature[:], data[:SIGNATURE_LENGTH])
	return data[SIGNATURE_LENGTH:], nil
}

func (s *Signature) UnmarshalBinary(data []byte) error {
    _, err := s.UnmarshalBinaryData(data)
    return err
}