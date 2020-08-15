// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

/**************************************
 * ISign
 *
 * Interface for RCB Signatures
 *
 * The signature block holds the signatures that validate one of the RCBs.
 * Each signature has an index, so if the RCD is a multisig, you can know
 * how to apply the signatures to the addresses in the RCD.
 **************************************/

type SignatureBlock struct {
	Signatures []interfaces.ISignature `json:"signatures"`
}

var _ interfaces.ISignatureBlock = (*SignatureBlock)(nil)

func (b *SignatureBlock) IsSameAs(s interfaces.ISignatureBlock) bool {
	if s == nil {
		return b == nil
	}

	sigs := s.GetSignatures()
	if len(b.Signatures) != len(sigs) {
		return false
	}
	for i := range b.Signatures {
		if b.Signatures[i].IsSameAs(sigs[i]) == false {
			return false
		}
	}

	return true
}

func (b SignatureBlock) UnmarshalBinary(data []byte) error {
	_, err := b.UnmarshalBinaryData(data)
	return err
}

func (e *SignatureBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *SignatureBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (b SignatureBlock) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (s *SignatureBlock) AddSignature(sig interfaces.ISignature) {
	if len(s.Signatures) > 0 {
		s.Signatures[0] = sig
	} else {
		s.Signatures = append(s.Signatures, sig)
	}
}

func (s SignatureBlock) GetSignature(index int) interfaces.ISignature {
	if len(s.Signatures) <= index {
		return nil
	}
	return s.Signatures[index]
}

func (s SignatureBlock) GetSignatures() []interfaces.ISignature {
	if s.Signatures == nil {
		s.Signatures = make([]interfaces.ISignature, 1, 1)
		s.Signatures[0] = new(FactoidSignature)
	}
	return s.Signatures
}

func (a SignatureBlock) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)
	for _, sig := range a.GetSignatures() {
		err := buf.PushBinaryMarshallable(sig)
		if err != nil {
			return nil, err
		}
	}
	return buf.DeepCopyBytes(), nil
}

func (s SignatureBlock) CustomMarshalText() ([]byte, error) {
	var out primitives.Buffer

	out.WriteString("Signature Block: \n")
	for _, sig := range s.Signatures {
		out.WriteString(" signature: ")
		txt, err := sig.CustomMarshalText()
		if err != nil {
			return nil, err
		}
		out.Write(txt)
		out.WriteString("\n ")

	}

	return out.DeepCopyBytes(), nil
}

func (s *SignatureBlock) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	s.Signatures = make([]interfaces.ISignature, 1)
	s.Signatures[0] = new(FactoidSignature)
	err := buf.PopBinaryMarshallable(s.Signatures[0])
	if err != nil {
		return nil, err
	}
	return buf.DeepCopyBytes(), nil
}

func NewSingleSignatureBlock(priv, data []byte) *SignatureBlock {
	s := new(SignatureBlock)
	s.AddSignature(NewED25519Signature(priv, data))
	return s
}
