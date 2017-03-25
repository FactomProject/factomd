// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type ECBlockHeader struct {
	BodyHash            interfaces.IHash `json:"bodyhash"`
	PrevHeaderHash      interfaces.IHash `json:"prevheaderhash"`
	PrevFullHash        interfaces.IHash `json:"prevfullhash"`
	DBHeight            uint32           `json:"dbheight"`
	HeaderExpansionArea []byte           `json:"headerexpansionarea"`
	ObjectCount         uint64           `json:"objectcount"`
	BodySize            uint64           `json:"bodysize"`
}

var _ = fmt.Print
var _ interfaces.Printable = (*ECBlockHeader)(nil)

func (c *ECBlockHeader) Init() {
	if c.BodyHash == nil {
		c.BodyHash = primitives.NewZeroHash()
	}
	if c.PrevHeaderHash == nil {
		c.PrevHeaderHash = primitives.NewZeroHash()
	}
	if c.PrevFullHash == nil {
		c.PrevFullHash = primitives.NewZeroHash()
	}
	if c.HeaderExpansionArea == nil {
		c.HeaderExpansionArea = make([]byte, 0)
	}
}

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

func (e *ECBlockHeader) SetBodySize(cnt uint64) {
	e.BodySize = cnt
}

func (e *ECBlockHeader) GetBodySize() uint64 {
	return e.BodySize
}

func (e *ECBlockHeader) SetObjectCount(cnt uint64) {
	e.ObjectCount = cnt
}

func (e *ECBlockHeader) GetObjectCount() uint64 {
	return e.ObjectCount
}

func (e *ECBlockHeader) SetHeaderExpansionArea(area []byte) {
	e.HeaderExpansionArea = area
}

func (e *ECBlockHeader) GetHeaderExpansionArea() (area []byte) {
	return e.HeaderExpansionArea
}

func (e *ECBlockHeader) SetBodyHash(prev interfaces.IHash) {
	e.BodyHash = prev
}

func (e *ECBlockHeader) GetBodyHash() interfaces.IHash {
	return e.BodyHash
}

func (e *ECBlockHeader) GetECChainID() interfaces.IHash {
	h := primitives.NewZeroHash()
	h.SetBytes(constants.EC_CHAINID)
	return h
}

func (e *ECBlockHeader) SetPrevHeaderHash(prev interfaces.IHash) {
	e.PrevHeaderHash = prev
}

func (e *ECBlockHeader) GetPrevHeaderHash() interfaces.IHash {
	return e.PrevHeaderHash
}

func (e *ECBlockHeader) SetPrevFullHash(prev interfaces.IHash) {
	e.PrevFullHash = prev
}

func (e *ECBlockHeader) GetPrevFullHash() interfaces.IHash {
	return e.PrevFullHash
}

func (e *ECBlockHeader) SetDBHeight(height uint32) {
	e.DBHeight = height
}

func (e *ECBlockHeader) GetDBHeight() (height uint32) {
	return e.DBHeight
}

func NewECBlockHeader() *ECBlockHeader {
	h := new(ECBlockHeader)
	h.Init()
	return h
}

func (e *ECBlockHeader) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ECBlockHeader) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *ECBlockHeader) MarshalBinary() ([]byte, error) {
	e.Init()
	buf := new(primitives.Buffer)

	// 32 byte ECChainID
	buf.Write(e.GetECChainID().Bytes())

	// 32 byte BodyHash
	buf.Write(e.GetBodyHash().Bytes())

	// 32 byte Previous Header Hash
	buf.Write(e.GetPrevHeaderHash().Bytes())

	// 32 byte Previous Full Hash
	buf.Write(e.GetPrevFullHash().Bytes())

	// 4 byte Directory Block Height
	if err := binary.Write(buf, binary.BigEndian, e.GetDBHeight()); err != nil {
		return nil, err
	}

	// variable Header Expansion Size
	if err := primitives.EncodeVarInt(buf,
		uint64(len(e.GetHeaderExpansionArea()))); err != nil {
		return nil, err
	}

	// varable byte Header Expansion Area
	buf.Write(e.GetHeaderExpansionArea())

	// 8 byte Object Count
	if err := binary.Write(buf, binary.BigEndian, e.GetObjectCount()); err != nil {
		return nil, err
	}

	// 8 byte size of the Body
	if err := binary.Write(buf, binary.BigEndian, e.GetBodySize()); err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *ECBlockHeader) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	buf := primitives.NewBuffer(data)
	hash := make([]byte, 32)

	if _, err = buf.Read(hash); err != nil {
		return
	} else {
		if fmt.Sprintf("%x", hash) != "000000000000000000000000000000000000000000000000000000000000000c" {
			err = fmt.Errorf("Invalid ChainID - %x", hash)
			return
		}
	}

	if _, err = buf.Read(hash); err != nil {
		return
	} else {
		e.BodyHash.SetBytes(hash)
	}

	if _, err = buf.Read(hash); err != nil {
		return
	} else {
		e.PrevHeaderHash.SetBytes(hash)
	}

	if _, err = buf.Read(hash); err != nil {
		return
	} else {
		e.PrevFullHash.SetBytes(hash)
	}

	if err = binary.Read(buf, binary.BigEndian, &e.DBHeight); err != nil {
		return
	}

	// read the Header Expansion Area
	hesize, tmp := primitives.DecodeVarInt(buf.DeepCopyBytes())
	buf = primitives.NewBuffer(tmp)
	e.HeaderExpansionArea = make([]byte, hesize)
	if _, err = buf.Read(e.HeaderExpansionArea); err != nil {
		return
	}

	if err = binary.Read(buf, binary.BigEndian, &e.ObjectCount); err != nil {
		return
	}

	if err = binary.Read(buf, binary.BigEndian, &e.BodySize); err != nil {
		return
	}

	newData = buf.DeepCopyBytes()
	return
}

func (e *ECBlockHeader) UnmarshalBinary(data []byte) error {
	_, err := e.UnmarshalBinaryData(data)
	return err
}

type ExpandedECBlockHeader ECBlockHeader

func (e ECBlockHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ExpandedECBlockHeader
		ChainID   string `json:"chainid"`
		ECChainID string `json:"ecchainid"`
	}{
		ExpandedECBlockHeader: ExpandedECBlockHeader(e),
		ChainID:               "000000000000000000000000000000000000000000000000000000000000000c",
		ECChainID:             "000000000000000000000000000000000000000000000000000000000000000c",
	})
}
