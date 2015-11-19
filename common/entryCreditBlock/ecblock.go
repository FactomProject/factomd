// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"io"
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
	Header interfaces.IECBlockHeader
	Body   interfaces.IECBlockBody
}

var _ interfaces.Printable = (*ECBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*ECBlock)(nil)
var _ interfaces.IEntryCreditBlock = (*ECBlock)(nil)

func (c *ECBlock) GetBody() interfaces.IECBlockBody {
	return c.Body
}

func (c *ECBlock) GetHeader() interfaces.IECBlockHeader {
	return c.Header
}

func (c *ECBlock) New() interfaces.BinaryMarshallableAndCopyable {
	return NewECBlock()
}

func (c *ECBlock) GetDatabaseHeight() uint32 {
	return c.Header.GetDBHeight()
}

func (c *ECBlock) GetChainID() []byte {
	return c.Header.GetECChainID().Bytes()
}

func (c *ECBlock) DatabasePrimaryIndex() interfaces.IHash {
	key, _ := c.HeaderHash()
	return key
}

func (c *ECBlock) DatabaseSecondaryIndex() interfaces.IHash {
	key, _ := c.Hash()
	return key
}

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
		e.GetHeader().SetPrevLedgerKeyMR(primitives.NewHash(constants.ZERO_HASH))
		e.GetHeader().SetDBHeight(1)
	} else {
		v, err := prev.HeaderHash()
		if err != nil {
			return nil, err
		}
		e.GetHeader().SetPrevHeaderHash(v)

		v, err = prev.Hash()
		if err != nil {
			return nil, err
		}
		e.GetHeader().SetPrevLedgerKeyMR(v)

		e.GetHeader().SetDBHeight(prev.GetHeader().GetDBHeight() + 1)
	}
	if err := e.(*ECBlock).BuildHeader(); err != nil {
		return nil, err
	}

	return e, nil
}

func (e *ECBlock) AddEntry(entries ...interfaces.IECBlockEntry) {
	e.GetBody().SetEntries(append(e.GetBody().GetEntries(), entries...))
}

func (e *ECBlock) Hash() (interfaces.IHash, error) {
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
	buf := new(bytes.Buffer)

	// Header
	if err := e.BuildHeader(); err != nil {
		return buf.Bytes(), err
	}
	if p, err := e.GetHeader().MarshalBinary(); err != nil {
		return buf.Bytes(), err
	} else {
		buf.Write(p)
	}

	// Body of ECBlockEntries
	if p, err := e.marshalBodyBinary(); err != nil {
		return buf.Bytes(), err
	} else {
		buf.Write(p)
	}

	return buf.Bytes(), nil
}

func (e *ECBlock) BuildHeader() error {
	// Marshal the Body
	p, err := e.marshalBodyBinary()
	if err != nil {
		return err
	}

	header := e.Header.(*ECBlockHeader)
	header.BodyHash = primitives.Sha(p)
	header.ObjectCount = uint64(len(e.GetBody().GetEntries()))
	header.BodySize = uint64(len(p))

	return nil
}

func (e *ECBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	// Unmarshal Header
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
	buf := new(bytes.Buffer)

	for _, v := range e.GetBody().GetEntries() {
		p, err := v.MarshalBinary()
		if err != nil {
			return buf.Bytes(), err
		}
		buf.WriteByte(v.ECID())
		buf.Write(p)
	}

	return buf.Bytes(), nil
}

func (e *ECBlock) unmarshalBodyBinaryData(data []byte) (newData []byte, err error) {
	buf := bytes.NewBuffer(data)

	for i := uint64(0); i < e.Header.GetObjectCount(); i++ {
		var id byte
		id, err = buf.ReadByte()
		if err != nil {
			newData = buf.Bytes()
			return
		}
		switch id {
		case ECIDServerIndexNumber:
			s := NewServerIndexNumber()
			if buf.Len() < ServerIndexNumberSize {
				err = io.EOF
				newData = buf.Bytes()
				return
			}
			_, err = s.UnmarshalBinaryData(buf.Next(ServerIndexNumberSize))
			if err != nil {
				newData = buf.Bytes()
				return
			}
			e.Body.SetEntries(append(e.Body.GetEntries(), s))
		case ECIDMinuteNumber:
			m := NewMinuteNumber()
			if buf.Len() < MinuteNumberSize {
				err = io.EOF
				newData = buf.Bytes()
				return
			}
			_, err = m.UnmarshalBinaryData(buf.Next(MinuteNumberSize))
			if err != nil {
				newData = buf.Bytes()
				return
			}
			e.Body.SetEntries(append(e.Body.GetEntries(), m))
		case ECIDChainCommit:
			if buf.Len() < CommitChainSize {
				err = io.EOF
				newData = buf.Bytes()
				return
			}
			c := NewCommitChain()
			_, err = c.UnmarshalBinaryData(buf.Next(CommitChainSize))
			if err != nil {
				return
			}
			e.Body.SetEntries(append(e.Body.GetEntries(), c))
		case ECIDEntryCommit:
			if buf.Len() < CommitEntrySize {
				err = io.EOF
				newData = buf.Bytes()
				return
			}
			c := NewCommitEntry()
			_, err = c.UnmarshalBinaryData(buf.Next(CommitEntrySize))
			if err != nil {
				return
			}
			e.Body.SetEntries(append(e.Body.GetEntries(), c))
		case ECIDBalanceIncrease:
			c := NewIncreaseBalance()
			tmp, err := c.UnmarshalBinaryData(buf.Bytes())
			if err != nil {
				return tmp, err
			}
			e.Body.SetEntries(append(e.Body.GetEntries(), c))
			buf = bytes.NewBuffer(tmp)
		default:
			err = fmt.Errorf("Unsupported ECID: %x\n", id)
			return
		}
	}

	newData = buf.Bytes()
	return
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

func (e *ECBlock) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *ECBlock) String() string {
	str, _ := e.JSONString()
	return str
}

type ECBlockBody struct {
	Entries []interfaces.IECBlockEntry
}

var _ interfaces.Printable = (*ECBlockBody)(nil)
var _ interfaces.IECBlockBody = (*ECBlockBody)(nil)

func NewECBlockBody() interfaces.IECBlockBody {
	b := new(ECBlockBody)
	b.Entries = make([]interfaces.IECBlockEntry, 0)
	return b
}

func (e *ECBlockBody) GetEntries() []interfaces.IECBlockEntry {
	return e.Entries
}

func (e *ECBlockBody) SetEntries(entries []interfaces.IECBlockEntry) {
	e.Entries = entries
}

func (e *ECBlockBody) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ECBlockBody) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *ECBlockBody) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *ECBlockBody) String() string {
	str, _ := e.JSONString()
	return str
}
