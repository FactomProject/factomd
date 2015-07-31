// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// This is a minimum wallet to be used to test the coin
// There isn't much in the way of interest in security
// here, but rather provides a mechanism to create keys
// and sign transactions, etc.

package wallet

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	fct "github.com/FactomProject/factoid"
)

type IWalletEntry interface {
	fct.IBlock
	// Set the RCD for this entry.  USE WITH CAUTION!  You change
	// the hash and thus the address returned by the wallet entry!
	SetRCD(fct.IRCD)
	// Get the RCD used to validate an input
	GetRCD() fct.IRCD
	// Add a public and private key.  USE WITH CAUTION! You change
	// the hash and thus the address returned by the wallet entry!
	AddKey(public, private []byte)
	// Get the name for this address
	GetName() []byte
	// Get the Public Key by its index
	GetKey(i int) []byte
	// Get the Private Key by its index
	GetPrivKey(i int) []byte
	// Set the name for this address
	SetName([]byte)
	// Get the address defined by the RCD for this wallet entry.
	GetAddress() (fct.IAddress, error)
	// Return "ec" for Entry Credit address, and "fct" for a Factoid address
	GetType() string
	SetType(string)
}

type WalletEntry struct {
	IWalletEntry
	// Type string for the address.  Either "ec" or "fct"
	addrtype string
	// 2 byte length not included here
	name []byte
	rcd  fct.IRCD // Verification block for this IWalletEntry
	// 1 byte count of public keys
	public [][]byte // Set of public keys necessary towe sign the rcd
	// 1 byte count of private keys
	private [][]byte // Set of private keys necessary to sign the rcd
}

var _ IWalletEntry = (*WalletEntry)(nil)

func (w WalletEntry) GetType() string {
	return w.addrtype
}

func (w *WalletEntry) SetType(addrtype string) {
	switch addrtype {
	case "ec":
		fallthrough
	case "fct":
		w.addrtype = addrtype
	default:
		panic("Invalid type passed to SetType()")
	}
}

func (b WalletEntry) String() string {
	txt, err := b.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

func (w1 WalletEntry) GetAddress() (fct.IAddress, error) {
	if w1.rcd == nil {
		return nil, fmt.Errorf("Should never happen. Missing the rcd block")
	}
	var adr fct.IHash
	var err error
	if w1.addrtype == "fct" {
		adr, err = w1.rcd.GetAddress()
	} else {
		if len(w1.public) == 0 {
			err = fmt.Errorf("No Public Key for WalletEntry")
		} else {
			adr = fct.NewHash(w1.public[0])
		}
	}
	if err != nil {
		return nil, err
	}
	return adr, nil
}

func (w1 WalletEntry) GetDBHash() fct.IHash {
	return fct.Sha([]byte("WalletEntry"))
}

func (WalletEntry) GetNewInstance() fct.IBlock {
	return new(WalletEntry)
}

func (w1 *WalletEntry) IsEqual(w fct.IBlock) []fct.IBlock {
	w2, ok := w.(*WalletEntry)
	if !ok || w1.GetType() != w2.GetType() {
		r := make([]fct.IBlock, 0, 3)
		return append(r, w1)
	}

	for i, public := range w1.public {
		if bytes.Compare(w2.public[i], public) != 0 {
			r := make([]fct.IBlock, 0, 3)
			return append(r, w1)
		}
	}
	return nil
}

func (w *WalletEntry) UnmarshalBinaryData(data []byte) ([]byte, error) {

	// handle the type byte
	if uint(data[0]) > 1 {
		return nil, fmt.Errorf("Invalid type byte")
	}
	if data[0] == 0 {
		w.addrtype = "fct"
	} else {
		w.addrtype = "ec"
	}
	data = data[1:]

	len, data := binary.BigEndian.Uint16(data[0:2]), data[2:]
	n := make([]byte, len, len) // build a place for the name
	copy(n, data[:len])         // copy it into that place
	data = data[len:]           // update data pointer
	w.name = n                  // Finally!  set the name

	if w.rcd == nil {
		w.rcd = fct.CreateRCD(data) // looks ahead, and creates the right RCD
	}
	data, err := w.rcd.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}

	blen, data := data[0], data[1:]
	w.public = make([][]byte, len, len)
	for i := 0; i < int(blen); i++ {
		w.public[i] = make([]byte, fct.ADDRESS_LENGTH, fct.ADDRESS_LENGTH)
		copy(w.public[i], data[:fct.ADDRESS_LENGTH])
		data = data[fct.ADDRESS_LENGTH:]
	}

	blen, data = data[0], data[1:]
	w.private = make([][]byte, len, len)
	for i := 0; i < int(blen); i++ {
		w.private[i] = make([]byte, fct.PRIVATE_LENGTH, fct.PRIVATE_LENGTH)
		copy(w.private[i], data[:fct.PRIVATE_LENGTH])
		data = data[fct.PRIVATE_LENGTH:]
	}
	return data, nil
}

func (w *WalletEntry) UnmarshalBinary(data []byte) error {
	_, err := w.UnmarshalBinaryData(data)
	return err
}

func (w WalletEntry) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	if w.addrtype == "fct" {
		out.WriteByte(0)
	} else if w.addrtype == "ec" {
		out.WriteByte(1)
	} else {
		panic("Address type not set")
	}

	binary.Write(&out, binary.BigEndian, uint16(len([]byte(w.name))))
	out.Write([]byte(w.name))
	data, err := w.rcd.MarshalBinary()
	if err != nil {
		return nil, err
	}
	out.Write(data)
	out.WriteByte(byte(len(w.public)))
	for _, public := range w.public {
		out.Write(public)
	}
	out.WriteByte(byte(len(w.private)))
	for _, private := range w.private {
		out.Write(private)
	}
	return out.Bytes(), nil
}

func (w WalletEntry) CustomMarshalText() (text []byte, err error) {
	var out bytes.Buffer

	out.WriteString("name:  ")
	out.Write(w.name)
	out.WriteString("\n factoid address:")
	hash, err := w.rcd.GetAddress()
	out.WriteString(hash.String())
	out.WriteString("\n")

	out.WriteString("\n public:  ")
	for i, public := range w.public {
		fct.WriteNumber16(&out, uint16(i))
		out.WriteString(" ")
		addr := hex.EncodeToString(public)
		out.WriteString(addr)
		out.WriteString("\n")
	}

	out.WriteString("\n private:  ")
	for i, private := range w.private {
		fct.WriteNumber16(&out, uint16(i))
		out.WriteString(" ")
		addr := hex.EncodeToString(private)
		out.WriteString(addr)
		out.WriteString("\n")
	}

	return out.Bytes(), nil
}

func (w *WalletEntry) SetRCD(rcd fct.IRCD) {
	w.rcd = rcd
}

func (w WalletEntry) GetRCD() fct.IRCD {
	return w.rcd
}

func (w *WalletEntry) AddKey(public, private []byte) {
	if len(public) != fct.ADDRESS_LENGTH ||
		len(private) != fct.PRIVATE_LENGTH {
		panic("Bad Keys presented to AddKey.  Should not happen.")
	}
	pu := make([]byte, fct.ADDRESS_LENGTH, fct.ADDRESS_LENGTH)
	pr := make([]byte, fct.PRIVATE_LENGTH, fct.PRIVATE_LENGTH)
	copy(pu, public)
	copy(pr, private)
	w.public = append(w.public, pu)
	w.private = append(w.private, pr)

	w.rcd = fct.NewRCD_1(pu)
}

func (we *WalletEntry) GetKey(i int) []byte {
	return we.public[i]
}

func (we *WalletEntry) GetPrivKey(i int) []byte {
	return we.private[i]
}

func (w *WalletEntry) SetName(name []byte) {
	w.name = name
}
