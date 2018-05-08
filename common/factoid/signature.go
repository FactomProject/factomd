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

// The default FactoidSignature doesn't care about indexing.  We will extend this
// FactoidSignature for multisig
type FactoidSignature struct {
	Signature [constants.SIGNATURE_LENGTH]byte `json:"signature"` // The FactoidSignature
}

var _ interfaces.ISignature = (*FactoidSignature)(nil)

func (s *FactoidSignature) IsSameAs(sig interfaces.ISignature) bool {
	return primitives.AreBytesEqual(s.Bytes(), sig.Bytes())
}

func (s *FactoidSignature) Verify([]byte) bool {
	fmt.Println("Verify is Broken")
	return true
}

func (sig *FactoidSignature) Bytes() []byte {
	return sig.Signature[:]
}

func (s *FactoidSignature) GetKey() []byte {
	return s.Signature[32:]
}

func (h *FactoidSignature) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "FactoidSignature.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(hex.EncodeToString(h.Signature[:])), nil
}

func (s *FactoidSignature) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(s)
}

func (s *FactoidSignature) JSONString() (string, error) {
	return primitives.EncodeJSONString(s)
}

func (s FactoidSignature) String() string {
	txt, err := s.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

// Index is ignored.  We only have one FactoidSignature
func (s *FactoidSignature) SetSignature(sig []byte) error {
	if len(sig) != constants.SIGNATURE_LENGTH {
		return fmt.Errorf("Bad FactoidSignature.  Should not happen")
	}
	copy(s.Signature[:], sig)
	return nil
}

func (s *FactoidSignature) GetSignature() *[constants.SIGNATURE_LENGTH]byte {
	return &s.Signature
}

func (s FactoidSignature) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(s.Signature[:])
	return buf.DeepCopyBytes(), nil
}

func (s FactoidSignature) CustomMarshalText() ([]byte, error) {
	var out primitives.Buffer

	out.WriteString(" FactoidSignature: ")
	out.WriteString(hex.EncodeToString(s.Signature[:]))
	out.WriteString("\n")

	return out.DeepCopyBytes(), nil
}

func (s *FactoidSignature) UnmarshalBinaryData(data []byte) ([]byte, error) {
	if data == nil || len(data) < constants.SIGNATURE_LENGTH {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	copy(s.Signature[:], data[:constants.SIGNATURE_LENGTH])
	return data[constants.SIGNATURE_LENGTH:], nil
}

func (s *FactoidSignature) UnmarshalBinary(data []byte) error {
	_, err := s.UnmarshalBinaryData(data)
	return err
}

func NewED25519Signature(priv, data []byte) *FactoidSignature {
	sig := primitives.Sign(priv, data)
	fs := new(FactoidSignature)
	copy(fs.Signature[:], sig[:constants.SIGNATURE_LENGTH])
	return fs
}
