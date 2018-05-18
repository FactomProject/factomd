// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"os"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

type MissingEntryBlock struct {
	EBHash   interfaces.IHash
	DBHeight uint32
}

var _ interfaces.BinaryMarshallable = (*MissingEntryBlock)(nil)

func RandomMissingEntryBlock() *MissingEntryBlock {
	meb := new(MissingEntryBlock)
	meb.EBHash = primitives.RandomHash()
	meb.DBHeight = random.RandUInt32()
	return meb
}

func (s *MissingEntryBlock) IsSameAs(b *MissingEntryBlock) bool {
	if s.EBHash.IsSameAs(b.EBHash) == false {
		return false
	}
	return s.DBHeight == b.DBHeight
}

func (s *MissingEntryBlock) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "MissingEntryBlock.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)

	err = buf.PushBinaryMarshallable(s.EBHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(s.DBHeight)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (s *MissingEntryBlock) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	s.EBHash = primitives.NewZeroHash()

	newData = p
	buf := primitives.NewBuffer(p)

	err = buf.PopBinaryMarshallable(s.EBHash)
	if err != nil {
		return
	}

	s.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return
	}

	newData = buf.DeepCopyBytes()
	return
}

func (s *MissingEntryBlock) UnmarshalBinary(p []byte) error {
	_, err := s.UnmarshalBinaryData(p)
	return err
}

type MissingEntry struct {
	Cnt       int
	LastTime  time.Time
	EBHash    interfaces.IHash
	EntryHash interfaces.IHash
	DBHeight  uint32
}

var _ interfaces.BinaryMarshallable = (*MissingEntry)(nil)

func RandomMissingEntry() *MissingEntry {
	me := new(MissingEntry)
	me.Cnt = random.RandIntBetween(0, 1000000)
	me.LastTime = time.Unix(random.RandInt64Between(0, 1000000), random.RandInt64Between(0, 1000000))
	me.EBHash = primitives.RandomHash()
	me.EntryHash = primitives.RandomHash()
	me.DBHeight = random.RandUInt32()
	return me
}

func (s *MissingEntry) IsSameAs(b *MissingEntry) bool {
	if s.EBHash.IsSameAs(b.EBHash) == false {
		return false
	}
	if s.EntryHash.IsSameAs(b.EntryHash) == false {
		return false
	}
	if s.Cnt != b.Cnt {
		return false
	}
	if s.LastTime.Sub(b.LastTime).Nanoseconds() != 0 {
		return false
	}
	return s.DBHeight == b.DBHeight
}

func (s *MissingEntry) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "MissingEntry.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)

	err = buf.PushVarInt(uint64(s.Cnt))
	if err != nil {
		return nil, err
	}
	err = buf.PushVarInt(uint64(s.LastTime.UnixNano()))
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(s.EBHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(s.EntryHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(s.DBHeight)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (s *MissingEntry) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	s.EBHash = primitives.NewZeroHash()
	s.EntryHash = primitives.NewZeroHash()

	newData = p
	buf := primitives.NewBuffer(p)

	i, err := buf.PopVarInt()
	if err != nil {
		return
	}
	s.Cnt = int(i)
	i, err = buf.PopVarInt()
	if err != nil {
		return
	}
	s.LastTime = time.Unix(int64(i/1000000000), int64(i%1000000000))

	err = buf.PopBinaryMarshallable(s.EBHash)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(s.EntryHash)
	if err != nil {
		return
	}

	s.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return
	}

	newData = buf.DeepCopyBytes()
	return
}

func (s *MissingEntry) UnmarshalBinary(p []byte) error {
	_, err := s.UnmarshalBinaryData(p)
	return err
}
