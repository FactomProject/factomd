// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

// HistoricKey contains a public key previously associated with a specific authority server, along with the height it was retired
type HistoricKey struct {
	ActiveDBHeight uint32               // Block height the old signing key was retired
	SigningKey     primitives.PublicKey // Old signing key
}

var _ interfaces.BinaryMarshallable = (*HistoricKey)(nil)

// RandomHistoricKey returns a new HistoricKey with random values
func RandomHistoricKey() *HistoricKey {
	hk := new(HistoricKey)

	hk.ActiveDBHeight = random.RandUInt32()
	hk.SigningKey = *primitives.RandomPrivateKey().Pub

	return hk
}

// IsSameAs returns true iff the input is identical this object
func (e *HistoricKey) IsSameAs(b *HistoricKey) bool {
	if e.ActiveDBHeight != b.ActiveDBHeight {
		return false
	}
	if e.SigningKey.IsSameAs(&b.SigningKey) == false {
		return false
	}

	return true
}

// MarshalBinary marshals this object
func (e *HistoricKey) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "HistoricKey.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)

	err = buf.PushUInt32(e.ActiveDBHeight)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(&e.SigningKey)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (e *HistoricKey) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	newData = p
	buf := primitives.NewBuffer(p)

	e.ActiveDBHeight, err = buf.PopUInt32()
	if err != nil {
		return
	}

	err = buf.PopBinaryMarshallable(&e.SigningKey)
	if err != nil {
		return
	}

	newData = buf.DeepCopyBytes()
	return
}

// UnmarshalBinary unmarshals the input data into this object
func (e *HistoricKey) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}
