// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"fmt"
	"io"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const (
	ECIDServerIndexNumber byte = iota
	ECIDMinuteNumber
	ECIDChainCommit
	ECIDEntryCommit
	ECIDBalanceIncrease
)

// The Entry Credit Block consists of a header and a body. The body is composed
// of primarily Commits and Balance Increases with Minute Markers and Server
// Markers distributed throughout.
type ECBlock struct {
	Header interfaces.IECBlockHeader `json:"header"`
	Body   interfaces.IECBlockBody   `json:"body"`
}

var _ interfaces.Printable = (*ECBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*ECBlock)(nil)
var _ interfaces.IEntryCreditBlock = (*ECBlock)(nil)
var _ interfaces.DatabaseBlockWithEntries = (*ECBlock)(nil)

func (c *ECBlock) Init() {
	if c.Header == nil {
		h := new(ECBlockHeader)
		h.Init()
		c.Header = h
	}
	if c.Body == nil {
		c.Body = new(ECBlockBody)
	}
}

func (c *ECBlock) UpdateState(state interfaces.IState) error {
	if state == nil {
		return fmt.Errorf("No State provided")
	}
	c.Init()
	state.UpdateECs(c)
	return nil
}

func (c *ECBlock) String() string {
	str := c.GetHeader().String()
	str = str + c.GetBody().String()
	return str
}

func (c *ECBlock) GetEntries() []interfaces.IECBlockEntry {
	c.Init()
	return c.GetBody().GetEntries()
}

func (c *ECBlock) GetEntryByHash(hash interfaces.IHash) interfaces.IECBlockEntry {
	if hash == nil {
		return nil
	}

	txs := c.GetEntries()
	for _, tx := range txs {
		if hash.IsSameAs(tx.Hash()) {
			return tx
		}
		if hash.IsSameAs(tx.GetSigHash()) {
			return tx
		}
	}
	return nil
}

func (c *ECBlock) GetEntryHashes() []interfaces.IHash {
	entries := c.GetBody().GetEntries()
	answer := make([]interfaces.IHash, 0, len(entries))
	for _, entry := range entries {
		if entry.ECID() == ECIDBalanceIncrease || entry.ECID() == ECIDChainCommit || entry.ECID() == ECIDEntryCommit {
			answer = append(answer, entry.Hash())
		}
	}
	return answer
}

func (c *ECBlock) GetEntrySigHashes() []interfaces.IHash {
	entries := c.GetBody().GetEntries()
	answer := make([]interfaces.IHash, 0, len(entries))
	for _, entry := range entries {
		if entry.ECID() == ECIDBalanceIncrease || entry.ECID() == ECIDChainCommit || entry.ECID() == ECIDEntryCommit {
			sHash := entry.GetSigHash()
			if sHash != nil {
				answer = append(answer, sHash)
			}
		}
	}
	return answer
}

func (c *ECBlock) GetBody() interfaces.IECBlockBody {
	c.Init()
	return c.Body
}

func (c *ECBlock) GetHeader() interfaces.IECBlockHeader {
	c.Init()
	return c.Header
}

func (c *ECBlock) New() interfaces.BinaryMarshallableAndCopyable {
	block, _ := NextECBlock(nil)
	return block
}

func (c *ECBlock) GetDatabaseHeight() uint32 {
	return c.GetHeader().GetDBHeight()
}

func (c *ECBlock) GetChainID() interfaces.IHash {
	return c.GetHeader().GetECChainID()
}

func (c *ECBlock) DatabasePrimaryIndex() interfaces.IHash {
	key, _ := c.HeaderHash()
	return key
}

func (c *ECBlock) DatabaseSecondaryIndex() interfaces.IHash {
	key, _ := c.GetFullHash()
	return key
}

func (e *ECBlock) AddEntry(entries ...interfaces.IECBlockEntry) {
	e.GetBody().SetEntries(append(e.GetBody().GetEntries(), entries...))
}

func (e *ECBlock) GetHash() interfaces.IHash {
	h, _ := e.GetFullHash()
	return h
}

// This is the FullHash.
func (e *ECBlock) GetFullHash() (interfaces.IHash, error) {
	p, err := e.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return primitives.Sha(p), nil
}

func (e *ECBlock) HeaderHash() (interfaces.IHash, error) {
	if err := e.BuildHeader(); err != nil {
		return nil, err
	}

	p, err := e.GetHeader().MarshalBinary()
	if err != nil {
		return nil, err
	}
	return primitives.Sha(p), nil
}

func (e *ECBlock) MarshalBinary() ([]byte, error) {
	e.Init()
	buf := new(primitives.Buffer)

	// Header
	if err := e.BuildHeader(); err != nil {
		return nil, err
	}
	if p, err := e.GetHeader().MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(p)
	}

	// Body of ECBlockEntries
	if p, err := e.marshalBodyBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(p)
	}

	return buf.DeepCopyBytes(), nil
}

func (e *ECBlock) BuildHeader() error {
	e.Init()
	// Marshal the Body
	p, err := e.marshalBodyBinary()
	if err != nil {
		return err
	}

	header := e.GetHeader().(*ECBlockHeader)
	header.BodyHash = primitives.Sha(p)
	header.ObjectCount = uint64(len(e.GetBody().GetEntries()))
	header.BodySize = uint64(len(p))

	return nil
}

func UnmarshalECBlock(data []byte) (interfaces.IEntryCreditBlock, error) {
	block, _ := NextECBlock(nil)

	err := block.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (e *ECBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	// Unmarshal Header
	if e.GetHeader() == nil {
		e.Header = NewECBlockHeader()
	}
	newData, err = e.GetHeader().UnmarshalBinaryData(data)
	if err != nil {
		return
	}

	// Unmarshal Body
	newData, err = e.unmarshalBodyBinaryData(newData)
	if err != nil {
		return
	}

	return
}

func (e *ECBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *ECBlock) marshalBodyBinary() ([]byte, error) {
	e.Init()
	buf := new(primitives.Buffer)

	for _, v := range e.GetBody().GetEntries() {
		p, err := v.MarshalBinary()
		if err != nil {
			return buf.Bytes(), err
		}
		buf.WriteByte(v.ECID())
		buf.Write(p)
	}

	return buf.DeepCopyBytes(), nil
}

func (e *ECBlock) unmarshalBodyBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	for i := uint64(0); i < e.GetHeader().GetObjectCount(); i++ {
		var id byte
		id, err = buf.ReadByte()
		if err != nil {
			return nil, err
		}
		switch id {
		case ECIDServerIndexNumber:
			s := NewServerIndexNumber()
			if buf.Len() < ServerIndexNumberSize {
				err = io.EOF
				return nil, err
			}
			_, err = s.UnmarshalBinaryData(buf.Next(ServerIndexNumberSize))
			if err != nil {
				return nil, err
			}
			e.GetBody().AddEntry(s)
		case ECIDMinuteNumber:
			m := NewMinuteNumber(0)
			if buf.Len() < MinuteNumberSize {
				err = io.EOF
				return nil, err
			}
			_, err = m.UnmarshalBinaryData(buf.Next(MinuteNumberSize))
			if err != nil {
				return nil, err
			}
			e.GetBody().AddEntry(m)
		case ECIDChainCommit:
			if buf.Len() < CommitChainSize {
				err = io.EOF
				return nil, err
			}
			c := NewCommitChain()
			_, err = c.UnmarshalBinaryData(buf.Next(CommitChainSize))
			if err != nil {
				return nil, err
			}
			e.GetBody().AddEntry(c)
		case ECIDEntryCommit:
			if buf.Len() < CommitEntrySize {
				err = io.EOF
				return nil, err
			}
			c := NewCommitEntry()
			_, err = c.UnmarshalBinaryData(buf.Next(CommitEntrySize))
			if err != nil {
				return nil, err
			}
			e.GetBody().AddEntry(c)
		case ECIDBalanceIncrease:
			c := NewIncreaseBalance()
			tmp, err := c.UnmarshalBinaryData(buf.DeepCopyBytes())
			if err != nil {
				return nil, err
			}
			e.GetBody().AddEntry(c)
			buf = primitives.NewBuffer(tmp)
		default:
			err = fmt.Errorf("Unsupported ECID: %x\n", id)
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (b *ECBlock) unmarshalBodyBinary(data []byte) (err error) {
	_, err = b.unmarshalBodyBinaryData(data)
	return
}

func (e *ECBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ECBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

/********************************************************
 * Support Functions
 ********************************************************/

func NewECBlock() interfaces.IEntryCreditBlock {
	e := new(ECBlock)
	e.Header = NewECBlockHeader()
	e.Body = NewECBlockBody()
	return e
}

func NextECBlock(prev interfaces.IEntryCreditBlock) (interfaces.IEntryCreditBlock, error) {
	e := NewECBlock()

	// Handle the really unusual case of the first block.
	if prev == nil {
		e.GetHeader().SetPrevHeaderHash(primitives.NewHash(constants.ZERO_HASH))
		e.GetHeader().SetPrevFullHash(primitives.NewHash(constants.ZERO_HASH))
		e.GetHeader().SetDBHeight(0)
	} else {
		v, err := prev.HeaderHash()
		if err != nil {
			return nil, err
		}
		e.GetHeader().SetPrevHeaderHash(v)

		v, err = prev.GetFullHash()
		if err != nil {
			return nil, err
		}
		e.GetHeader().SetPrevFullHash(v)

		e.GetHeader().SetDBHeight(prev.GetHeader().GetDBHeight() + 1)
	}
	if err := e.(*ECBlock).BuildHeader(); err != nil {
		return nil, err
	}

	return e, nil
}

func CheckBlockPairIntegrity(block interfaces.IEntryCreditBlock, prev interfaces.IEntryCreditBlock) error {
	if block == nil {
		return fmt.Errorf("No block specified")
	}

	if prev == nil {
		if block.GetHeader().GetPrevHeaderHash().IsZero() == false {
			return fmt.Errorf("Invalid PrevHeaderHash")
		}
		if block.GetHeader().GetPrevFullHash().IsZero() == false {
			return fmt.Errorf("Invalid PrevFullHash")
		}
		if block.GetHeader().GetDBHeight() != 0 {
			return fmt.Errorf("Invalid DBHeight")
		}
	} else {
		h, err := prev.HeaderHash()
		if err != nil {
			return err
		}
		if block.GetHeader().GetPrevHeaderHash().IsSameAs(h) == false {
			return fmt.Errorf("Invalid PrevHeaderHash")
		}
		h, err = prev.GetFullHash()
		if err != nil {
			return err
		}
		if block.GetHeader().GetPrevFullHash().IsSameAs(h) == false {
			return fmt.Errorf("Invalid PrevFullHash")
		}
		if block.GetHeader().GetDBHeight() != (prev.GetHeader().GetDBHeight() + 1) {
			return fmt.Errorf("Invalid DBHeight")
		}
	}

	return nil
}
