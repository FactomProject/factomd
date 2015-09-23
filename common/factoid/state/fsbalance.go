// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Defines the state for factoid.  By using the proper
// interfaces, the functionality of factoid can be imported
// into any framework.
package state

import (
	"bytes"
	"encoding/binary"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
)

type FSbalance struct {
	IBlock
	number uint64
}

func (FSbalance) GetNewInstance() IBlock {
	return new(FSbalance)
}

func (FSbalance) GetDBHash() IHash {
	return Sha([]byte("FSbalance"))
}

func (f *FSbalance) UnmarshalBinaryData(data []byte) ([]byte, error) {
	num, data := binary.BigEndian.Uint64(data), data[8:]
	f.number = num
	return data, nil
}

func (f *FSbalance) UnmarshalBinary(data []byte) error {
	data, err := f.UnmarshalBinaryData(data)
	return err
}

func (f FSbalance) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer
	binary.Write(&out, binary.BigEndian, uint64(f.number))
	return out.Bytes(), nil
}
