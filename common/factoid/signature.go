// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)

// The default FactoidSignature doesn't care about indexing.  We will extend this
// FactoidSignature for multisig
type FactoidSignature struct {
	FactoidSignature [constants.SIGNATURE_LENGTH]byte // The FactoidSignature
}

var _ interfaces.ISignature = (*FactoidSignature)(nil)

func (t *FactoidSignature) GetHash() interfaces.IHash {
	return nil
}

func (e *FactoidSignature) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *FactoidSignature) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (e *FactoidSignature) JSONBuffer(b *bytes.Buffer) error {
	return EncodeJSONToBuffer(e, b)
}

func (b FactoidSignature) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (w1 FactoidSignature) GetNewInstance() interfaces.IBlock {
	return new(FactoidSignature)
}

// Checks that the FactoidSignatures are the same.
func (s1 *FactoidSignature) IsEqual(sig interfaces.IBlock) []interfaces.IBlock {
	s2, ok := sig.(*FactoidSignature)
	if !ok || // Not the right kind of interfaces.IBlock
		s1.FactoidSignature != s2.FactoidSignature { // Not the right rcd
		r := make([]interfaces.IBlock, 0, 5)
		return append(r, s1)
	}
	return nil
}

// Index is ignored.  We only have one FactoidSignature
func (s *FactoidSignature) SetSignature(sig []byte) error {
	if len(sig) != constants.SIGNATURE_LENGTH {
		return fmt.Errorf("Bad FactoidSignature.  Should not happen")
	}
	copy(s.FactoidSignature[:], sig)
	return nil
}

func (s *FactoidSignature) GetSignature() *[constants.SIGNATURE_LENGTH]byte {
	return &s.FactoidSignature
}

func (s FactoidSignature) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	out.Write(s.FactoidSignature[:])

	return out.Bytes(), nil
}
func (b FactoidSignature) MarshalledSize() uint64 {
	hex, _ := b.MarshalBinary()
	return uint64(len(hex))
}

func (s FactoidSignature) CustomMarshalText() ([]byte, error) {
	var out bytes.Buffer

	out.WriteString(" FactoidSignature: ")
	out.WriteString(hex.EncodeToString(s.FactoidSignature[:]))
	out.WriteString("\n")

	return out.Bytes(), nil
}

func (s *FactoidSignature) UnmarshalBinaryData(data []byte) ([]byte, error) {
	copy(s.FactoidSignature[:], data[:constants.SIGNATURE_LENGTH])
	return data[constants.SIGNATURE_LENGTH:], nil
}

func (s *FactoidSignature) UnmarshalBinary(data []byte) error {
	_, err := s.UnmarshalBinaryData(data)
	return err
}
