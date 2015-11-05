// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package adminBlock

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Admin Block Header
type ABlockHeader struct {
	PrevLedgerKeyMR interfaces.IHash
	DBHeight        uint32

	HeaderExpansionSize uint64
	HeaderExpansionArea []byte
	MessageCount        uint32
	BodySize            uint32
}

var _ interfaces.Printable = (*ABlockHeader)(nil)
var _ interfaces.BinaryMarshallable = (*ABlockHeader)(nil)

func (b *ABlockHeader) GetMessageCount() uint32 {
	return b.MessageCount
}

func (b *ABlockHeader) SetMessageCount(messageCount uint32) {
	b.MessageCount = messageCount
}

func (b *ABlockHeader) GetBodySize() uint32 {
	return b.BodySize
}

func (b *ABlockHeader) SetBodySize(bodySize uint32) {
	b.BodySize = bodySize
}

func (b *ABlockHeader) GetAdminChainID() interfaces.IHash {
	return primitives.NewHash(constants.ADMIN_CHAINID)
}

func (b *ABlockHeader) GetDBHeight() uint32 {
	return b.DBHeight
}

func (b *ABlockHeader) GetHeaderExpansionArea() []byte {
	return b.HeaderExpansionArea
}

func (b *ABlockHeader) GetHeaderExpansionSize() uint64 {
	return b.HeaderExpansionSize
}

func (b *ABlockHeader) GetPrevLedgerKeyMR() interfaces.IHash {
	return b.PrevLedgerKeyMR
}

func (b *ABlockHeader) SetDBHeight(dbheight uint32) {
	b.DBHeight = dbheight
}

func (b *ABlockHeader) SetHeaderExpansionArea(area []byte) {
	b.HeaderExpansionArea = area
}

func (b *ABlockHeader) SetPrevLedgerKeyMR(keyMR interfaces.IHash) {
	b.PrevLedgerKeyMR = keyMR
}

// Write out the ABlockHeader to binary.
func (b *ABlockHeader) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, err = b.GetAdminChainID().MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = b.PrevLedgerKeyMR.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, b.DBHeight)

	primitives.EncodeVarInt(&buf, b.HeaderExpansionSize)
	buf.Write(b.HeaderExpansionArea)

	binary.Write(&buf, binary.BigEndian, b.MessageCount)
	binary.Write(&buf, binary.BigEndian, b.BodySize)

	return buf.Bytes(), err
}

func (b *ABlockHeader) MarshalledSize() uint64 {
	var size uint64 = 0

	size += uint64(constants.HASH_LENGTH)                  //AdminChainID
	size += uint64(constants.HASH_LENGTH)                  //PrevFullHash
	size += 4                                              //DBHeight
	size += primitives.VarIntLength(b.HeaderExpansionSize) //HeaderExpansionSize
	size += b.HeaderExpansionSize                          //HeadderExpansionArea
	size += 4                                              //MessageCount
	size += 4                                              //BodySize

	return size
}

func (b *ABlockHeader) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	newData, err = b.GetAdminChainID().UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.PrevLedgerKeyMR = new(primitives.Hash)
	newData, err = b.PrevLedgerKeyMR.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	b.HeaderExpansionSize, newData = primitives.DecodeVarInt(newData)
	b.HeaderExpansionArea, newData = newData[:b.HeaderExpansionSize], newData[b.HeaderExpansionSize:]

	b.MessageCount, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	b.BodySize, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	return
}

// Read in the binary into the ABlockHeader.
func (b *ABlockHeader) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

func (e *ABlockHeader) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ABlockHeader) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *ABlockHeader) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *ABlockHeader) String() string {
	str, _ := e.JSONString()
	return str
}
