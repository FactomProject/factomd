// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

type DBlockHeader struct {
	Version   byte
	NetworkID uint32

	BodyMR       interfaces.IHash
	PrevKeyMR    interfaces.IHash
	PrevFullHash interfaces.IHash

	Timestamp  uint32 //in minutes
	DBHeight   uint32
	BlockCount uint32
}

var _ interfaces.Printable = (*DBlockHeader)(nil)
var _ interfaces.BinaryMarshallable = (*DBlockHeader)(nil)
var _ interfaces.IDirectoryBlockHeader = (*DBlockHeader)(nil)

func (h *DBlockHeader) GetVersion() byte {
	return h.Version
}

func (h *DBlockHeader) SetVersion(version byte) {
	h.Version = version
}

func (h *DBlockHeader) GetNetworkID() uint32 {
	return h.NetworkID
}

func (h *DBlockHeader) SetNetworkID(networkID uint32) {
	h.NetworkID = networkID
}

func (h *DBlockHeader) GetBodyMR() interfaces.IHash {
	return h.BodyMR
}

func (h *DBlockHeader) SetBodyMR(bodyMR interfaces.IHash) {
	h.BodyMR = bodyMR
}

func (h *DBlockHeader) GetPrevKeyMR() interfaces.IHash {
	return h.PrevKeyMR
}

func (h *DBlockHeader) SetPrevKeyMR(prevKeyMR interfaces.IHash) {
	h.PrevKeyMR = prevKeyMR
}

func (h *DBlockHeader) GetPrevFullHash() interfaces.IHash {
	return h.PrevFullHash
}

func (h *DBlockHeader) SetPrevFullHash(PrevFullHash interfaces.IHash) {
	h.PrevFullHash = PrevFullHash
}

func (h *DBlockHeader) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampFromMinutes(h.Timestamp)
}

func (h *DBlockHeader) SetTimestamp(timestamp interfaces.Timestamp) {
	h.Timestamp = timestamp.GetTimeMinutesUInt32()
}

func (h *DBlockHeader) GetDBHeight() uint32 {
	return h.DBHeight
}

func (h *DBlockHeader) SetDBHeight(dbheight uint32) {
	h.DBHeight = dbheight
}

func (h *DBlockHeader) GetBlockCount() uint32 {
	return h.BlockCount
}

func (h *DBlockHeader) SetBlockCount(blockcount uint32) {
	h.BlockCount = blockcount
}

func (e *DBlockHeader) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DBlockHeader) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *DBlockHeader) String() string {
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("  Version:         %v\n", e.Version))
	out.WriteString(fmt.Sprintf("  NetworkID:       %x\n", e.NetworkID))
	out.WriteString(fmt.Sprintf("  BodyMR:          %s\n", e.BodyMR.String()))
	out.WriteString(fmt.Sprintf("  PrevKeyMR:       %s\n", e.PrevKeyMR.String()))
	out.WriteString(fmt.Sprintf("  PrevFullHash:    %s\n", e.PrevFullHash.String()))
	out.WriteString(fmt.Sprintf("  Timestamp:       %d\n", e.Timestamp))
	out.WriteString(fmt.Sprintf("  Timestamp Str:   %s\n", e.GetTimestamp().String()))
	out.WriteString(fmt.Sprintf("  DBHeight:        %d\n", e.DBHeight))
	out.WriteString(fmt.Sprintf("  BlockCount:      %d\n", e.BlockCount))

	return (string)(out.DeepCopyBytes())
}

func (b *DBlockHeader) MarshalBinary() ([]byte, error) {
	var buf primitives.Buffer

	buf.WriteByte(b.Version)
	binary.Write(&buf, binary.BigEndian, b.NetworkID)

	data, err := b.BodyMR.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = b.PrevKeyMR.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = b.PrevFullHash.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, b.Timestamp)

	binary.Write(&buf, binary.BigEndian, b.DBHeight)

	binary.Write(&buf, binary.BigEndian, b.BlockCount)

	if b.BlockCount > 100000 {
		panic("Send: Blockcount too great in directory block")
	}

	return buf.DeepCopyBytes(), err
}

func (b *DBlockHeader) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Directory Block Header: %v", r)
		}
	}()

	//	fmt.Printf("Unmarshal %x\n",data[:113])

	newData = data
	b.Version, newData = newData[0], newData[1:]

	b.NetworkID, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	b.BodyMR = new(primitives.Hash)
	newData, err = b.BodyMR.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.PrevKeyMR = new(primitives.Hash)
	newData, err = b.PrevKeyMR.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.PrevFullHash = new(primitives.Hash)
	newData, err = b.PrevFullHash.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.Timestamp, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	b.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	b.BlockCount, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	if b.BlockCount > 100000 {
		panic("Receive: Blockcount too great in directory block" + fmt.Sprintf(":::: %d", b.BlockCount))
	}
	return
}

func (b *DBlockHeader) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

type ExpandedDBlockHeader DBlockHeader

func (e DBlockHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ExpandedDBlockHeader
		ChainID string
	}{
		ExpandedDBlockHeader: ExpandedDBlockHeader(e),
		ChainID:              "000000000000000000000000000000000000000000000000000000000000000d",
	})
}

/************************************************
 * Support Functions
 ************************************************/

func NewDBlockHeader() *DBlockHeader {
	d := new(DBlockHeader)
	d.BodyMR = primitives.NewZeroHash()
	d.PrevKeyMR = primitives.NewZeroHash()
	d.PrevFullHash = primitives.NewZeroHash()

	return d
}
