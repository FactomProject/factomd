// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

type DBlockHeader struct {
	Version   byte   `json:"version"`
	NetworkID uint32 `json:"networkid"`

	BodyMR       interfaces.IHash `json:"bodymr"`
	PrevKeyMR    interfaces.IHash `json:"prevkeymr"`
	PrevFullHash interfaces.IHash `json:"prevfullhash"`

	Timestamp  uint32 `json:"timestamp"` //in minutes
	DBHeight   uint32 `json:"dbheight"`
	BlockCount uint32 `json:"blockcount"`
}

var _ interfaces.Printable = (*DBlockHeader)(nil)
var _ interfaces.BinaryMarshallable = (*DBlockHeader)(nil)
var _ interfaces.IDirectoryBlockHeader = (*DBlockHeader)(nil)

func (h *DBlockHeader) Init() {
	if h.BodyMR == nil {
		h.BodyMR = primitives.NewZeroHash()
	}
	if h.PrevKeyMR == nil {
		h.PrevKeyMR = primitives.NewZeroHash()
	}
	if h.PrevFullHash == nil {
		h.PrevFullHash = primitives.NewZeroHash()
	}
}

func (b *DBlockHeader) GetHeaderHash() (interfaces.IHash, error) {

	binaryEBHeader, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}

	h := primitives.Sha(binaryEBHeader)

	return h, nil
}

func (a *DBlockHeader) IsSameAs(b interfaces.IDirectoryBlockHeader) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}

	bb, ok := b.(*DBlockHeader)
	if ok == false {
		return false
	}

	if a.Version != bb.Version {
		return false
	}
	if a.NetworkID != bb.NetworkID {
		return false
	}

	if a.BodyMR.IsSameAs(bb.BodyMR) == false {
		return false
	}
	if a.PrevKeyMR.IsSameAs(bb.PrevKeyMR) == false {
		return false
	}
	if a.PrevFullHash.IsSameAs(bb.PrevFullHash) == false {
		return false
	}

	if a.Timestamp != bb.Timestamp {
		return false
	}
	if a.DBHeight != bb.DBHeight {
		return false
	}
	if a.BlockCount != bb.BlockCount {
		return false
	}

	return true
}

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
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("  version:         %v\n", e.Version))
	out.WriteString(fmt.Sprintf("  networkid:       %x\n", e.NetworkID))
	out.WriteString(fmt.Sprintf("  bodymr:          %s\n", e.BodyMR.String()))
	out.WriteString(fmt.Sprintf("  prevkeymr:       %s\n", e.PrevKeyMR.String()))
	out.WriteString(fmt.Sprintf("  prevfullhash:    %s\n", e.PrevFullHash.String()))
	out.WriteString(fmt.Sprintf("  timestamp:       %d\n", e.Timestamp))
	out.WriteString(fmt.Sprintf("  timestamp str:   %s\n", e.GetTimestamp().String()))
	out.WriteString(fmt.Sprintf("  dbheight:        %d\n", e.DBHeight))
	out.WriteString(fmt.Sprintf("  blockcount:      %d\n", e.BlockCount))

	return (string)(out.DeepCopyBytes())
}

func (b *DBlockHeader) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DBlockHeader.MarshalBinary err:%v", *pe)
		}
	}(&err)
	b.Init()
	buf := primitives.NewBuffer(nil)

	err = buf.PushByte(b.Version)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(b.NetworkID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(b.BodyMR)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(b.PrevKeyMR)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(b.PrevFullHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(b.Timestamp)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(b.DBHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(b.BlockCount)
	if err != nil {
		return nil, err
	}

	if b.BlockCount > 100000 {
		panic("Send: Blockcount too great in directory block")
	}

	return buf.DeepCopyBytes(), err
}

func (b *DBlockHeader) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	var err error

	b.Version, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	b.NetworkID, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	b.BodyMR = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(b.BodyMR)
	if err != nil {
		return nil, err
	}
	b.PrevKeyMR = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(b.PrevKeyMR)
	if err != nil {
		return nil, err
	}
	b.PrevFullHash = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(b.PrevFullHash)
	if err != nil {
		return nil, err
	}

	b.Timestamp, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	b.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	b.BlockCount, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	if b.BlockCount > 100000 {
		panic("Receive: Blockcount too great in directory block" + fmt.Sprintf(":::: %d", b.BlockCount))
	}

	return buf.DeepCopyBytes(), nil
}

func (b *DBlockHeader) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

type ExpandedDBlockHeader DBlockHeader

func (e DBlockHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ExpandedDBlockHeader
		ChainID string `json:"chainid"`
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
