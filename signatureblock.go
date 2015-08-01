// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"bytes"
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
	signatures []ISignature
}

var _ ISignatureBlock = (*SignatureBlock)(nil)
var _ IBlock          = (*SignatureBlock)(nil)
func (b SignatureBlock) GetHash() IHash { return nil }

func (b SignatureBlock) UnmarshalBinary(data []byte) error { 
    _, err := b.UnmarshalBinaryData(data)
    return err
}

func (b SignatureBlock) String() string {
	txt, err := b.MarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (s *SignatureBlock) IsEqual(signatureBlock IBlock) []IBlock {

	sb, ok := signatureBlock.(ISignatureBlock)

	if !ok {
        r := make([]IBlock,0,5)
        return append(r,s)
	}

	sigs1 := s.GetSignatures()
	sigs2 := sb.GetSignatures()
	if len(sigs1) != len(sigs2) {
        r := make([]IBlock,0,5)
        return append(r,s)
	}
	for i, sig := range sigs1 {
		r := sig.IsEqual(sigs2[i]) 
        if r != nil {
			return append(r,s)
		}
	}

	return nil
}

func (s *SignatureBlock) AddSignature(sig ISignature) {
	if len(s.signatures)>0 {
        s.signatures[0]=sig
    }else{
        s.signatures = append(s.signatures, sig)
    }
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
		s.signatures = make([]ISignature, 1, 1)
        s.signatures[0] = new(Signature)
	}
	return s.signatures
}

func (a SignatureBlock) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	for _, sig := range a.GetSignatures() {
		data, err := sig.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("Signature failed to Marshal in RCD_1")
		}
		out.Write(data)
	}

	return out.Bytes(), nil
}

func (s SignatureBlock) MarshalText() ([]byte, error) {
	var out bytes.Buffer

	out.WriteString("Signature Block: \n")
	for _, sig := range s.signatures {

		out.WriteString(" signature: ")
		txt, err := sig.MarshalText()
		if err != nil {
			return nil, err
		}
		out.Write(txt)
		out.WriteString("\n ")

	}

	return out.Bytes(), nil
}

func (s *SignatureBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	s.signatures = make([]ISignature, 1)
    s.signatures[0] = new(Signature)
    data, err = s.signatures[0].UnmarshalBinaryData(data)
    if err != nil {
        fmt.Println("error")
        return nil, fmt.Errorf("Failure to unmarshal Signature")
    }
	
	return data, nil
}
