package entryBlock

import (
	"fmt"
	"os"

	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

// EBlockHeader holds relevant metadata about the Entry Block and the data
// necessary to verify the previous block in the Entry Block Chain.
type EBlockHeader struct {
	ChainID      interfaces.IHash `json:"chainid"`      // The chain id associated with this entry block's entries (hash array)
	BodyMR       interfaces.IHash `json:"bodymr"`       // The Merkle root of the Entry block's entries (hash array)
	PrevKeyMR    interfaces.IHash `json:"prevkeymr"`    // The Merkle root of the previous entry block for this chain id
	PrevFullHash interfaces.IHash `json:"prevfullhash"` // The full hash of the previous entry block for this chain id
	EBSequence   uint32           `json:"ebsequence"`   // Entry block sequence number: ie 7 = the seventh entry block for this chain id
	DBHeight     uint32           `json:"dbheight"`     // The directory block height this entry block is located in
	EntryCount   uint32           `json:"entrycount"`   // How many entries are in the hash array for this entry block
}

var _ interfaces.Printable = (*EBlockHeader)(nil)
var _ interfaces.IEntryBlockHeader = (*EBlockHeader)(nil)

// Init initializes the objects hashes to the zero hash if they are nil
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

// IsSameAs returns true iff the input object is the same as this object
func (e *EBlockHeader) IsSameAs(b interfaces.IEntryBlockHeader) bool {
	if e == nil || b == nil {
		if e == nil && b == nil {
			return true
		}
		return false
	}

	bb, ok := b.(*EBlockHeader)
	if ok == false {
		return false
	}

	if e.ChainID.IsSameAs(bb.ChainID) == false {
		return false
	}
	if e.BodyMR.IsSameAs(bb.BodyMR) == false {
		return false
	}
	if e.PrevKeyMR.IsSameAs(bb.PrevKeyMR) == false {
		return false
	}
	if e.PrevFullHash.IsSameAs(bb.PrevFullHash) == false {
		return false
	}
	if e.EBSequence != bb.EBSequence {
		return false
	}
	if e.DBHeight != bb.DBHeight {
		return false
	}
	if e.EntryCount != bb.EntryCount {
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

// JSONByte returns the json encoded byte array
func (e *EBlockHeader) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *EBlockHeader) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

// String returns this object as a string
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

// GetChainID returns the chain id of this entry block
func (e *EBlockHeader) GetChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "EBlockHeader.GetChainID") }()

	return e.ChainID
}

// SetChainID sets the chain id associated with this entry block's entries
func (e *EBlockHeader) SetChainID(chainID interfaces.IHash) {
	e.ChainID = chainID
}

// GetBodyMR returns the Merkle root of the entry blocks body (the hash array)
func (e *EBlockHeader) GetBodyMR() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "EBlockHeader.GetBodyMR") }()

	return e.BodyMR
}

// SetBodyMR sets the body Merkle root to th einput value
func (e *EBlockHeader) SetBodyMR(bodyMR interfaces.IHash) {
	e.BodyMR = bodyMR
}

// GetPrevKeyMR return the previous entry blocks Merkle root
func (e *EBlockHeader) GetPrevKeyMR() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "EBlockHeader.GetPrevKeyMR") }()

	return e.PrevKeyMR
}

// SetPrevKeyMR sets the previous entry block key Merkle root to the input value
func (e *EBlockHeader) SetPrevKeyMR(prevKeyMR interfaces.IHash) {
	e.PrevKeyMR = prevKeyMR
}

// GetPrevFullHash returns the previous entry blocks full hash associated with this chain id
func (e *EBlockHeader) GetPrevFullHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "EBlockHeader.GetPrevFullHash") }()

	return e.PrevFullHash
}

// SetPrevFullHash sets the previous entry blocks full hash associated with this chain id
func (e *EBlockHeader) SetPrevFullHash(prevFullHash interfaces.IHash) {
	e.PrevFullHash = prevFullHash
}

// GetEBSequence returns this entry block's sequence number
func (e *EBlockHeader) GetEBSequence() uint32 {
	return e.EBSequence
}

// SetEBSequence sets this entry block's sequence number
func (e *EBlockHeader) SetEBSequence(sequence uint32) {
	e.EBSequence = sequence
}

// GetDBHeight returns the directory block height this entry block is from
func (e *EBlockHeader) GetDBHeight() uint32 {
	return e.DBHeight
}

// SetDBHeight sets the directory block height this entry block is from
func (e *EBlockHeader) SetDBHeight(dbHeight uint32) {
	e.DBHeight = dbHeight
}

// GetEntryCount returns the number of entries in this entry block
func (e *EBlockHeader) GetEntryCount() uint32 {
	return e.EntryCount
}

// SetEntryCount sets the number of entries in this entry block
func (e *EBlockHeader) SetEntryCount(entryCount uint32) {
	e.EntryCount = entryCount
}

// MarshalBinary returns a serialized binary Entry Block Header
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

// UnmarshalBinaryData builds the Entry Block Header from the serialized binary.
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

// UnmarshalBinary builds the Entry Block Header from the serialized binary.
func (e *EBlockHeader) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}
