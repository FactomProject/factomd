// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryblock

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)

type DBlockHeader struct {
	Version   byte
	NetworkID uint32

	BodyMR          interfaces.IHash
	PrevKeyMR       interfaces.IHash
	PrevLedgerKeyMR interfaces.IHash

	Timestamp  uint32
	DBHeight   uint32
	BlockCount uint32
}

var _ interfaces.Printable = (*DBlockHeader)(nil)
var _ interfaces.BinaryMarshallable = (*DBlockHeader)(nil)
var _ interfaces.IDirectoryBlockHeader = (*DBlockHeader)(nil)

func (e *DBlockHeader) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *DBlockHeader) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (e *DBlockHeader) JSONBuffer(b *bytes.Buffer) error {
	return EncodeJSONToBuffer(e, b)
}

func (e *DBlockHeader) String() string {
	str, _ := e.JSONString()
	return str
}

func (b *DBlockHeader) EncodableFields() map[string]reflect.Value {
	fields := map[string]reflect.Value{
		`DBHeight`:        reflect.ValueOf(b.DBHeight),
		`BlockCount`:      reflect.ValueOf(b.BlockCount),
		`BodyMR`:          reflect.ValueOf(b.BodyMR),
		`PrevLedgerKeyMR`: reflect.ValueOf(b.PrevLedgerKeyMR),
	}
	return fields
}

func (b *DBlockHeader) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{b.Version})
	binary.Write(&buf, binary.BigEndian, b.NetworkID)

	if b.BodyMR == nil {
		b.BodyMR = new(Hash)
		b.BodyMR.SetBytes(new([32]byte)[:])
	}
	data, err = b.BodyMR.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	data, err = b.PrevKeyMR.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	data, err = b.PrevLedgerKeyMR.MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, b.Timestamp)

	binary.Write(&buf, binary.BigEndian, b.DBHeight)

	binary.Write(&buf, binary.BigEndian, b.BlockCount)

	return buf.Bytes(), err
}

func (b *DBlockHeader) MarshalledSize() uint64 {
	var size uint64 = 0
	size += 1 //Version
	size += 4 //NetworkID
	size += uint64(constants.HASH_LENGTH)
	size += uint64(constants.HASH_LENGTH)
	size += uint64(constants.HASH_LENGTH)
	size += 4 //Timestamp
	size += 4 //DBHeight
	size += 4 //BlockCount

	return size
}

func (b *DBlockHeader) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	b.Version, newData = newData[0], newData[1:]

	b.NetworkID, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	b.BodyMR = new(Hash)
	newData, err = b.BodyMR.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.PrevKeyMR = new(Hash)
	newData, err = b.PrevKeyMR.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.PrevLedgerKeyMR = new(Hash)
	newData, err = b.PrevLedgerKeyMR.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.Timestamp, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	b.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	b.BlockCount, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	return
}

func (b *DBlockHeader) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

/************************************************
 * Support Functions
 ************************************************/

func NewDBlockHeader() *DBlockHeader {
	d := new(DBlockHeader)
	d.BodyMR = NewZeroHash()
	d.PrevKeyMR = NewZeroHash()
	d.PrevLedgerKeyMR = NewZeroHash()

	return d
}
