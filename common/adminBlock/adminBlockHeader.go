// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package adminBlock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"errors"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Admin Block Header
type ABlockHeader struct {
	PrevBackRefHash interfaces.IHash `json:"prevbackrefhash"`
	DBHeight        uint32           `json:"dbheight"`

	HeaderExpansionSize uint64 `json:"headerexpansionsize"`
	HeaderExpansionArea []byte `json:"headerexpansionarea"`
	MessageCount        uint32 `json:"messagecount"`
	BodySize            uint32 `json:"bodysize"`
}

var _ interfaces.Printable = (*ABlockHeader)(nil)
var _ interfaces.BinaryMarshallable = (*ABlockHeader)(nil)

func (e *ABlockHeader) IsSameAs(e2 interfaces.IABlockHeader) bool {
	if !e.PrevBackRefHash.IsSameAs(e2.GetPrevBackRefHash()) {
		return false
	}
	if e.DBHeight != e2.GetDBHeight() {
		return false
	}
	if int(e.HeaderExpansionSize) != len(e2.GetHeaderExpansionArea()) {
		return false
	}
	if e.MessageCount != e2.GetMessageCount() {
		return false
	}
	if e.BodySize != e2.GetBodySize() {
		return false
	}
	if bytes.Compare(e.HeaderExpansionArea, e2.GetHeaderExpansionArea()) != 0 {
		return false
	}
	return true
}

func (e *ABlockHeader) Init() {
	if e.PrevBackRefHash == nil {
		e.PrevBackRefHash = primitives.NewZeroHash()
	}
	if e.HeaderExpansionSize == 0 {
		e.HeaderExpansionArea = make([]byte, 0)
	}
}

func (e *ABlockHeader) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString("  Admin Block Header\n")
	out.WriteString(fmt.Sprintf("    %20s: %10v\n", "PrevBackRefHash", e.PrevBackRefHash.String()))
	out.WriteString(fmt.Sprintf("    %20s: %10v\n", "DBHeight", e.DBHeight))
	out.WriteString(fmt.Sprintf("    %20s: %10v\n", "HeaderExpansionSize", e.HeaderExpansionSize))
	out.WriteString(fmt.Sprintf("    %20s: %x\n", "HeaderExpansionArea", e.HeaderExpansionArea))
	out.WriteString(fmt.Sprintf("    %20s: %x\n", "MessageCount", e.MessageCount))
	out.WriteString(fmt.Sprintf("    %20s: %x\n", "BodySize", e.BodySize))
	return (string)(out.DeepCopyBytes())
}

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

func (b *ABlockHeader) GetPrevBackRefHash() interfaces.IHash {
	b.Init()
	return b.PrevBackRefHash
}

func (b *ABlockHeader) SetDBHeight(dbheight uint32) {
	b.DBHeight = dbheight
}

func (b *ABlockHeader) SetHeaderExpansionArea(area []byte) {
	b.HeaderExpansionArea = area
	b.HeaderExpansionSize = uint64(len(area))
}

func (b *ABlockHeader) SetPrevBackRefHash(BackRefHash interfaces.IHash) {
	b.PrevBackRefHash = BackRefHash
}

// Write out the ABlockHeader to binary.
func (b *ABlockHeader) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ABlockHeader.MarshalBinary err:%v", *pe)
		}
	}(&err)
	b.Init()
	var buf primitives.Buffer

	err = buf.PushBinaryMarshallable(b.GetAdminChainID())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(b.PrevBackRefHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(b.DBHeight)
	if err != nil {
		return nil, err
	}

	err = buf.PushVarInt(b.HeaderExpansionSize)
	if err != nil {
		return nil, err
	}
	err = buf.Push(b.HeaderExpansionArea)
	if err != nil {
		return nil, err
	}

	err = buf.PushUInt32(b.MessageCount)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(b.BodySize)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), err
}

func (b *ABlockHeader) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	h := primitives.NewZeroHash()
	err := buf.PopBinaryMarshallable(h)
	if err != nil {
		return nil, err
	}
	if h.String() != "000000000000000000000000000000000000000000000000000000000000000a" {
		return nil, errors.New("Block does not begin with the ABlock ChainID")
	}

	b.PrevBackRefHash = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(b.PrevBackRefHash)
	if err != nil {
		return nil, err
	}

	b.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	b.HeaderExpansionSize, err = buf.PopVarInt()
	if err != nil {
		return nil, err
	}
	if b.HeaderExpansionSize > 0 {
		b.HeaderExpansionArea, err = buf.PopLen(int(b.HeaderExpansionSize))
		if err != nil {
			return nil, err
		}
	}

	b.MessageCount, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	b.BodySize, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
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

type ExpandedABlockHeader ABlockHeader

func (e ABlockHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ExpandedABlockHeader
		AdminChainID string `json:"adminchainid"`
		ChainID      string `json:"chainid"`
	}{
		ExpandedABlockHeader: ExpandedABlockHeader(e),
		AdminChainID:         "000000000000000000000000000000000000000000000000000000000000000a",
		ChainID:              "000000000000000000000000000000000000000000000000000000000000000a",
	})
}
