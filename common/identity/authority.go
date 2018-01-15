// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"encoding/json"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/util/atomic"
)

type Authority struct {
	AuthorityChainID  interfaces.IHash
	ManagementChainID interfaces.IHash
	MatryoshkaHash    interfaces.IHash
	SigningKey        primitives.PublicKey
	Status            atomic.AtomicUint8
	AnchorKeys        []AnchorSigningKey

	KeyHistory []HistoricKey
}

var _ interfaces.BinaryMarshallable = (*Authority)(nil)

func RandomAuthority() *Authority {
	a := new(Authority)

	a.AuthorityChainID = primitives.RandomHash()
	a.ManagementChainID = primitives.RandomHash()
	a.MatryoshkaHash = primitives.RandomHash()

	a.SigningKey = *primitives.RandomPrivateKey().Pub
	a.Status.Store(random.RandUInt8())

	l := random.RandIntBetween(1, 10)
	for i := 0; i < l; i++ {
		a.AnchorKeys = append(a.AnchorKeys, *RandomAnchorSigningKey())
	}

	l = random.RandIntBetween(1, 10)
	for i := 0; i < l; i++ {
		a.KeyHistory = append(a.KeyHistory, *RandomHistoricKey())
	}

	return a
}

func (e *Authority) IsSameAs(b *Authority) bool {
	if e.AuthorityChainID.IsSameAs(b.AuthorityChainID) == false {
		return false
	}
	if e.ManagementChainID.IsSameAs(b.ManagementChainID) == false {
		return false
	}
	if e.MatryoshkaHash.IsSameAs(b.MatryoshkaHash) == false {
		return false
	}
	if e.SigningKey.IsSameAs(&b.SigningKey) == false {
		return false
	}
	if e.Status != b.Status {
		return false
	}

	if len(e.AnchorKeys) != len(b.AnchorKeys) {
		return false
	}
	for i := range e.AnchorKeys {
		if e.AnchorKeys[i].IsSameAs(&b.AnchorKeys[i]) == false {
			return false
		}
	}
	if len(e.KeyHistory) != len(b.KeyHistory) {
		return false
	}
	for i := range e.KeyHistory {
		if e.KeyHistory[i].IsSameAs(&b.KeyHistory[i]) == false {
			return false
		}
	}

	return true
}

func (e *Authority) Init() {
	if e.AuthorityChainID == nil {
		e.AuthorityChainID = primitives.NewZeroHash()
	}
	if e.ManagementChainID == nil {
		e.ManagementChainID = primitives.NewZeroHash()
	}
	if e.MatryoshkaHash == nil {
		e.MatryoshkaHash = primitives.NewZeroHash()
	}
}

func (e *Authority) MarshalBinary() ([]byte, error) {
	e.Init()
	buf := primitives.NewBuffer(nil)

	err := buf.PushBinaryMarshallable(e.AuthorityChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.ManagementChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.MatryoshkaHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(&e.SigningKey)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(byte(e.Status))
	if err != nil {
		return nil, err
	}

	l := len(e.AnchorKeys)
	err = buf.PushVarInt(uint64(l))
	if err != nil {
		return nil, err
	}
	for _, v := range e.AnchorKeys {
		err = buf.PushBinaryMarshallable(&v)
		if err != nil {
			return nil, err
		}
	}

	l = len(e.KeyHistory)
	err = buf.PushVarInt(uint64(l))
	if err != nil {
		return nil, err
	}
	for _, v := range e.KeyHistory {
		err = buf.PushBinaryMarshallable(&v)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (e *Authority) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	e.Init()
	newData = p
	buf := primitives.NewBuffer(p)

	err = buf.PopBinaryMarshallable(e.AuthorityChainID)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.ManagementChainID)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(e.MatryoshkaHash)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(&e.SigningKey)
	if err != nil {
		return
	}
	status, err := buf.PopByte()
	if err != nil {
		return
	}
	e.Status.Store(uint8(status))

	l, err := buf.PopVarInt()
	if err != nil {
		return
	}
	for i := 0; i < int(l); i++ {
		var ask AnchorSigningKey
		err = buf.PopBinaryMarshallable(&ask)
		if err != nil {
			return
		}
		e.AnchorKeys = append(e.AnchorKeys, ask)
	}

	l, err = buf.PopVarInt()
	if err != nil {
		return
	}
	for i := 0; i < int(l); i++ {
		var hk HistoricKey
		err = buf.PopBinaryMarshallable(&hk)
		if err != nil {
			return
		}
		e.KeyHistory = append(e.KeyHistory, hk)
	}

	newData = buf.DeepCopyBytes()
	return
}

func (e *Authority) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}

// 1 if fed, 0 if audit, -1 if neither
func (auth *Authority) Type() int {
	if auth.Status.Load() == constants.IDENTITY_FEDERATED_SERVER {
		return 1
	} else if auth.Status.Load() == constants.IDENTITY_AUDIT_SERVER {
		return 0
	}
	return -1
}

func (auth *Authority) VerifySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte) (bool, error) {
	//return true, nil // Testing
	var pub [32]byte
	tmp, err := auth.SigningKey.MarshalBinary()
	if err != nil {
		return false, err
	} else {
		copy(pub[:], tmp)
		valid := ed.VerifyCanonical(&pub, msg, sig)
		if !valid {
			for _, histKey := range auth.KeyHistory {
				histTemp, err := histKey.SigningKey.MarshalBinary()
				if err != nil {
					continue
				}
				copy(pub[:], histTemp)
				if ed.VerifyCanonical(&pub, msg, sig) {
					return true, nil
				}
			}
		} else {
			return true, nil
		}
	}
	return false, nil
}

func (auth *Authority) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		AuthorityChainID  interfaces.IHash   `json:"chainid"`
		ManagementChainID interfaces.IHash   `json:"manageid"`
		MatryoshkaHash    interfaces.IHash   `json:"matroyshka"`
		SigningKey        string             `json:"signingkey"`
		Status            string             `json:"status"`
		AnchorKeys        []AnchorSigningKey `json:"anchorkeys"`
	}{
		AuthorityChainID:  auth.AuthorityChainID,
		ManagementChainID: auth.ManagementChainID,
		MatryoshkaHash:    auth.MatryoshkaHash,
		SigningKey:        auth.SigningKey.String(),
		Status:            statusToJSONString(auth.Status.Load()),
		AnchorKeys:        auth.AnchorKeys,
	})
}

// Only used for marshaling JSON
func statusToJSONString(status uint8) string {
	switch status {
	case constants.IDENTITY_UNASSIGNED:
		return "none"
	case constants.IDENTITY_FEDERATED_SERVER:
		return "federated"
	case constants.IDENTITY_AUDIT_SERVER:
		return "audit"
	case constants.IDENTITY_FULL:
		return "none"
	case constants.IDENTITY_PENDING_FEDERATED_SERVER:
		return "federated"
	case constants.IDENTITY_PENDING_AUDIT_SERVER:
		return "audit"
	case constants.IDENTITY_PENDING_FULL:
		return "none"
	case constants.IDENTITY_SKELETON:
		return "skeleton"
	}
	return "NA"
}
