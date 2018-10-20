// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

/************************
 * RCD 2
 ************************/

// Type 2 RCD implement multisig
// m of n
// Must have m addresses from which to choose, no fewer, no more
// Must have n RCD, no fewer no more.
// NOTE: This does mean you can have a multisig nested in a
// multisig.  It just works.

type RCD_2 struct {
	M           int                   // Number signatures required
	N           int                   // Total sigatures possible
	N_Addresses []interfaces.IAddress // n addresses
}

var _ interfaces.IRCD = (*RCD_2)(nil)

/*************************************
 *       Stubs
 *************************************/

func (b RCD_2) GetAddress() (interfaces.IAddress, error) {
	return nil, nil
}

func (b RCD_2) NumberOfSignatures() int {
	return 1
}

/***************************************
 *       Methods
 ***************************************/

func (b RCD_2) IsSameAs(rcd interfaces.IRCD) bool {
	return b.String() == rcd.String()
}

func (b RCD_2) UnmarshalBinary(data []byte) error {
	_, err := b.UnmarshalBinaryData(data)
	return err
}

func (b RCD_2) CheckSig(trans interfaces.ITransaction, sigblk interfaces.ISignatureBlock) bool {
	return false
}

func (e *RCD_2) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

// TODO: Fix Json marshaling of RCD_2. Right now the RCD type
// is not included in the json marshal.
func (e *RCD_2) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (b RCD_2) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (w RCD_2) Clone() interfaces.IRCD {
	c := new(RCD_2)
	c.M = w.M
	c.N = w.N
	c.N_Addresses = make([]interfaces.IAddress, len(w.N_Addresses))
	for i, address := range w.N_Addresses {
		c.N_Addresses[i] = CreateAddress(address)
	}
	return c
}

func (t *RCD_2) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	if data == nil || len(data) < 5 {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	typ := int8(data[0])
	data = data[1:]
	if typ != 2 {
		return nil, fmt.Errorf("Bad data fed to RCD_2 UnmarshalBinaryData()")
	}

	t.N, data = int(binary.BigEndian.Uint16(data[0:2])), data[2:]
	t.M, data = int(binary.BigEndian.Uint16(data[0:2])), data[2:]
	// TODO: remove printing unmarshal count numbers once we have good data on
	// what they should be.
	log.Print("RCD_2 unmarshaled signatures possible: ", t.N)
	log.Print("RCD_2 unmarshaled signatures required: ", t.M)
	if t.N > 1000 {
		// TODO: replace this message with a proper error
		return nil, fmt.Errorf("Error: RCD_2.UnmarshalBinary: signatures possible too many (uint underflow?)")
	}
	if t.M > 1000 {
		// TODO: replace this message with a proper error
		return nil, fmt.Errorf("Error: RCD_2.UnmarshalBinary: signatures required too many (uint underflow?)")
	}

	t.N_Addresses = make([]interfaces.IAddress, t.M, t.M)

	for i, _ := range t.N_Addresses {
		t.N_Addresses[i] = new(Address)
		data, err = t.N_Addresses[i].UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (a RCD_2) MarshalBinary() ([]byte, error) {
	var out primitives.Buffer

	binary.Write(&out, binary.BigEndian, uint8(2))
	binary.Write(&out, binary.BigEndian, uint16(a.N))
	binary.Write(&out, binary.BigEndian, uint16(a.M))
	for i := 0; i < a.M; i++ {
		data, err := a.N_Addresses[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
	}

	return out.DeepCopyBytes(), nil
}

func (a RCD_2) CustomMarshalText() ([]byte, error) {
	var out primitives.Buffer

	primitives.WriteNumber8(&out, uint8(2)) // Type 2 Authorization
	out.WriteString("\n n: ")
	primitives.WriteNumber16(&out, uint16(a.N))
	out.WriteString(" m: ")
	primitives.WriteNumber16(&out, uint16(a.M))
	out.WriteString("\n")
	for i := 0; i < a.M; i++ {
		out.WriteString("  m: ")
		out.WriteString(hex.EncodeToString(a.N_Addresses[i].Bytes()))
		out.WriteString("\n")
	}

	return out.DeepCopyBytes(), nil
}
