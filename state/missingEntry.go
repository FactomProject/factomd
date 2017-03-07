// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

type MissingEntryBlock struct {
	ebhash   interfaces.IHash
	dbheight uint32
}

var _ interfaces.BinaryMarshallable = (*MissingEntryBlock)(nil)

func (s *MissingEntryBlock) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (s *MissingEntryBlock) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	return
}

func (s *MissingEntryBlock) UnmarshalBinary(p []byte) error {
	return nil
}

type MissingEntry struct {
	ebhash    interfaces.IHash
	entryhash interfaces.IHash
	dbheight  uint32
}

var _ interfaces.BinaryMarshallable = (*MissingEntry)(nil)

func (s *MissingEntry) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (s *MissingEntry) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	return
}

func (s *MissingEntry) UnmarshalBinary(p []byte) error {
	return nil
}
