package entryBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// EBlockHeader holds relevant metadata about the Entry Block and the data
// nessisary to verify the previous block in the Entry Block Chain.
type EBlockHeader struct {
	ChainID      interfaces.IHash `json:"chainid"`
	BodyMR       interfaces.IHash `json:"bodymr"`
	PrevKeyMR    interfaces.IHash `json:"prevkeymr"`
	PrevFullHash interfaces.IHash `json:"prevfullhash"`
	EBSequence   uint32           `json:"ebsequence"`
	DBHeight     uint32           `json:"dbheight"`
	EntryCount   uint32           `json:"entrycount"`
}

var _ interfaces.Printable = (*EBlockHeader)(nil)
var _ interfaces.IEntryBlockHeader = (*EBlockHeader)(nil)

func (e *EBlockHeader) Init() {
	if e.ChainID == nil {
		e.ChainID = primitives.NewZeroHash()
	}
	if e.BodyMR == nil {
		e.BodyMR = primitives.NewZeroHash()
	}
	if e.PrevKeyMR == nil {
		e.PrevKeyMR = primitives.NewZeroHash()
	}
	if e.PrevFullHash == nil {
		e.PrevFullHash = primitives.NewZeroHash()
	}
}

func (a *EBlockHeader) IsSameAs(b interfaces.IEntryBlockHeader) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}

	bb, ok := b.(*EBlockHeader)
	if ok == false {
		return false
	}

	if a.ChainID.IsSameAs(bb.ChainID) == false {
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
	if a.EBSequence != bb.EBSequence {
		return false
	}
	if a.DBHeight != bb.DBHeight {
		return false
	}
	if a.EntryCount != bb.EntryCount {
		return false
	}

	return true
}

// NewEBlockHeader initializes a new empty Entry Block Header.
func NewEBlockHeader() *EBlockHeader {
	e := new(EBlockHeader)
	e.Init()
	return e
}

func (e *EBlockHeader) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *EBlockHeader) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *EBlockHeader) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString("  Entry Block Header\n")
	out.WriteString(fmt.Sprintf("    %20s: %x\n", "ChainID", e.ChainID.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("    %20s: %x\n", "BodyMR", e.BodyMR.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("    %20s: %x\n", "PrevKeyMR", e.PrevKeyMR.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("    %20s: %x\n", "PrevFullHash", e.PrevFullHash.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("    %20s: %10v\n", "EBSequence", e.EBSequence))
	out.WriteString(fmt.Sprintf("    %20s: %10v\n", "DBHeight", e.DBHeight))
	out.WriteString(fmt.Sprintf("    %20s: %x\n", "EntryCount", e.EntryCount))
	return (string)(out.DeepCopyBytes())
}

func (c *EBlockHeader) GetChainID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EBlockHeader.GetChainID() saw an interface that was nil")
		}
	}()

	return c.ChainID
}

func (c *EBlockHeader) SetChainID(chainID interfaces.IHash) {
	c.ChainID = chainID
}

func (c *EBlockHeader) GetBodyMR() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EBlockHeader.GetBodyMR() saw an interface that was nil")
		}
	}()

	return c.BodyMR
}

func (c *EBlockHeader) SetBodyMR(bodyMR interfaces.IHash) {
	c.BodyMR = bodyMR
}

func (c *EBlockHeader) GetPrevKeyMR() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EBlockHeader.GetPrevKeyMR() saw an interface that was nil")
		}
	}()

	return c.PrevKeyMR
}

func (c *EBlockHeader) SetPrevKeyMR(prevKeyMR interfaces.IHash) {
	c.PrevKeyMR = prevKeyMR
}

func (c *EBlockHeader) GetPrevFullHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("EBlockHeader.GetPrevFullHash() saw an interface that was nil")
		}
	}()

	return c.PrevFullHash
}

func (c *EBlockHeader) SetPrevFullHash(prevFullHash interfaces.IHash) {
	c.PrevFullHash = prevFullHash
}

func (c *EBlockHeader) GetEBSequence() uint32 {
	return c.EBSequence
}

func (c *EBlockHeader) SetEBSequence(sequence uint32) {
	c.EBSequence = sequence
}

func (c *EBlockHeader) GetDBHeight() uint32 {
	return c.DBHeight
}

func (c *EBlockHeader) SetDBHeight(dbHeight uint32) {
	c.DBHeight = dbHeight
}

func (c *EBlockHeader) GetEntryCount() uint32 {
	return c.EntryCount
}

func (c *EBlockHeader) SetEntryCount(entryCount uint32) {
	c.EntryCount = entryCount
}

// marshalHeaderBinary returns a serialized binary Entry Block Header
func (e *EBlockHeader) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "EBlockHeader.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	buf := primitives.NewBuffer(nil)

	err = buf.PushBinaryMarshallable(e.ChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.BodyMR)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.PrevKeyMR)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.PrevFullHash)
	if err != nil {
		return nil, err
	}

	err = buf.PushUInt32(e.EBSequence)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.DBHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.EntryCount)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// unmarshalHeaderBinary builds the Entry Block Header from the serialized binary.
func (e *EBlockHeader) UnmarshalBinaryData(data []byte) ([]byte, error) {
	e.Init()
	buf := primitives.NewBuffer(data)

	err := buf.PopBinaryMarshallable(e.ChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PopBinaryMarshallable(e.BodyMR)
	if err != nil {
		return nil, err
	}
	err = buf.PopBinaryMarshallable(e.PrevKeyMR)
	if err != nil {
		return nil, err
	}
	err = buf.PopBinaryMarshallable(e.PrevFullHash)
	if err != nil {
		return nil, err
	}

	e.EBSequence, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	e.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	e.EntryCount, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *EBlockHeader) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}
