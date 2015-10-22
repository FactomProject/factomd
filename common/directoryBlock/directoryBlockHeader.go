// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryblock

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type DBlockHeader struct {
	version   byte
	networkID uint32

	bodyMR          interfaces.IHash
	prevKeyMR       interfaces.IHash
	prevLedgerKeyMR interfaces.IHash

	timestamp  uint32
	dbHeight   uint32
	blockCount uint32
}

var _ interfaces.Printable = (*DBlockHeader)(nil)
var _ interfaces.BinaryMarshallable = (*DBlockHeader)(nil)
var _ interfaces.IDirectoryBlockHeader = (*DBlockHeader)(nil)

func (h *DBlockHeader) Version()   byte {
	return h.version
}

func (h *DBlockHeader) SetVersion(version byte)   {
	h.version = version
}

func (h *DBlockHeader) NetworkID() uint32 {
	return h.networkID
}

func (h *DBlockHeader) SetNetworkID(networkID uint32)   {
	h.networkID = networkID
}

func (h *DBlockHeader) BodyMR()          interfaces.IHash {
	return h.bodyMR
}

func (h *DBlockHeader) SetBodyMR(bodyMR interfaces.IHash) {
	h.bodyMR = bodyMR
}

func (h *DBlockHeader) PrevKeyMR()       interfaces.IHash {
	return h.prevKeyMR
}

func (h *DBlockHeader) SetPrevKeyMR(prevKeyMR interfaces.IHash) {
	h.prevKeyMR = prevKeyMR
}

func (h *DBlockHeader) PrevLedgerKeyMR() interfaces.IHash {
	return h.prevLedgerKeyMR
}

func (h *DBlockHeader) SetPrevLedgerKeyMR(PrevLedgerKeyMR interfaces.IHash) {
	h.prevLedgerKeyMR = PrevLedgerKeyMR
}

func (h *DBlockHeader) Timestamp()  uint32 {
	return h.timestamp
}

func (h *DBlockHeader) SetTimestamp(timestamp uint32) {
	h.timestamp = timestamp
}

func (h *DBlockHeader) DBHeight()   uint32 {
	return h.dbHeight
}

func (h *DBlockHeader) SetDBHeight(dbheight uint32) {
	h.dbHeight = dbheight
}


func (h *DBlockHeader) BlockCount() uint32 {
	return h.blockCount
}

func (h *DBlockHeader) SetBlockCount(blockcount uint32) {
	h.blockCount = blockcount
}


func (e *DBlockHeader) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DBlockHeader) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *DBlockHeader) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *DBlockHeader) String() string {
	str, _ := e.JSONString()
	return str
}

func (b *DBlockHeader) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.WriteByte(b.version)
	binary.Write(&buf, binary.BigEndian, b.NetworkID())
	
	if b.BodyMR == nil {
		b.SetBodyMR(new(primitives.Hash))
		b.BodyMR().SetBytes(new([32]byte)[:])
	}
	data, err = b.BodyMR().MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	data, err = b.PrevKeyMR().MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	data, err = b.PrevLedgerKeyMR().MarshalBinary()
	if err != nil {
		return
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, b.Timestamp())

	binary.Write(&buf, binary.BigEndian, b.DBHeight())

	binary.Write(&buf, binary.BigEndian, b.BlockCount())

	return buf.Bytes(), err
}

func (b *DBlockHeader) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	b.version, newData = newData[0], newData[1:]

	b.networkID, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	b.bodyMR = new(primitives.Hash)
	newData, err = b.BodyMR().UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.prevKeyMR = new(primitives.Hash)
	newData, err = b.PrevKeyMR().UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.prevLedgerKeyMR = new(primitives.Hash)
	newData, err = b.PrevLedgerKeyMR().UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.timestamp, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	b.dbHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	b.blockCount, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

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
	d.bodyMR = primitives.NewZeroHash()
	d.prevKeyMR = primitives.NewZeroHash()
	d.prevLedgerKeyMR = primitives.NewZeroHash()

	return d
}
