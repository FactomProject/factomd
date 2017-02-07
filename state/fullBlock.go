package state

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
)

// Can take a directory block and package all the data into a file to be torrented.
// Can also unpack packages

type WholeBlock struct {
	// Header and entry headers
	DBlock interfaces.IDirectoryBlock

	// Block and admin block entries
	ABlock interfaces.IAdminBlock

	// Block and transactions
	FBlock interfaces.IFBlock

	// Block and entries
	ECBlock interfaces.IEntryCreditBlock

	EBlocks []interfaces.IEntryBlock

	Entries []interfaces.IEBEntry

	SigList []interfaces.IFullSignature
}

func NewWholeBlock() *WholeBlock {
	w := new(WholeBlock)
	w.EBlocks = make([]interfaces.IEntryBlock, 0)
	w.Entries = make([]interfaces.IEBEntry, 0)
	w.SigList = make([]interfaces.IFullSignature, 0)

	return w
}

func (wb *WholeBlock) BlockToDBStateMsg() interfaces.IMsg {
	ts := primitives.NewTimestampNow()
	m := messages.NewDBStateMsg(ts,
		wb.DBlock,
		wb.ABlock,
		wb.FBlock,
		wb.ECBlock,
		wb.EBlocks,
		wb.Entries,
		wb.SigList)

	return m
}

func (wb *WholeBlock) AddEblock(eb interfaces.IEntryBlock) {
	wb.EBlocks = append(wb.EBlocks, eb)
}

func (wb *WholeBlock) AddEntry(e interfaces.IEntry) {
	wb.Entries = append(wb.Entries, e)
}

func (wb *WholeBlock) AddIEBEntry(e interfaces.IEBEntry) {
	wb.Entries = append(wb.Entries, e)
}

func (a *WholeBlock) IsSameAs(b *WholeBlock) (resp bool) {
	/*defer func() {
		if r := recover(); r != nil {
			resp = false
			return
		}
	}()*/

	if !a.DBlock.GetHash().IsSameAs(b.DBlock.GetHash()) {
		return false
	}

	if !a.ABlock.GetHash().IsSameAs(b.ABlock.GetHash()) {
		return false
	}

	if !a.FBlock.GetHash().IsSameAs(b.FBlock.GetHash()) {
		return false
	}

	if !a.ECBlock.GetHash().IsSameAs(b.ECBlock.GetHash()) {
		return false
	}

	if len(a.EBlocks) != len(b.EBlocks) {
		return false
	}

	for i, _ := range a.EBlocks {
		if !a.EBlocks[i].GetHash().IsSameAs(b.EBlocks[i].GetHash()) {
			return false
		}
	}

	if len(a.Entries) != len(b.Entries) {
		return false
	}

	for i, _ := range a.Entries {
		if !a.Entries[i].GetHash().IsSameAs(b.Entries[i].GetHash()) {
			return false
		}
	}

	if len(a.SigList) != len(b.SigList) {
		return false
	}

	for i, _ := range a.SigList {
		if !a.SigList[i].IsSameAs(b.SigList[i]) {
			return false
		}
	}

	return true
}

func (wb *WholeBlock) MarshalBinary() (b []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("A panic has occurred while marshaling: %s", r)
			return
		}
	}()

	buf := new(bytes.Buffer)

	data, err := marshalHelper(wb.DBlock)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(data)
	if err != nil {
		return nil, err
	}

	data, err = marshalHelper(wb.ABlock)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(data)
	if err != nil {
		return nil, err
	}

	data, err = marshalHelper(wb.FBlock)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(data)
	if err != nil {
		return nil, err
	}

	data, err = marshalHelper(wb.ECBlock)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(data)
	if err != nil {
		return nil, err
	}

	data, err = Uint32ToBytes(uint32(len(wb.EBlocks)))
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(data)
	if err != nil {
		return nil, err
	}

	var i uint32
	for i = 0; i < uint32(len(wb.EBlocks)); i++ {
		data, err = marshalHelper(wb.EBlocks[i])
		if err != nil {
			return nil, err
		}
		_, err = buf.Write(data)
		if err != nil {
			return nil, err
		}
	}

	data, err = Uint32ToBytes(uint32(len(wb.Entries)))
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(data)
	if err != nil {
		return nil, err
	}

	for i = 0; i < uint32(len(wb.Entries)); i++ {
		data, err = marshalHelper(wb.Entries[i])
		if err != nil {
			return nil, err
		}
		_, err = buf.Write(data)
		if err != nil {
			return nil, err
		}
	}

	data, err = Uint32ToBytes(uint32(len(wb.SigList)))
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(data)
	if err != nil {
		return nil, err
	}

	for i = 0; i < uint32(len(wb.SigList)); i++ {
		data, err = marshalHelper(wb.SigList[i])
		if err != nil {
			return nil, err
		}
		_, err = buf.Write(data)
		if err != nil {
			return nil, err
		}
	}
	return buf.Next(buf.Len()), nil
}

func (wb *WholeBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = wb.UnmarshalBinaryData(data)
	return
}

func (wb *WholeBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("A panic has occurred while unmarshaling: %s", r)
			return
		}
	}()

	newData = data

	d := new(directoryBlock.DirectoryBlock)
	newData, err = unmarshalHelper(d, newData)
	if err != nil {
		return
	}
	wb.DBlock = d

	a := new(adminBlock.AdminBlock)
	newData, err = unmarshalHelper(a, newData)
	if err != nil {
		return
	}
	wb.ABlock = a

	f := new(factoid.FBlock)
	newData, err = unmarshalHelper(f, newData)
	if err != nil {
		return
	}
	wb.FBlock = f

	ec := entryCreditBlock.NewECBlock()
	newData, err = unmarshalHelper(ec, newData)
	if err != nil {
		return
	}
	wb.ECBlock = ec

	// Eblocks
	u, err := BytesToUint32(newData[:4])
	if err != nil {
		return
	}
	newData = newData[4:]
	wb.EBlocks = make([]interfaces.IEntryBlock, u)

	for i := range wb.EBlocks {
		eb := entryBlock.NewEBlock()

		newData, err = unmarshalHelper(eb, newData)
		if err != nil {
			return
		}
		wb.EBlocks[i] = eb
	}

	// Entries
	u, err = BytesToUint32(newData[:4])
	if err != nil {
		return
	}
	newData = newData[4:]
	wb.Entries = make([]interfaces.IEBEntry, u)

	for i := range wb.Entries {
		e := entryBlock.NewEntry()
		newData, err = unmarshalHelper(e, newData)
		if err != nil {
			return
		}
		wb.Entries[i] = e
	}

	// Signatures
	u, err = BytesToUint32(newData[:4])
	if err != nil {
		return
	}
	newData = newData[4:]
	wb.SigList = make([]interfaces.IFullSignature, u)

	for i := range wb.Entries {
		s := new(primitives.Signature)
		newData, err = unmarshalHelper(s, newData)
		if err != nil {
			return
		}
		wb.SigList[i] = s
	}

	return
}

// unmarshalHelper prepends it's []byte length for unmarshaler
func unmarshalHelper(obj interfaces.BinaryMarshallable, data []byte) (newData []byte, err error) {
	newData = data
	u, err := BytesToUint32(newData[:4])
	if err != nil {
		return nil, err
	}
	newData = newData[4:]

	marData := newData[:u]
	err = obj.UnmarshalBinary(marData)
	newData = newData[u:]
	return
}

// marshalHelper will marshal the obj, and prepend it's length so we don't overmarshal
func marshalHelper(obj interfaces.BinaryMarshallable) ([]byte, error) {
	data, err := obj.MarshalBinary()
	if err != nil {
		return nil, err
	}

	length, err := Uint32ToBytes(uint32(len(data)))
	if err != nil {
		return nil, err
	}

	res := append(length, data...)
	return res, nil
}

func Uint32ToBytes(val uint32) ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, val)

	return b, nil
}

func BytesToUint32(data []byte) (ret uint32, err error) {
	buf := bytes.NewBuffer(data)
	err = binary.Read(buf, binary.LittleEndian, &ret)
	return
}
