// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Structure for managing Addresses.  Addresses can be literally the public
// key for holding some value, requiring a signature to release that value.
// Or they can be a Hash of an Authentication block.  In which case, if the
// the authentication block is valid, the value is released (and we can
// prove this is okay, because the hash of the authentication block must
// match this address.

package database

import (
	fct "github.com/FactomProject/factoid"
	"fmt"
	"encoding/binary"
	"bytes"
)

var _ = fmt.Println

type IByteStore interface {
	fct.IBlock
	Bytes() []byte
	SetBytes([]byte)
}

type ByteStore struct {
	byteData [] byte
}

var _ IByteStore = (*ByteStore)(nil)

func (b ByteStore) Bytes() []byte {
	return b.byteData
}
func (b ByteStore) GetHash() fct.IHash {
	return fct.Sha(b.byteData)
}

func (b *ByteStore) SetBytes(data []byte) {
	b.byteData = data
}


func (b ByteStore) String() string {
	return string(b.byteData)
}

func (ByteStore) GetDBHash() fct.IHash {
	return fct.Sha([]byte("ByteStore"))
}

func (b ByteStore)CustomMarshalText() ([]byte, error) {
	return b.byteData, nil
}

// We need the progress through the slice, so we really can't use the stock spec
// for the UnmarshalBinary() method from encode.  We define our own method that
// makes the code easier to read and way more efficent.
func (b *ByteStore)UnmarshalBinaryData(data []byte) ([]byte, error) {
	size, data := binary.BigEndian.Uint32(data), data[4:]
	b.byteData = make([]byte, size,size)
	copy(b.byteData,data[:size])
	return data[size:], nil
}

func (b *ByteStore)UnmarshalBinary(data []byte) error {
	_,err := b.UnmarshalBinaryData(data)
	return err
}

func (b ByteStore)MarshalBinary() ([]byte, error){
	var out bytes.Buffer
	binary.Write(&out, binary.BigEndian, uint32(len(b.byteData)))
	out.Write(b.byteData)
	return out.Bytes(), nil
}

func (b1 ByteStore)IsEqual(b fct.IBlock) []fct.IBlock {
	b2, ok := b.(*ByteStore)
	if !ok || !bytes.Equal(b1.byteData,b2.byteData){
		return []fct.IBlock{&b1} 
	}
	return nil
}

func (ByteStore) GetNewInstance() fct.IBlock {
	return new(ByteStore)
}