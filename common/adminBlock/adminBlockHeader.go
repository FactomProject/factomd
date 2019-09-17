// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package adminBlock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"errors"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// ABlockHeader contains header information for the Admin Block
type ABlockHeader struct {
	PrevBackRefHash     interfaces.IHash `json:"prevbackrefhash"`
	DBHeight            uint32           `json:"dbheight"` // The directory block height where this admin block is located
	HeaderExpansionSize uint64           `json:"headerexpansionsize"`
	HeaderExpansionArea []byte           `json:"headerexpansionarea"`
	MessageCount        uint32           `json:"messagecount"`
	BodySize            uint32           `json:"bodysize"`
}

var _ interfaces.Printable = (*ABlockHeader)(nil)
var _ interfaces.BinaryMarshallable = (*ABlockHeader)(nil)

// IsSameAs returns true iff the input object is identical to this object
func (b *ABlockHeader) IsSameAs(b2 interfaces.IABlockHeader) bool {
	if !b.PrevBackRefHash.IsSameAs(b2.GetPrevBackRefHash()) {
		return false
	}
	if b.DBHeight != b2.GetDBHeight() {
		return false
	}
	if int(b.HeaderExpansionSize) != len(b2.GetHeaderExpansionArea()) {
		return false
	}
	if b.MessageCount != b2.GetMessageCount() {
		return false
	}
	if b.BodySize != b2.GetBodySize() {
		return false
	}
	if bytes.Compare(b.HeaderExpansionArea, b2.GetHeaderExpansionArea()) != 0 {
		return false
	}
	return true
}

// Init initializes all nil hashes to the zero hash, and makes all arrays
func (b *ABlockHeader) Init() {
	if b.PrevBackRefHash == nil {
		b.PrevBackRefHash = primitives.NewZeroHash()
	}
	if b.HeaderExpansionSize == 0 {
		b.HeaderExpansionArea = make([]byte, 0)
	}
}

// String returns this objects string
func (b *ABlockHeader) String() string {
	b.Init()
	var out primitives.Buffer
	out.WriteString("  Admin Block Header\n")
	out.WriteString(fmt.Sprintf("    %20s: %10v\n", "PrevBackRefHash", b.PrevBackRefHash.String()))
	out.WriteString(fmt.Sprintf("    %20s: %10v\n", "DBHeight", b.DBHeight))
	out.WriteString(fmt.Sprintf("    %20s: %10v\n", "HeaderExpansionSize", b.HeaderExpansionSize))
	out.WriteString(fmt.Sprintf("    %20s: %x\n", "HeaderExpansionArea", b.HeaderExpansionArea))
	out.WriteString(fmt.Sprintf("    %20s: %x\n", "MessageCount", b.MessageCount))
	out.WriteString(fmt.Sprintf("    %20s: %x\n", "BodySize", b.BodySize))
	return (string)(out.DeepCopyBytes())
}

// GetMessageCount returns the current message count
func (b *ABlockHeader) GetMessageCount() uint32 {
	return b.MessageCount
}

// SetMessageCount sets the message count to the incoming value
func (b *ABlockHeader) SetMessageCount(messageCount uint32) {
	b.MessageCount = messageCount
}

// GetBodySize returns the body size
func (b *ABlockHeader) GetBodySize() uint32 {
	return b.BodySize
}

// SetBodySize sets the body size to the incoming value
func (b *ABlockHeader) SetBodySize(bodySize uint32) {
	b.BodySize = bodySize
}

// GetAdminChainID returns the admin chain id 0x0a
func (b *ABlockHeader) GetAdminChainID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("ABlockHeader.GetAdminChainID() saw an interface that was nil")
		}
	}()
	return primitives.NewHash(constants.ADMIN_CHAINID)
}

// GetDBHeight returns the directory block height this admin block header is associated with
func (b *ABlockHeader) GetDBHeight() uint32 {
	return b.DBHeight
}

// GetHeaderExpansionArea returns the header expansion area
func (b *ABlockHeader) GetHeaderExpansionArea() []byte {
	return b.HeaderExpansionArea
}

// GetHeaderExpansionSize returns the header expansion size
func (b *ABlockHeader) GetHeaderExpansionSize() uint64 {
	return b.HeaderExpansionSize
}

// GetPrevBackRefHash returns the previous back reference hash
func (b *ABlockHeader) GetPrevBackRefHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("ABlockHeader.GetPrevBackRefHash() saw an interface that was nil")
		}
	}()
	b.Init()
	return b.PrevBackRefHash
}

// SetDBHeight sets the directory block height
func (b *ABlockHeader) SetDBHeight(dbheight uint32) {
	b.DBHeight = dbheight
}

// SetHeaderExpansionArea sets the header expansion area and size based on the input byte array
func (b *ABlockHeader) SetHeaderExpansionArea(area []byte) {
	b.HeaderExpansionArea = area
	b.HeaderExpansionSize = uint64(len(area))
}

// SetPrevBackRefHash sets the previous back reference hash
func (b *ABlockHeader) SetPrevBackRefHash(BackRefHash interfaces.IHash) {
	b.PrevBackRefHash = BackRefHash
}

// MarshalBinary marshals the object
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

// UnmarshalBinaryData unmarshals the input data to this object
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

// UnmarshalBinary unmarshals the input data into this object
func (b *ABlockHeader) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (b *ABlockHeader) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(b)
}

// JSONString returns the json encoded string
func (b *ABlockHeader) JSONString() (string, error) {
	return primitives.EncodeJSONString(b)
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
