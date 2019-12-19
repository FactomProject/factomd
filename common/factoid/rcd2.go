// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

/************************
 * RCD 2
 ************************/

// RCD_2 is a type 2 RCD implementing multisignatures
// m of n
// Must have m addresses from which to choose, no fewer, no more
// Must have n RCD, no fewer no more.
// NOTE: This does mean you can have a multisig nested in a
// multisig.  It just works.
type RCD_2 struct {
	M           int                   // Number signatures required
	N           int                   // Total signatures possible
	N_Addresses []interfaces.IAddress // n addresses
}

var _ interfaces.IRCD = (*RCD_2)(nil)

/*************************************
 *       Stubs
 *************************************/

// GetAddress returns nil,nil always
func (r RCD_2) GetAddress() (interfaces.IAddress, error) {
	return nil, nil
}

// NumberOfSignatures returns a hardcoded 1 always
func (r RCD_2) NumberOfSignatures() int {
	return 1
}

/***************************************
 *       Methods
 ***************************************/

// IsSameAs returns true iff the input rcd is identical to this one
func (r RCD_2) IsSameAs(rcd interfaces.IRCD) bool {
	return r.String() == rcd.String()
}

// UnmarshalBinary unmarshals the input data into this object
func (r RCD_2) UnmarshalBinary(data []byte) error {
	_, err := r.UnmarshalBinaryData(data)
	return err
}

// CheckSig always returns a hardcoded false
func (r RCD_2) CheckSig(trans interfaces.ITransaction, sigblk interfaces.ISignatureBlock) bool {
	return false
}

// JSONByte returns the json encoded byte array
func (r *RCD_2) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(r)
}

// JSONString returns the json encoded string. TODO: Fix Json marshaling of RCD_2. Right now the RCD type
// is not included in the json marshal.
func (r *RCD_2) JSONString() (string, error) {
	return primitives.EncodeJSONString(r)
}

// String returns this object as a string
func (r RCD_2) String() string {
	txt, err := r.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

// Clone returns an exact copy of this RCD_2
func (r RCD_2) Clone() interfaces.IRCD {
	c := new(RCD_2)
	c.M = r.M
	c.N = r.N
	c.N_Addresses = make([]interfaces.IAddress, len(r.N_Addresses))
	for i, address := range r.N_Addresses {
		c.N_Addresses[i] = CreateAddress(address)
	}
	return c
}

// MarshalBinary marshals this object
func (r RCD_2) MarshalBinary() ([]byte, error) {
	var out primitives.Buffer

	binary.Write(&out, binary.BigEndian, uint8(2))
	binary.Write(&out, binary.BigEndian, uint16(r.N))
	binary.Write(&out, binary.BigEndian, uint16(r.M))
	for i := 0; i < r.M; i++ {
		data, err := r.N_Addresses[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
	}

	return out.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (r *RCD_2) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	if data == nil || len(data) < 5 {
		return nil, fmt.Errorf("Not enough data to unmarshal")
	}
	typ := int8(data[0])
	data = data[1:]
	if typ != 2 {
		return nil, fmt.Errorf("Bad data fed to RCD_2 UnmarshalBinaryData()")
	}

	r.N, data = int(binary.BigEndian.Uint16(data[0:2])), data[2:]
	r.M, data = int(binary.BigEndian.Uint16(data[0:2])), data[2:]
	if r.N > r.M {
		return nil, fmt.Errorf(
			"Error: RCD_2.UnmarshalBinary: signatures possible %d is lower "+
				"than signatures reqired %d",
			r.M, r.N,
		)
	}

	sigLimit := len(data) / 32
	if r.M > sigLimit {
		// TODO: replace this message with a proper error
		return nil, fmt.Errorf(
			"Error: RCD_2.UnmarshalBinary: signatures required %d is larger "+
				"than space in binary %d (uint underflow?)",
			r.M, sigLimit,
		)

	}

	r.N_Addresses = make([]interfaces.IAddress, r.M, r.M)

	for i := range r.N_Addresses {
		r.N_Addresses[i] = new(Address)
		data, err = r.N_Addresses[i].UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// CustomMarshalText marshals this object into a string and returns its bytes
func (r RCD_2) CustomMarshalText() ([]byte, error) {
	var out primitives.Buffer

	primitives.WriteNumber8(&out, uint8(2)) // Type 2 Authorization
	out.WriteString("\n n: ")
	primitives.WriteNumber16(&out, uint16(r.N))
	out.WriteString(" m: ")
	primitives.WriteNumber16(&out, uint16(r.M))
	out.WriteString("\n")
	for i := 0; i < r.M; i++ {
		out.WriteString("  m: ")
		out.WriteString(hex.EncodeToString(r.N_Addresses[i].Bytes()))
		out.WriteString("\n")
	}

	return out.DeepCopyBytes(), nil
}
