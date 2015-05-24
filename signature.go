// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package simplecoin

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	//    "github.com/agl/ed25519"
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
	GetIndex() int       // Used with multisig to know where to apply the signature
	SetIndex(int)        // Set the index
	Validate(hash IHash) // Signatures are validated against the Hash of digital things.
	SetSignature([]byte) // Set or update the signature
}

// We need an index into m.  We could search, but that could make transaction
// processing time slow.
type Signature struct {
	ISignature
	index     int                    // Index into m for this signature
	signature [SIGNATURE_LENGTH]byte // The signature
}

var _ ISignature = (*Signature)(nil)

func (w1 Signature)GetDBHash() IHash {
    return Sha([]byte("Signature"))
}

func (w1 Signature)GetNewInstance() IBlock {
    return new(Signature)
}

// Checks that the signatures are the same.  The index does NOT have to be the same.
// This way, you can check if two signatures on the same transaction are actually
// the same. Or if two transactions are the same.  We don't know what you are signing,
// or why you might need to compare signatures.
func (s1 Signature) IsEqual(sig IBlock) bool {
	s2, ok := sig.(*Signature)
	if !ok || // Not the right kind of IBlock
		s1.signature != s2.signature { // Not the right rcd
		return false
	}
	return true
}

func (s *Signature) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	if len(data) < 2 {
		return nil, fmt.Errorf("Data source too short to unmarshal a Signature: %d", len(data))
	}

	s.index, data = int(binary.BigEndian.Uint16(data[0:2])), data[2:]
	copy(s.signature[:], data[:SIGNATURE_LENGTH])
	data = data[SIGNATURE_LENGTH:]

	return data, nil
}

func (s Signature) GetIndex() int {
	return s.index
}

func (s *Signature) SetIndex(i int) {
	s.index = i
}

func (s *Signature) SetSignature(sig []byte) {
	if len(sig) != SIGNATURE_LENGTH {
		panic("Bad signature.  Should not happen")
	}
	copy(s.signature[:], sig)
}

func (s Signature) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	binary.Write(&out, binary.BigEndian, uint16(s.index))
	out.Write(s.signature[:])

	return out.Bytes(), nil
}

func (s Signature) MarshalText() ([]byte, error) {
	var out bytes.Buffer

	out.WriteString("index: ")
	WriteNumber16(&out, uint16(s.index))
	out.WriteString(" signature: ")
	out.WriteString(hex.EncodeToString(s.signature[:]))
	out.WriteString("\n")

	return out.Bytes(), nil
}
