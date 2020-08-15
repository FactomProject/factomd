// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

// ECBlockHeader contains information related to this EC block as well as the previous EC block
type ECBlockHeader struct {
	BodyHash            interfaces.IHash `json:"bodyhash"`            // The hash of the EC block's body
	PrevHeaderHash      interfaces.IHash `json:"prevheaderhash"`      // The hash of the previous EC block's header
	PrevFullHash        interfaces.IHash `json:"prevfullhash"`        // The full hash of the previous EC block
	DBHeight            uint32           `json:"dbheight"`            // The directory block height this EC block is in
	HeaderExpansionArea []byte           `json:"headerexpansionarea"` // Future expansion area for data
	ObjectCount         uint64           `json:"objectcount"`         // The number of entries in the EC block
	BodySize            uint64           `json:"bodysize"`            // The length of the marshalled body
}

var _ interfaces.Printable = (*ECBlockHeader)(nil)
var _ interfaces.IECBlockHeader = (*ECBlockHeader)(nil)

// Init initializes all nil hashes to the zero hash
func (e *ECBlockHeader) Init() {
	if e.BodyHash == nil {
		e.BodyHash = primitives.NewZeroHash()
	}
	if e.PrevHeaderHash == nil {
		e.PrevHeaderHash = primitives.NewZeroHash()
	}
	if e.PrevFullHash == nil {
		e.PrevFullHash = primitives.NewZeroHash()
	}
	if e.HeaderExpansionArea == nil {
		e.HeaderExpansionArea = make([]byte, 0)
	}
}

// IsSameAs returns true iff the input object is identical to this object
func (e *ECBlockHeader) IsSameAs(b interfaces.IECBlockHeader) bool {
	if e == nil || b == nil {
		if e == nil && b == nil {
			return true
		}
		return false
	}

	bb, ok := b.(*ECBlockHeader)
	if ok == false {
		return false
	}

	if e.BodyHash.IsSameAs(bb.BodyHash) {
		return false
	}
	if e.PrevHeaderHash.IsSameAs(bb.PrevHeaderHash) {
		return false
	}
	if e.PrevFullHash.IsSameAs(bb.PrevFullHash) {
		return false
	}
	if e.DBHeight != bb.DBHeight {
		return false
	}
	if primitives.AreBytesEqual(e.HeaderExpansionArea, bb.HeaderExpansionArea) == false {
		return false
	}
	if e.ObjectCount != bb.ObjectCount {
		return false
	}
	if e.BodySize != bb.BodySize {
		return false
	}

	return true
}

// String returns this object as a string
func (e *ECBlockHeader) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "ECChainID", e.GetECChainID().Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "BodyHash", e.BodyHash.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "PrevHeaderHash", e.PrevHeaderHash.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "PrevFullHash", e.PrevFullHash.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "DBHeight", e.DBHeight))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "HeaderExpansionArea", e.HeaderExpansionArea))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "ObjectCount", e.ObjectCount))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "BodySize", e.BodySize))

	return (string)(out.DeepCopyBytes())
}

// SetBodySize sets the body size to the input
func (e *ECBlockHeader) SetBodySize(cnt uint64) {
	e.BodySize = cnt
}

// GetBodySize returns the body size
func (e *ECBlockHeader) GetBodySize() uint64 {
	return e.BodySize
}

// SetObjectCount sets the object count to the input
func (e *ECBlockHeader) SetObjectCount(cnt uint64) {
	e.ObjectCount = cnt
}

// GetObjectCount returns the object count
func (e *ECBlockHeader) GetObjectCount() uint64 {
	return e.ObjectCount
}

// SetHeaderExpansionArea sets the header expansion area to the input array
func (e *ECBlockHeader) SetHeaderExpansionArea(area []byte) {
	e.HeaderExpansionArea = area
}

// GetHeaderExpansionArea returns the header expansion area array
func (e *ECBlockHeader) GetHeaderExpansionArea() (area []byte) {
	return e.HeaderExpansionArea
}

// SetBodyHash sets the body hash to the input value
func (e *ECBlockHeader) SetBodyHash(hash interfaces.IHash) {
	e.BodyHash = hash
}

// GetBodyHash returns the body hash
func (e *ECBlockHeader) GetBodyHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "ECBlockHeader.GetBodyHash") }()

	return e.BodyHash
}

// GetECChainID returns the EC chain id (see constants.EC_CHAINID)
func (e *ECBlockHeader) GetECChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "ECBlockHeader.GetECChainID") }()

	h := primitives.NewZeroHash()
	h.SetBytes(constants.EC_CHAINID)
	return h
}

// SetPrevHeaderHash sets the previous header hash to the input value
func (e *ECBlockHeader) SetPrevHeaderHash(prev interfaces.IHash) {
	e.PrevHeaderHash = prev
}

// GetPrevHeaderHash returns the previous header hash
func (e *ECBlockHeader) GetPrevHeaderHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "ECBlockHeader.GetPrevHeaderHash") }()

	return e.PrevHeaderHash
}

// SetPrevFullHash sets the previous full hash to the input
func (e *ECBlockHeader) SetPrevFullHash(prev interfaces.IHash) {
	e.PrevFullHash = prev
}

// GetPrevFullHash returns the previous full hash
func (e *ECBlockHeader) GetPrevFullHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "ECBlockHeader.GetPrevFullHash") }()

	return e.PrevFullHash
}

// SetDBHeight sets the directory block height for this EC block to the input value
func (e *ECBlockHeader) SetDBHeight(height uint32) {
	e.DBHeight = height
}

// GetDBHeight returns the directory block height for this EC block
func (e *ECBlockHeader) GetDBHeight() (height uint32) {
	return e.DBHeight
}

// NewECBlockHeader creates a new initialized EC block header
func NewECBlockHeader() *ECBlockHeader {
	h := new(ECBlockHeader)
	h.Init()
	return h
}

// JSONByte returns the json encoded byte array
func (e *ECBlockHeader) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *ECBlockHeader) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

// MarshalBinary marshals the object
func (e *ECBlockHeader) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ECBlockHeader.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	buf := primitives.NewBuffer(nil)

	// 32 byte ECChainID
	err = buf.PushBinaryMarshallable(e.GetECChainID())
	if err != nil {
		return nil, err
	}

	// 32 byte BodyHash
	err = buf.PushBinaryMarshallable(e.GetBodyHash())
	if err != nil {
		return nil, err
	}

	// 32 byte Previous Header Hash
	err = buf.PushBinaryMarshallable(e.GetPrevHeaderHash())
	if err != nil {
		return nil, err
	}

	// 32 byte Previous Full Hash
	err = buf.PushBinaryMarshallable(e.GetPrevFullHash())
	if err != nil {
		return nil, err
	}

	// 4 byte Directory Block Height
	err = buf.PushUInt32(e.GetDBHeight())
	if err != nil {
		return nil, err
	}

	// variable Header Expansion Size
	err = buf.PushVarInt(uint64(len(e.GetHeaderExpansionArea())))
	if err != nil {
		return nil, err
	}

	// varable byte Header Expansion Area
	err = buf.Push(e.GetHeaderExpansionArea())
	if err != nil {
		return nil, err
	}

	// 8 byte Object Count
	err = buf.PushUInt64(e.GetObjectCount())
	if err != nil {
		return nil, err
	}

	// 8 byte size of the Body
	err = buf.PushUInt64(e.GetBodySize())
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (e *ECBlockHeader) UnmarshalBinaryData(data []byte) ([]byte, error) {
	e.Init()
	buf := primitives.NewBuffer(data)

	h := primitives.NewZeroHash()
	err := buf.PopBinaryMarshallable(h)
	if err != nil {
		return nil, err
	}
	if h.String() != constants.EC_CHAINID_STRING {
		return nil, fmt.Errorf("Invalid ChainID - %s", h)
	}

	err = buf.PopBinaryMarshallable(e.BodyHash)
	if err != nil {
		return nil, err
	}
	err = buf.PopBinaryMarshallable(e.PrevHeaderHash)
	if err != nil {
		return nil, err
	}
	err = buf.PopBinaryMarshallable(e.PrevFullHash)
	if err != nil {
		return nil, err
	}

	e.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	// read the Header Expansion Area
	hesize, err := buf.PopVarInt()
	e.HeaderExpansionArea, err = buf.PopLen(int(hesize))
	if err != nil {
		return nil, err
	}

	e.ObjectCount, err = buf.PopUInt64()
	if err != nil {
		return nil, err
	}
	e.BodySize, err = buf.PopUInt64()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this object
func (e *ECBlockHeader) UnmarshalBinary(data []byte) error {
	_, err := e.UnmarshalBinaryData(data)
	return err
}

// ExpandedECBlockHeader is used to help in the function below
type ExpandedECBlockHeader ECBlockHeader

// MarshalJSON marshals the object into an expansed ec block header
func (e ECBlockHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ExpandedECBlockHeader
		ChainID   string `json:"chainid"`
		ECChainID string `json:"ecchainid"`
	}{
		ExpandedECBlockHeader: ExpandedECBlockHeader(e),
		ChainID:               constants.EC_CHAINID_STRING,
		ECChainID:             constants.EC_CHAINID_STRING,
	})
}
