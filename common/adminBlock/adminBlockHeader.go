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
	PrevFullHash interfaces.IHash
	DBHeight     uint32

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

func (b *ABlockHeader) GetPrevFullHash() interfaces.IHash {
	return b.PrevFullHash
}

func (b *ABlockHeader) SetDBHeight(dbheight uint32) {
	b.DBHeight = dbheight
}

func (b *ABlockHeader) SetHeaderExpansionArea(area []byte) {
	b.HeaderExpansionArea = area
}

func (b *ABlockHeader) SetPrevFullHash(keyMR interfaces.IHash) {
	b.PrevFullHash = keyMR
}

// Write out the ABlockHeader to binary.
func (b *ABlockHeader) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, err = b.GetAdminChainID().MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = b.PrevFullHash.MarshalBinary()
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

	b.PrevFullHash = new(primitives.Hash)
	newData, err = b.PrevFullHash.UnmarshalBinaryData(newData)
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
