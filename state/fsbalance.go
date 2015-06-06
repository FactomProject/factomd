// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Defines the state for simplecoin.  By using the proper
// interfaces, the functionality of simplecoin can be imported
// into any framework.
package state

import (
    "bytes"
    "encoding/binary"
    sc "github.com/FactomProject/simplecoin"
)

type IFSbalance interface {
    sc.IBlock
    getNumber() uint64
    setNumber(uint64)
}

type FSbalance struct {
    sc.IBlock
    number uint64  
}

func (FSbalance) GetNewInstance() sc.IBlock {
    return new(FSbalance)
}

func (FSbalance) GetDBHash() sc.IHash {
    return sc.Sha([]byte("FSbalance"))
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