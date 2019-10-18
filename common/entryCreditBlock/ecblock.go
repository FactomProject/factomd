// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// ECBlock is an Entry Credit Block and consists of a header and a body. The body is composed
// of primarily Commits and Balance Increases with Minute Markers and Server
// Markers distributed throughout.
type ECBlock struct {
	Header interfaces.IECBlockHeader `json:"header"` // The entry credit block header
	Body   interfaces.IECBlockBody   `json:"body"`   // The entry credit block body
}

var _ interfaces.Printable = (*ECBlock)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*ECBlock)(nil)
var _ interfaces.IEntryCreditBlock = (*ECBlock)(nil)
var _ interfaces.DatabaseBlockWithEntries = (*ECBlock)(nil)

// Init initializes the header and body to new objects if they are nil
func (e *ECBlock) Init() {
	if e.Header == nil {
		h := new(ECBlockHeader)
		h.Init()
		e.Header = h
	}
	if e.Body == nil {
		e.Body = new(ECBlockBody)
	}
}

// IsSameAs returns true iff the input block is the same as this object
func (e *ECBlock) IsSameAs(b interfaces.IEntryCreditBlock) bool {
	if e == nil || b == nil {
		if e == nil && b == nil {
			return true
		}
		return false
	}

	if e.Header.IsSameAs(b.GetHeader()) == false {
		return false
	}
	if e.Body.IsSameAs(b.GetBody()) == false {
		return false
	}

	return true
}

// UpdateState executes the EC transactions
func (e *ECBlock) UpdateState(state interfaces.IState) error {
	if state == nil {
		return fmt.Errorf("No State provided")
	}
	e.Init()
	state.UpdateECs(e)
	return nil
}

// String returns this object as a string (header and body included)
func (e *ECBlock) String() string {
	str := e.GetHeader().String()
	str = str + e.GetBody().String()
	return str
}

// GetEntries returns the entries in this block
func (e *ECBlock) GetEntries() []interfaces.IECBlockEntry {
	e.Init()
	return e.GetBody().GetEntries()
}

// GetEntryByHash returns the entry whose hash matches the input hash, or whose signature hash matches the input hash.
// If no hash is found, returns nil
func (e *ECBlock) GetEntryByHash(hash interfaces.IHash) interfaces.IECBlockEntry {
	if hash == nil {
		return nil
	}

	txs := e.GetEntries()
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

// GetEntryHashes returns a list of hashes for each entry in the ECBlock that is either a balance increase, chain commit, or entry commit
func (e *ECBlock) GetEntryHashes() []interfaces.IHash {
	entries := e.GetBody().GetEntries()
	answer := make([]interfaces.IHash, 0, len(entries))
	for _, entry := range entries {
		if entry.ECID() == constants.ECIDBalanceIncrease ||
			entry.ECID() == constants.ECIDChainCommit ||
			entry.ECID() == constants.ECIDEntryCommit {
			answer = append(answer, entry.Hash())
		}
	}
	return answer
}

// GetEntrySigHashes returns a list of signature hashes for each entry in the ECBlock that is either a balance increase, chain commit,
// or entry commit
func (e *ECBlock) GetEntrySigHashes() []interfaces.IHash {
	entries := e.GetBody().GetEntries()
	answer := make([]interfaces.IHash, 0, len(entries))
	for _, entry := range entries {
		if entry.ECID() == constants.ECIDBalanceIncrease ||
			entry.ECID() == constants.ECIDChainCommit ||
			entry.ECID() == constants.ECIDEntryCommit {
			sHash := entry.GetSigHash()
			if sHash != nil {
				answer = append(answer, sHash)
			}
		}
	}
	return answer
}

// GetBody returns the EC block body
func (e *ECBlock) GetBody() interfaces.IECBlockBody {
	e.Init()
	return e.Body
}

// GetHeader returns the EC block header
func (e *ECBlock) GetHeader() interfaces.IECBlockHeader {
	e.Init()
	return e.Header
}

// New returns a new EC block
func (e *ECBlock) New() interfaces.BinaryMarshallableAndCopyable {
	block, _ := NextECBlock(nil)
	return block
}

// GetDatabaseHeight returns the directory block height this EC block is part of
func (e *ECBlock) GetDatabaseHeight() uint32 {
	return e.GetHeader().GetDBHeight()
}

// GetChainID returns the chain id of this EC block
func (e *ECBlock) GetChainID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("ECBlock.GetChainID() saw an interface that was nil")
		}
	}()

	return e.GetHeader().GetECChainID()
}

// DatabasePrimaryIndex returns the hash of the header
func (e *ECBlock) DatabasePrimaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("ECBlock.DatabasePrimaryIndex() saw an interface that was nil")
		}
	}()

	key, _ := e.HeaderHash()
	return key
}

// DatabaseSecondaryIndex returns the full hash (header and body) of the object
func (e *ECBlock) DatabaseSecondaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("ECBlock.DatabaseSecondaryIndex() saw an interface that was nil")
		}
	}()

	key, _ := e.GetFullHash()
	return key
}

// AddEntry appends the input entries into the EC block
func (e *ECBlock) AddEntry(entries ...interfaces.IECBlockEntry) {
	e.GetBody().SetEntries(append(e.GetBody().GetEntries(), entries...))
}

// GetHash returns the full hash (header and body) of the object
func (e *ECBlock) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("ECBlock.GetHash() saw an interface that was nil")
		}
	}()

	h, _ := e.GetFullHash()
	return h
}

// GetFullHash returns the full hash (header and body) of the object
func (e *ECBlock) GetFullHash() (interfaces.IHash, error) {
	p, err := e.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return primitives.Sha(p), nil
}

// HeaderHash returns the hash of the header
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

// MarshalBinary marshals the object (header and body)
func (e *ECBlock) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ECBlock.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	buf := primitives.NewBuffer(nil)

	// Header
	err = e.BuildHeader()
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(e.GetHeader())
	if err != nil {
		return nil, err
	}
	x := buf.DeepCopyBytes()
	buf = primitives.NewBuffer(x)

	// Body of ECBlockEntries
	p, err := e.marshalBodyBinary()
	if err != nil {
		return nil, err
	}
	err = buf.Push(p)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// BuildHeader sets all relevant header values based on the current body contents
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

// UnmarshalECBlock unmarshals the input data into a new EC block
func UnmarshalECBlock(data []byte) (interfaces.IEntryCreditBlock, error) {
	block, _ := NextECBlock(nil)

	err := block.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// UnmarshalBinaryData unmarshals the input data into this EC block
func (e *ECBlock) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	// Unmarshal Header
	if e.GetHeader() == nil {
		e.Header = NewECBlockHeader()
	}
	err := buf.PopBinaryMarshallable(e.GetHeader())
	if err != nil {
		return nil, err
	}

	// Unmarshal Body
	newData, err := e.unmarshalBodyBinaryData(buf.DeepCopyBytes())
	if err != nil {
		return nil, err
	}

	return newData, err
}

// UnmarshalBinary unmarshals the input data into this EC block
func (e *ECBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// marshalBodyBinary marshals only the body of the EC block
func (e *ECBlock) marshalBodyBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "ECBlock.marshalBodyBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	buf := primitives.NewBuffer(nil)
	entries := e.GetBody().GetEntries()

	for _, v := range entries {
		err := buf.PushByte(v.ECID())
		if err != nil {
			return nil, err
		}
		err = buf.PushBinaryMarshallable(v)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

// unmarshalBodyBinaryData unmarshals the input data into the body (no header info)
func (e *ECBlock) unmarshalBodyBinaryData(data []byte) ([]byte, error) {
	var err error
	e.Init()
	// No longer use buffers due to the lots of copying of the same data

	//buf := primitives.NewBuffer(data)
	newData := data

	allentries := make([]interfaces.IECBlockEntry, e.GetHeader().GetObjectCount())
	for i := uint64(0); i < e.GetHeader().GetObjectCount(); i++ {
		id := newData[0]
		newData = newData[1:]

		switch id {
		case constants.ECIDServerIndexNumber:
			s := NewServerIndexNumber()
			newData, err = s.UnmarshalBinaryData(newData)
			if err != nil {
				return nil, err
			}
			allentries[i] = s
		case constants.ECIDMinuteNumber:
			m := NewMinuteNumber(0)
			_, err = m.UnmarshalBinaryData(newData[:1])
			if err != nil {
				return nil, err
			}
			allentries[i] = m
			newData = newData[1:]
		case constants.ECIDChainCommit:
			c := NewCommitChain()
			_, err = c.UnmarshalBinaryData(newData[:200])
			if err != nil {
				return nil, err
			}
			allentries[i] = c
			newData = newData[200:]
		case constants.ECIDEntryCommit:
			c := NewCommitEntry()
			_, err = c.UnmarshalBinaryData(newData[:136])
			if err != nil {
				return nil, err
			}
			allentries[i] = c
			newData = newData[136:]
		case constants.ECIDBalanceIncrease:
			c := NewIncreaseBalance()
			newData, err = c.UnmarshalBinaryData(newData)
			if err != nil {
				return nil, err
			}
			allentries[i] = c
		default:
			err = fmt.Errorf("Unsupported ECID: %x\n", id)
			return nil, err
		}
	}

	e.Body.SetEntries(allentries)

	//buf.Reset()
	//buf.Write(newData)
	return newData, nil
}

// unmarshalBodyBinary unmarshals the input data into the body (no header info)
func (e *ECBlock) unmarshalBodyBinary(data []byte) (err error) {
	_, err = e.unmarshalBodyBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *ECBlock) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *ECBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

/********************************************************
 * Support Functions
 ********************************************************/

// NewECBlock creates a new empty EC block
func NewECBlock() interfaces.IEntryCreditBlock {
	e := new(ECBlock)
	e.Header = NewECBlockHeader()
	e.Body = NewECBlockBody()
	return e
}

// NextECBlock creates a new EC block with header information filled in from the input previous block information
func NextECBlock(prev interfaces.IEntryCreditBlock) (interfaces.IEntryCreditBlock, error) {
	e := NewECBlock()

	// Handle the really unusual case of the first block.
	if prev == nil {
		e.GetHeader().SetPrevHeaderHash(primitives.NewZeroHash())
		e.GetHeader().SetPrevFullHash(primitives.NewZeroHash())
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

// CheckBlockPairIntegrity checks that the input block is derived from the previous block via their header information
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
