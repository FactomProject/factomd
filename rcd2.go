// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
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
	m           int        // Number signatures required
	n           int        // Total sigatures possible
	n_addresses []IAddress // n addresses
}

var _ IRCD = (*RCD_2)(nil)

/*************************************
 *       Stubs
 *************************************/

func (b RCD_2) GetAddress() (IAddress, error) {
	return nil, nil
}

func (b RCD_2) GetHash() IHash {
	return nil
}

func (b RCD_2) NumberOfSignatures() int {
	return 1
}

/***************************************
 *       Methods
 ***************************************/

func (b RCD_2) UnmarshalBinary(data []byte) error {
	_, err := b.UnmarshalBinaryData(data)
	return err
}

func (b RCD_2) CheckSig(trans ITransaction, sigblk ISignatureBlock) bool {
	return false
}

func (b RCD_2) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (w RCD_2) Clone() IRCD {
	c := new(RCD_2)
	c.m = w.m
	c.n = w.n
	c.n_addresses = make([]IAddress, len(w.n_addresses))
	for i, address := range w.n_addresses {
		c.n_addresses[i] = CreateAddress(address)
	}
	return c
}

func (RCD_2) GetDBHash() IHash {
	return Sha([]byte("RCD_2"))
}

func (w1 RCD_2) GetNewInstance() IBlock {
	return new(RCD_2)
}

func (a1 *RCD_2) IsEqual(addr IBlock) []IBlock {
	a2, ok := addr.(*RCD_2)
	if !ok || // Not the right kind of IBlock
		a1.n != a2.n || // Size of sig has to match
		a1.m != a2.m || // Size of sig has to match
		len(a1.n_addresses) != len(a2.n_addresses) { // Size of arrays has to match
		r := make([]IBlock, 0, 5)
		return append(r, a1)
	}

	for i, addr := range a1.n_addresses {
		r := addr.IsEqual(a2.n_addresses[i])
		if r != nil {
			return append(r, a1)
		}
	}

	return nil
}

func (t *RCD_2) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	typ := int8(data[0])
	data = data[1:]
	if typ != 2 {
		return nil, fmt.Errorf("Bad data fed to RCD_2 UnmarshalBinaryData()")
	}

	t.n, data = int(binary.BigEndian.Uint16(data[0:2])), data[2:]
	t.m, data = int(binary.BigEndian.Uint16(data[0:2])), data[2:]

	t.n_addresses = make([]IAddress, t.m, t.m)

	for i, _ := range t.n_addresses {
		t.n_addresses[i] = new(Address)
		data, err = t.n_addresses[i].UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (a RCD_2) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	binary.Write(&out, binary.BigEndian, uint8(2))
	binary.Write(&out, binary.BigEndian, uint16(a.n))
	binary.Write(&out, binary.BigEndian, uint16(a.m))
	for i := 0; i < a.m; i++ {
		data, err := a.n_addresses[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
	}

	return out.Bytes(), nil
}

func (a RCD_2) CustomMarshalText() ([]byte, error) {
	var out bytes.Buffer

	WriteNumber8(&out, uint8(2)) // Type 2 Authorization
	out.WriteString("\n n: ")
	WriteNumber16(&out, uint16(a.n))
	out.WriteString(" m: ")
	WriteNumber16(&out, uint16(a.m))
	out.WriteString("\n")
	for i := 0; i < a.m; i++ {
		out.WriteString("  m: ")
		out.WriteString(hex.EncodeToString(a.n_addresses[i].Bytes()))
		out.WriteString("\n")
	}

	return out.Bytes(), nil
}
