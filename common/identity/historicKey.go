// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

type HistoricKey struct {
	ActiveDBHeight uint32
	SigningKey     primitives.PublicKey
}

var _ interfaces.BinaryMarshallable = (*HistoricKey)(nil)

func RandomHistoricKey() *HistoricKey {
	hk := new(HistoricKey)

	hk.ActiveDBHeight = random.RandUInt32()
	hk.SigningKey = *primitives.RandomPrivateKey().Pub

	return hk
}

func (e *HistoricKey) IsSameAs(b *HistoricKey) bool {
	if e.ActiveDBHeight != b.ActiveDBHeight {
		return false
	}
	if e.SigningKey.IsSameAs(&b.SigningKey) == false {
		return false
	}

	return true
}

func (e *HistoricKey) MarshalBinary() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushUInt32(e.ActiveDBHeight)
	if err != nil {
		return nil, err
	}

	err = buf.PushBinaryMarshallable(&e.SigningKey)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

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

func (e *HistoricKey) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}
