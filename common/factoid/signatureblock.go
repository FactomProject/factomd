// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
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

// SignatureBlock is an object for holding multiple signatures
type SignatureBlock struct {
	Signatures []interfaces.ISignature `json:"signatures"` // Slice of signatures
}

var _ interfaces.ISignatureBlock = (*SignatureBlock)(nil)

// IsSameAs returns true iff the input signature block contains the same signatures in the same ordering
func (s *SignatureBlock) IsSameAs(si interfaces.ISignatureBlock) bool {
	if si == nil {
		return s == nil
	}

	sigs := si.GetSignatures()
	if len(s.Signatures) != len(sigs) {
		return false
	}
	for i := range s.Signatures {
		if s.Signatures[i].IsSameAs(sigs[i]) == false {
			return false
		}
	}

	return true
}

// UnmarshalBinary unmarshals the input data into this object
func (s SignatureBlock) UnmarshalBinary(data []byte) error {
	_, err := s.UnmarshalBinaryData(data)
	return err
}

// JSONByte returns the json encoded byte array
func (s *SignatureBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(s)
}

// JSONString returns the json encoded string
func (s *SignatureBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(s)
}

// String returns this object as a string
func (s SignatureBlock) String() string {
	txt, err := s.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

// AddSignature appends the input signature to the existing signatures - FIX, is this how its supposed to work? always set to zeroth?
func (s *SignatureBlock) AddSignature(sig interfaces.ISignature) {
	if len(s.Signatures) > 0 {
		s.Signatures[0] = sig
	} else {
		s.Signatures = append(s.Signatures, sig)
	}
}

// GetSignature returns the signature at the input index
func (s SignatureBlock) GetSignature(index int) interfaces.ISignature {
	if len(s.Signatures) <= index {
		return nil
	}
	return s.Signatures[index]
}

// GetSignatures returns the signature slice
func (s SignatureBlock) GetSignatures() []interfaces.ISignature {
	if s.Signatures == nil {
		s.Signatures = make([]interfaces.ISignature, 1, 1)
		s.Signatures[0] = new(FactoidSignature)
	}
	return s.Signatures
}

// MarshalBinary marshals this object
func (s SignatureBlock) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)
	for _, sig := range s.GetSignatures() {
		err := buf.PushBinaryMarshallable(sig)
		if err != nil {
			return nil, err
		}
	}
	return buf.DeepCopyBytes(), nil
}

// CustomMarshalText returns this object as a custom string
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

// UnmarshalBinaryData unmarshals the input data into this object
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

// NewSingleSignatureBlock creates a new signature block with a single signature created from the
// input private key and data
func NewSingleSignatureBlock(priv, data []byte) *SignatureBlock {
	s := new(SignatureBlock)
	s.AddSignature(NewED25519Signature(priv, data))
	return s
}
