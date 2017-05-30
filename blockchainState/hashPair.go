// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type HashPair struct {
	KeyMR  *primitives.Hash
	Hash   *primitives.Hash
	Height uint32
}

var _ interfaces.BinaryMarshallable = (*HashPair)(nil)

func NewHashPair() *HashPair {
	hp := new(HashPair)
	hp.Init()
	return hp
}

func (e *HashPair) Copy() *HashPair {
	hp := NewHashPair()
	hp.KeyMR = e.KeyMR.Copy().(*primitives.Hash)
	hp.Hash = e.Hash.Copy().(*primitives.Hash)
	hp.Height = e.Height
	return hp
}

func (e *HashPair) Init() {
	if e.KeyMR == nil {
		e.KeyMR = primitives.NewZeroHash().(*primitives.Hash)
	}
	if e.Hash == nil {
		e.Hash = primitives.NewZeroHash().(*primitives.Hash)
	}
}

func (e *HashPair) MarshalBinary() ([]byte, error) {
	e.Init()
	buf := primitives.NewBuffer(nil)

	err := buf.PushBinaryMarshallable(e.KeyMR)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.Hash)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.Height)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *HashPair) UnmarshalBinaryData(data []byte) ([]byte, error) {
	e.Init()
	buf := primitives.NewBuffer(data)

	err := buf.PopBinaryMarshallable(e.KeyMR)
	if err != nil {
		return nil, err
	}
	err = buf.PopBinaryMarshallable(e.Hash)
	if err != nil {
		return nil, err
	}
	e.Height, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *HashPair) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}
