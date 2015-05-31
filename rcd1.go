// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package simplecoin

import (
	"bytes"
	"encoding/hex"
	"fmt"
	//    "github.com/agl/ed25519"
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
	IRCD_1
	publicKey []byte
}

var _ IRCD = (*RCD_1)(nil)

func (w RCD_1)Clone() IRCD {
    c := new (RCD_1)
    c.publicKey = make([]byte,len(w.publicKey))
    copy(c.publicKey,w.publicKey)
    return c
}


func (w RCD_1)GetAddress() (IHash, error){
    data, err := w.MarshalBinary()
    if err != nil {
        return nil, fmt.Errorf("This should never happen.  If I have a RCD_1, it should hash.")
    }
    return Sha(data), nil
}

func (w1 RCD_1)GetDBHash() IHash {
    return Sha([]byte("RCD_1"))
}

func (w1 RCD_1)GetNewInstance() IBlock {
    return new(RCD_1)
}

func (a RCD_1) GetPublicKey() []byte {
	return a.publicKey
}

func (w1 RCD_1)NumberOfSignatures() int {
    return 1
}

func (a1 RCD_1) IsEqual(addr IBlock) bool {
	a2, ok := addr.(*RCD_1)
	if !ok || // Not the right kind of IBlock
		bytes.Compare(a1.publicKey, a2.publicKey) != 0 { // Not the right sigature
		return false
	}

	return true
}

func (t *RCD_1) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	typ := int8(data[0])
	data = data[1:]

	if typ != 1 {
		PrtStk()
		return nil, fmt.Errorf("Bad type byte: %d", typ)
	}

	if len(data) < ADDRESS_LENGTH {
		PrtStk()
		return nil, fmt.Errorf("Data source too short to unmarshal an address: %d", len(data))
	}

	t.publicKey = make([]byte, ADDRESS_LENGTH, ADDRESS_LENGTH)
	copy(t.publicKey, data[:ADDRESS_LENGTH])
    data = data[ADDRESS_LENGTH:]
    
	return data, nil
}

func (a RCD_1) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer
	if len(a.publicKey) != ADDRESS_LENGTH {
        Prtln("Length of Public Key: ", len(a.publicKey))
		return nil, fmt.Errorf("Bad length in rcd1.  Should not happen")
	}
	out.WriteByte(byte(1)) // The First Authorization method
	out.Write(a.publicKey)
    
	return out.Bytes(), nil
}

func (a RCD_1) MarshalText() (text []byte, err error) {
	var out bytes.Buffer
	out.WriteString("RCD 1: ")
	WriteNumber8(&out, uint8(1)) // Type Zero Authorization
	out.WriteString(" ")
	out.WriteString(hex.EncodeToString(a.publicKey[:]))
	out.WriteString("\n")

	return out.Bytes(), nil
}
