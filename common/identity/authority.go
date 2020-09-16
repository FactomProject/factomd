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

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

// AuthoritySort is a slice of IAuthoritys
// sort.Sort interface implementation
type AuthoritySort []interfaces.IAuthority

// Len returns the length of the slice
func (p AuthoritySort) Len() int {
	return len(p)
}

// Swap swaps the data at the input indices 'i' and 'j' in the slice
func (p AuthoritySort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Less returns true if the data at the ith index is less than the data at the jth index
func (p AuthoritySort) Less(i, j int) bool {
	return bytes.Compare(p[i].GetAuthorityChainID().Bytes(), p[j].GetAuthorityChainID().Bytes()) < 0
}

// Authority is a struct containing information related to an authority server for the factom protocol.
type Authority struct {
	AuthorityChainID  interfaces.IHash     `json:"chainid"` // chain id for the authority server
	ManagementChainID interfaces.IHash     `json:"manageid"`
	MatryoshkaHash    interfaces.IHash     `json:"matroyshka"`
	SigningKey        primitives.PublicKey `json:"signingkey"`
	Status            uint8                `json:"status"` // Identity status type, see constants.go
	AnchorKeys        []AnchorSigningKey   `json:"anchorkeys"`

	KeyHistory      []HistoricKey       `json:"-"`               // Slice of old signing keys previously associated with this authority server along with the dbheight it was retired
	Efficiency      uint16              `json:"efficiency"`      // The efficiency of this authority server in hundredths of a percent: 10000 = 100%
	CoinbaseAddress interfaces.IAddress `json:"coinbaseaddress"` // The Factoid address payments are sent to for this server
}

// NewAuthority returns a new Authority with hashes initialized to the zero hash, and an efficiency of 10000 (100%)
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

// RandomAuthority returns a new Authority with random starting values
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
	a.Efficiency = uint16(rand.Intn(10001)) // rand is non-inclusive of the number, so this returns 0-10000

	return a
}

// GetCoinbaseHumanReadable returns the human readable factoid address associated with this authority server
func (a *Authority) GetCoinbaseHumanReadable() string {
	if a.CoinbaseAddress.IsZero() {
		return "No Address"
	}
	add := factoid.NewAddress(a.CoinbaseAddress.Bytes())
	//primitives.ConvertFctAddressToUserStr(add)
	return primitives.ConvertFctAddressToUserStr(add)
}

// Clone returns a new, identical copy of this Authority
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

// IsSameAs returns true iff the input object is identical to this object
func (a *Authority) IsSameAs(b *Authority) bool {
	if a.AuthorityChainID.IsSameAs(b.AuthorityChainID) == false {
		return false
	}
	if a.ManagementChainID.IsSameAs(b.ManagementChainID) == false {
		return false
	}
	if a.MatryoshkaHash.IsSameAs(b.MatryoshkaHash) == false {
		return false
	}
	if a.SigningKey.IsSameAs(&b.SigningKey) == false {
		return false
	}
	if a.Status != b.Status {
		return false
	}

	if len(a.AnchorKeys) != len(b.AnchorKeys) {
		return false
	}
	for i := range a.AnchorKeys {
		if a.AnchorKeys[i].IsSameAs(&b.AnchorKeys[i]) == false {
			return false
		}
	}
	if len(a.KeyHistory) != len(b.KeyHistory) {
		return false
	}
	for i := range a.KeyHistory {
		if a.KeyHistory[i].IsSameAs(&b.KeyHistory[i]) == false {
			return false
		}
	}

	if a.Efficiency != b.Efficiency {
		return false
	}

	if !a.CoinbaseAddress.IsSameAs(b.CoinbaseAddress) {
		return false
	}

	return true
}

// Init initializes any nil hashes to the zero hash
func (a *Authority) Init() {
	if a.AuthorityChainID == nil {
		a.AuthorityChainID = primitives.NewZeroHash()
	}
	if a.ManagementChainID == nil {
		a.ManagementChainID = primitives.NewZeroHash()
	}
	if a.MatryoshkaHash == nil {
		a.MatryoshkaHash = primitives.NewZeroHash()
	}
	if a.CoinbaseAddress == nil {
		a.CoinbaseAddress = primitives.NewZeroHash()
	}
}

// MarshalBinary marshals the object
func (a *Authority) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Authority.MarshalBinary err:%v", *pe)
		}
	}(&err)
	a.Init()
	buf := primitives.NewBuffer(nil)

	err = buf.PushBinaryMarshallable(a.AuthorityChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(a.ManagementChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(a.MatryoshkaHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(&a.SigningKey)
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(byte(a.Status))
	if err != nil {
		return nil, err
	}

	l := len(a.AnchorKeys)
	err = buf.PushVarInt(uint64(l))
	if err != nil {
		return nil, err
	}
	for _, v := range a.AnchorKeys {
		err = buf.PushBinaryMarshallable(&v)
		if err != nil {
			return nil, err
		}
	}

	l = len(a.KeyHistory)
	err = buf.PushVarInt(uint64(l))
	if err != nil {
		return nil, err
	}
	for _, v := range a.KeyHistory {
		err = buf.PushBinaryMarshallable(&v)
		if err != nil {
			return nil, err
		}
	}

	err = buf.PushUInt16(a.Efficiency)
	if err != nil {
		return nil, err
	}

	err = buf.PushIHash(a.CoinbaseAddress)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this object
func (a *Authority) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	a.Init()
	newData = p
	buf := primitives.NewBuffer(p)

	err = buf.PopBinaryMarshallable(a.AuthorityChainID)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(a.ManagementChainID)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(a.MatryoshkaHash)
	if err != nil {
		return
	}
	err = buf.PopBinaryMarshallable(&a.SigningKey)
	if err != nil {
		return
	}
	status, err := buf.PopByte()
	if err != nil {
		return
	}
	a.Status = uint8(status)

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
		a.AnchorKeys = append(a.AnchorKeys, ask)
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
		a.KeyHistory = append(a.KeyHistory, hk)
	}

	a.Efficiency, err = buf.PopUInt16()
	if err != nil {
		return nil, err
	}

	a.CoinbaseAddress, err = buf.PopIHash()
	if err != nil {
		return nil, err
	}

	newData = buf.DeepCopyBytes()
	return
}

// UnmarshalBinary unmarshals the input data into this object
func (a *Authority) UnmarshalBinary(p []byte) error {
	_, err := a.UnmarshalBinaryData(p)
	return err
}

// GetAuthorityChainID returns the authority chain id
func (a *Authority) GetAuthorityChainID() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "Authority.GetAuthorityChainID") }()

	return a.AuthorityChainID
}

// Type returns: 1 if fed, 0 if audit, -1 if neither
func (a *Authority) Type() int {
	if a.Status == constants.IDENTITY_FEDERATED_SERVER {
		return 1
	} else if a.Status == constants.IDENTITY_AUDIT_SERVER {
		return 0
	}
	return -1
}

// GetSigningKey returns the signing key
func (a *Authority) GetSigningKey() []byte {
	if a == nil {
		return constants.ZERO_HASH // probably bad we got here but worse to let it cause a panic
	}
	return a.SigningKey[:]
}

// VerifySignature verifies input message and signature with the internal public signing key
func (a *Authority) VerifySignature(msg []byte, sig *[constants.SIGNATURE_LENGTH]byte) (bool, error) {
	//return true, nil // Testing
	var pub [32]byte
	tmp, err := a.SigningKey.MarshalBinary()
	if err != nil {
		return false, err
	}
	copy(pub[:], tmp)
	valid := ed.VerifyCanonical(&pub, msg, sig)
	if !valid {
		for _, histKey := range a.KeyHistory {
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

	return false, nil
}

// MarshalJSON marshals this object as json
func (a *Authority) MarshalJSON() (rval []byte, err error) {
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
		CoinbaseAddress   string             `json:"coinbaseaddress"`
	}{
		AuthorityChainID:  a.AuthorityChainID,
		ManagementChainID: a.ManagementChainID,
		MatryoshkaHash:    a.MatryoshkaHash,
		SigningKey:        a.SigningKey.String(),
		Status:            statusToJSONString(a.Status),
		AnchorKeys:        a.AnchorKeys,
		Efficiency:        int(a.Efficiency),
		CoinbaseAddress:   primitives.ConvertFctAddressToUserStr(a.CoinbaseAddress),
	})
}

// statusToJSONString returns the input status as a string. Only used for marshaling JSON
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
