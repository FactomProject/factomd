package entryBlock

import (
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// EBlockHeader holds relevent metadata about the Entry Block and the data
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

func (c *EBlockHeader) GetChainID() interfaces.IHash {
	return c.ChainID
}

func (c *EBlockHeader) SetChainID(chainID interfaces.IHash) {
	c.ChainID = chainID
}

func (c *EBlockHeader) GetBodyMR() interfaces.IHash {
	return c.BodyMR
}

func (c *EBlockHeader) SetBodyMR(bodyMR interfaces.IHash) {
	c.BodyMR = bodyMR
}

func (c *EBlockHeader) GetPrevKeyMR() interfaces.IHash {
	return c.PrevKeyMR
}

func (c *EBlockHeader) SetPrevKeyMR(prevKeyMR interfaces.IHash) {
	c.PrevKeyMR = prevKeyMR
}

func (c *EBlockHeader) GetPrevFullHash() interfaces.IHash {
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
func (e *EBlockHeader) MarshalBinary() ([]byte, error) {
	e.Init()
	buf := new(primitives.Buffer)

	buf.Write(e.ChainID.Bytes())
	buf.Write(e.BodyMR.Bytes())
	buf.Write(e.PrevKeyMR.Bytes())
	buf.Write(e.PrevFullHash.Bytes())

	if err := binary.Write(buf, binary.BigEndian, e.EBSequence); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, e.DBHeight); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, e.EntryCount); err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// unmarshalHeaderBinary builds the Entry Block Header from the serialized binary.
func (e *EBlockHeader) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	e.Init()
	buf := primitives.NewBuffer(data)
	hash := make([]byte, 32)
	newData = data

	if _, err = buf.Read(hash); err != nil {
		return
	} else {
		e.ChainID.SetBytes(hash)
	}

	if _, err = buf.Read(hash); err != nil {
		return
	} else {
		e.BodyMR.SetBytes(hash)
	}

	if _, err = buf.Read(hash); err != nil {
		return
	} else {
		e.PrevKeyMR.SetBytes(hash)
	}

	if _, err = buf.Read(hash); err != nil {
		return
	} else {
		e.PrevFullHash.SetBytes(hash)
	}

	if err = binary.Read(buf, binary.BigEndian, &e.EBSequence); err != nil {
		return
	}

	if err = binary.Read(buf, binary.BigEndian, &e.DBHeight); err != nil {
		return
	}

	if err = binary.Read(buf, binary.BigEndian, &e.EntryCount); err != nil {
		return
	}

	newData = buf.DeepCopyBytes()

	return
}

func (e *EBlockHeader) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}
