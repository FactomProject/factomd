// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

// sort.Sort interface implementation
type AuthoritySort []interfaces.IAuthority

func (p AuthoritySort) Len() int {
	return len(p)
}
func (p AuthoritySort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p AuthoritySort) Less(i, j int) bool {
	return bytes.Compare(p[i].GetAuthorityChainID().Bytes(), p[j].GetAuthorityChainID().Bytes()) < 0
}

type Authority struct {
	AuthorityChainID  interfaces.IHash     `json:"identity_chainid"`
	ManagementChainID interfaces.IHash     `json:"management_chaind"`
	MatryoshkaHash    interfaces.IHash     `json:"matryoshka_hash"`
	SigningKey        primitives.PublicKey `json:"signing_key"`
	Status            uint8                `json:"status"`
	AnchorKeys        []AnchorSigningKey   `json:"anchor_keys"`

	KeyHistory      []HistoricKey       `json:"-"`
	Efficiency      uint16              `json:"efficiency"`
	CoinbaseAddress interfaces.IAddress `json:"coinbase_address"`
}

func NewAuthority() *Authority {
	a := new(Authority)
	a.AuthorityChainID = primitives.NewZeroHash()
	a.ManagementChainID = primitives.NewZeroHash()
	a.MatryoshkaHash = primitives.NewZeroHash()
	a.CoinbaseAddress = primitives.NewZeroHash()
	a.Efficiency = 10000

	return a
}

var _ interfaces.BinaryMarshallable = (*Authority)(nil)

func RandomAuthority() *Authority {
	a := NewAuthority()

	a.AuthorityChainID = primitives.RandomHash()
	a.ManagementChainID = primitives.RandomHash()
	a.MatryoshkaHash = primitives.RandomHash()

	a.SigningKey = *primitives.RandomPrivateKey().Pub
	a.Status = random.RandUInt8()

	l := random.RandIntBetween(1, 10)
	for i := 0; i < l; i++ {
		a.AnchorKeys = append(a.AnchorKeys, *RandomAnchorSigningKey())
	}

	l = random.RandIntBetween(1, 10)
	for i := 0; i < l; i++ {
		a.KeyHistory = append(a.KeyHistory, *RandomHistoricKey())
	}

	a.CoinbaseAddress = primitives.RandomHash()
	a.Efficiency = uint16(rand.Intn(10000))

	return a
}

func (a *Authority) GetCoinbaseHumanReadable() string {
	if a.CoinbaseAddress.IsZero() {
		return "No Address"
	}
	add := factoid.NewAddress(a.CoinbaseAddress.Bytes())
	//primitives.ConvertFctAddressToUserStr(add)
	return primitives.ConvertFctAddressToUserStr(add)
}

func (a *Authority) Clone() *Authority {
	b := NewAuthority()
	b.AuthorityChainID.SetBytes(a.AuthorityChainID.Bytes())
	b.ManagementChainID.SetBytes(a.ManagementChainID.Bytes())
	b.MatryoshkaHash.SetBytes(a.MatryoshkaHash.Bytes())
	b.SigningKey = a.SigningKey
	b.Status = a.Status

	b.AnchorKeys = make([]AnchorSigningKey, len(a.AnchorKeys))
	for i := range a.AnchorKeys {
		b.AnchorKeys[i] = a.AnchorKeys[i]
	}

	b.KeyHistory = make([]HistoricKey, len(a.KeyHistory))
	for i := range a.KeyHistory {
		b.KeyHistory[i] = a.KeyHistory[i]
	}

	b.Efficiency = a.Efficiency
	b.CoinbaseAddress = a.CoinbaseAddress

	return b
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

	if e.Efficiency != b.Efficiency {
		return false
	}

	if !e.CoinbaseAddress.IsSameAs(b.CoinbaseAddress) {
		return false
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
	if e.CoinbaseAddress == nil {
		e.CoinbaseAddress = primitives.NewZeroHash()
	}
}

func (e *Authority) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Authority.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	buf := primitives.NewBuffer(nil)

	err = buf.PushBinaryMarshallable(e.AuthorityChainID)
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

	err = buf.PushUInt16(e.Efficiency)
	if err != nil {
		return nil, err
	}

	err = buf.PushIHash(e.CoinbaseAddress)
	if err != nil {
		return nil, err
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
	e.Status = uint8(status)

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

	e.Efficiency, err = buf.PopUInt16()
	if err != nil {
		return nil, err
	}

	e.CoinbaseAddress, err = buf.PopIHash()
	if err != nil {
		return nil, err
	}

	newData = buf.DeepCopyBytes()
	return
}

func (e *Authority) UnmarshalBinary(p []byte) error {
	_, err := e.UnmarshalBinaryData(p)
	return err
}

func (e *Authority) GetAuthorityChainID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("Authority.GetAuthorityChainID() saw an interface that was nil")
		}
	}()

	return e.AuthorityChainID
}

// 1 if fed, 0 if audit, -1 if neither
func (auth *Authority) Type() int {
	if auth.Status == constants.IDENTITY_FEDERATED_SERVER {
		return 1
	} else if auth.Status == constants.IDENTITY_AUDIT_SERVER {
		return 0
	}
	return -1
}

func (auth *Authority) GetSigningKey() []byte {
	if auth == nil {
		return constants.ZERO_HASH // probably bad we got here but worse to let it cause a panic
	}
	return auth.SigningKey[:]
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

func (auth *Authority) MarshalJSON() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Authority.MarshalJSON err:%v", *pe)
		}
	}(&err)
	return json.Marshal(struct {
		AuthorityChainID  interfaces.IHash   `json:"chainid"`
		ManagementChainID interfaces.IHash   `json:"manageid"`
		MatryoshkaHash    interfaces.IHash   `json:"matroyshka"`
		SigningKey        string             `json:"signingkey"`
		Status            string             `json:"status"`
		AnchorKeys        []AnchorSigningKey `json:"anchorkeys"`
		Efficiency        int                `json:"efficiency"`
		CoinbaseAddress   string             `json:"coinbase_address"`
	}{
		AuthorityChainID:  auth.AuthorityChainID,
		ManagementChainID: auth.ManagementChainID,
		MatryoshkaHash:    auth.MatryoshkaHash,
		SigningKey:        auth.SigningKey.String(),
		Status:            statusToJSONString(auth.Status),
		AnchorKeys:        auth.AnchorKeys,
		Efficiency:        int(auth.Efficiency),
		CoinbaseAddress:   primitives.ConvertFctAddressToUserStr(auth.CoinbaseAddress),
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
