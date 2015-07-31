// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
type ISignatureBlock interface {
	IBlock
	GetSignatures() []ISignature
	AddSignature(sig ISignature)
	GetSignature(int) ISignature
}

type SignatureBlock struct {
	ISignatureBlock `json:"-"`
	signatures      []ISignature
}

var _ ISignatureBlock = (*SignatureBlock)(nil)

func (b SignatureBlock) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (s *SignatureBlock) IsEqual(signatureBlock IBlock) []IBlock {

	sb, ok := signatureBlock.(ISignatureBlock)

	if !ok {
		r := make([]IBlock, 0, 5)
		return append(r, s)
	}

	sigs1 := s.GetSignatures()
	sigs2 := sb.GetSignatures()
	if len(sigs1) != len(sigs2) {
		r := make([]IBlock, 0, 5)
		return append(r, s)
	}
	for i, sig := range sigs1 {
		r := sig.IsEqual(sigs2[i])
		if r != nil {
			return append(r, s)
		}
	}

	return nil
}

func (s *SignatureBlock) AddSignature(sig ISignature) {
	s.signatures = append(s.signatures, sig)
}

func (s SignatureBlock) GetSignature(index int) ISignature {
	if len(s.signatures) <= index {
		return nil
	}
	return s.signatures[index]
}

func (SignatureBlock) GetDBHash() IHash {
	return Sha([]byte("SignatureBlock"))
}

func (s SignatureBlock) GetNewInstance() IBlock {
	return new(SignatureBlock)
}

func (s SignatureBlock) GetSignatures() []ISignature {
	if s.signatures == nil {
		s.signatures = make([]ISignature, 0, 1)
	}
	return s.signatures
}

func (a SignatureBlock) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	binary.Write(&out, binary.BigEndian, uint16(len(a.signatures)))
	for _, sig := range a.GetSignatures() {

		data, err := sig.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("Signature failed to Marshal in RCD_1")
		}
		out.Write(data)
	}

	return out.Bytes(), nil
}

func (s SignatureBlock) CustomMarshalText() ([]byte, error) {
	var out bytes.Buffer

	out.WriteString("Signature Block: ")
	WriteNumber16(&out, uint16(len(s.signatures)))
	out.WriteString("\n")
	for _, sig := range s.signatures {

		out.WriteString(" signature: ")
		txt, err := sig.CustomMarshalText()
		if err != nil {
			return nil, err
		}
		out.Write(txt)
		out.WriteString("\n ")

	}

	return out.Bytes(), nil
}

func (s *SignatureBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	numSignatures, data := binary.BigEndian.Uint16(data[0:2]), data[2:]
	s.signatures = make([]ISignature, numSignatures)
	for i := uint16(0); i < numSignatures; i++ {
		s.signatures[i] = new(Signature)
		data, err = s.signatures[i].UnmarshalBinaryData(data)
		if err != nil {
			fmt.Println("error")
			return nil, fmt.Errorf("Failure to unmarshal Signature")
		}
	}

	return data, nil
}
