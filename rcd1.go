// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/ed25519"
)

/**************************
 * RCD_1 Simple Signature
 **************************/

type IRCD_1 interface {
	IRCD
	GetPublicKey() []byte
}

// In this case, we are simply validating one address to ensure it signed
// this transaction.
type RCD_1 struct {
	publicKey [ADDRESS_LENGTH]byte
}

var _ IRCD = (*RCD_1)(nil)

/*************************************
 *       Stubs
 *************************************/

func (b RCD_1) GetHash() IHash {
    return nil
}

/***************************************
 *       Methods
 ***************************************/

func (b RCD_1   ) UnmarshalBinary(data []byte) error { 
    _, err := b.UnmarshalBinaryData(data)
    return err
}
func (b RCD_1) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (w RCD_1) CheckSig(trans ITransaction, sigblk ISignatureBlock) bool {
    if sigblk == nil { return false }
	data, err := trans.MarshalBinarySig()
	if err != nil {
      	return false
	}
	signature := sigblk.GetSignature(0)
    if signature == nil { return false }
    cryptosig := signature.GetSignature()
    if cryptosig == nil { return false }
    
	return ed25519.VerifyCanonical(&w.publicKey, data, cryptosig) 
}

func (w RCD_1) Clone() IRCD {
	c := new(RCD_1)
	copy(c.publicKey[:], w.publicKey[:])
	return c
}

func (w RCD_1) GetAddress() (IAddress, error) {
	return CreateAddress(Shad(w.publicKey[:])), nil
}

func (RCD_1) GetDBHash() IHash {
	return Sha([]byte("RCD_1"))
}

func (w1 RCD_1) GetNewInstance() IBlock {
	return new(RCD_1)
}

func (a RCD_1) GetPublicKey() []byte {
	return a.publicKey[:]
}

func (w1 RCD_1) NumberOfSignatures() int {
	return 1
}

func (a1 *RCD_1) IsEqual(addr IBlock) []IBlock {
	a2, ok := addr.(*RCD_1)

	if	!ok || a1.publicKey != a2.publicKey { // Not the right object or sigature
            r := make([]IBlock,0,5)
            return append(r,a1) 
	}

	return nil
}

func (t *RCD_1) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	typ := int8(data[0])
	data = data[1:]

	if typ != 1 {
		return nil, fmt.Errorf("Bad type byte: %d", typ)
	}

	if len(data) < ADDRESS_LENGTH {
		return nil, fmt.Errorf("Data source too short to unmarshal an address: %d", len(data))
	}

	copy(t.publicKey[:], data[:ADDRESS_LENGTH])
	data = data[ADDRESS_LENGTH:]

	return data, nil
}

func (a RCD_1) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer
	out.WriteByte(byte(1)) // The First Authorization method
	out.Write(a.publicKey[:])

	return out.Bytes(), nil
}

func (a RCD_1) CustomMarshalText() (text []byte, err error) {
	var out bytes.Buffer
	out.WriteString("RCD 1: ")
	WriteNumber8(&out, uint8(1)) // Type Zero Authorization
	out.WriteString(" ")
	out.WriteString(hex.EncodeToString(a.publicKey[:]))
	out.WriteString("\n")

	return out.Bytes(), nil
}
