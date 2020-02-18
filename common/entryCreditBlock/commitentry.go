// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"encoding/binary"
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const (
	// CommitEntrySize = 1 + 6 + 32 + 1 + 32 + 64
	CommitEntrySize int = 136
)

type CommitEntry struct {
	Version   uint8                   `json:"version"`
	MilliTime *primitives.ByteSlice6  `json:"millitime"`
	EntryHash interfaces.IHash        `json:"entryhash"`
	Credits   uint8                   `json:"credits"`
	ECPubKey  *primitives.ByteSlice32 `json:"ecpubkey"`
	Sig       *primitives.ByteSlice64 `json:"sig"`
}

var _ interfaces.Printable = (*CommitEntry)(nil)
var _ interfaces.BinaryMarshallable = (*CommitEntry)(nil)
var _ interfaces.ShortInterpretable = (*CommitEntry)(nil)
var _ interfaces.IECBlockEntry = (*CommitEntry)(nil)
var _ interfaces.ISignable = (*CommitEntry)(nil)

func (c *CommitEntry) Init() {
	if c.MilliTime == nil {
		c.MilliTime = new(primitives.ByteSlice6)
	}
	if c.EntryHash == nil {
		c.EntryHash = primitives.NewZeroHash()
	}
	if c.ECPubKey == nil {
		c.ECPubKey = new(primitives.ByteSlice32)
	}
	if c.Sig == nil {
		c.Sig = new(primitives.ByteSlice64)
	}
	/*
		if e.SigHash == nil {
			e.SigHash = primitives.NewZeroHash()
		}
	*/
}

//this function only checks if everything in the item is identical.
// It does not catch if the private key holder has created a malleated version
//which is functionally identical in come cases from the protocol perspective,
//but would fail comparison here
func (c *CommitEntry) IsSameAs(b interfaces.IECBlockEntry) bool {
	if c == nil || b == nil {
		if c == nil && b == nil {
			return true
		}
		return false
	}
	if c.ECID() != b.ECID() {
		return false
	}

	bb, ok := b.(*CommitEntry)
	if ok == false {
		return false
	}

	if c.Version != bb.Version {
		return false
	}
	if c.MilliTime.IsSameAs(bb.MilliTime) == false {
		return false
	}
	if c.EntryHash.IsSameAs(bb.EntryHash) == false {
		return false
	}
	if c.Credits != bb.Credits {
		return false
	}
	if c.ECPubKey.IsSameAs(bb.ECPubKey) == false {
		return false
	}
	if c.Sig.IsSameAs(bb.Sig) == false {
		return false
	}

	return true
}

func (c *CommitEntry) String() string {
	//var out primitives.Buffer
	//out.WriteString(fmt.Sprintf(" %s\n", "CommitEntry"))
	//out.WriteString(fmt.Sprintf("   %-20s %d\n", "Version", e.Version))
	//out.WriteString(fmt.Sprintf("   %-20s %s\n", "MilliTime", e.MilliTime))
	//out.WriteString(fmt.Sprintf("   %-20s %x\n", "EntryHash", e.EntryHash.Bytes()[:3]))
	//out.WriteString(fmt.Sprintf("   %-20s %d\n", "Credits", e.Credits))
	//out.WriteString(fmt.Sprintf("   %-20s %x\n", "ECPubKey", e.ECPubKey[:3]))
	//out.WriteString(fmt.Sprintf("   %-20s %x\n", "Sig", e.Sig[:3]))
	//
	//return (string)(out.DeepCopyBytes())
	return fmt.Sprintf("ehash[%x] Credits[%d] PublicKey[%x] Sig[%x]", c.EntryHash.Bytes()[:3], c.Credits, c.ECPubKey[:3], c.Sig[:3])
}

func (c *CommitEntry) GetEntryHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("CommitEntry.GetEntryHash() saw an interface that was nil")
		}
	}()

	return c.EntryHash
}

func NewCommitEntry() *CommitEntry {
	c := new(CommitEntry)
	c.Version = 0
	c.MilliTime = new(primitives.ByteSlice6)
	c.EntryHash = primitives.NewZeroHash()
	c.Credits = 0
	c.ECPubKey = new(primitives.ByteSlice32)
	c.Sig = new(primitives.ByteSlice64)
	return c
}

func (c *CommitEntry) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("CommitEntry.Hash() saw an interface that was nil")
		}
	}()

	bin, err := c.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

func (c *CommitEntry) IsInterpretable() bool {
	return false
}

func (c *CommitEntry) Interpret() string {
	return ""
}

// CommitMsg returns the binary marshalled message section of the CommitEntry
// that is covered by the CommitEntry.Sig.
func (c *CommitEntry) CommitMsg() []byte {
	p, err := c.MarshalBinarySig()
	if err != nil {
		return []byte{byte(0)}
	}
	return p
}

// Return the timestamp
func (c *CommitEntry) GetTimestamp() interfaces.Timestamp {
	a := make([]byte, 2, 8)
	a = append(a, c.MilliTime[:]...)
	milli := uint64(binary.BigEndian.Uint64(a))
	return primitives.NewTimestampFromMilliseconds(milli)
}

func (c *CommitEntry) IsValid() bool {
	//double check the credits in the commit
	if c.Credits < 1 || c.Version != 0 || c.Credits > 10 {
		return false
	}

	//if there were no errors in processing the signature, formatting or if didn't validate
	if nil == c.ValidateSignatures() {
		return true
	} else {
		return false
	}
}

func (c *CommitEntry) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("CommitEntry.GetHash() saw an interface that was nil")
		}
	}()

	h, _ := c.MarshalBinary()
	return primitives.Sha(h)
}

func (c *CommitEntry) GetSigHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("CommitEntry.GetSigHash() saw an interface that was nil")
		}
	}()

	data := c.CommitMsg()
	return primitives.Sha(data)
}

func (c *CommitEntry) MarshalBinarySig() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "CommitEntry.MarshalBinarySig err:%v", *pe)
		}
	}(&err)
	c.Init()
	buf := primitives.NewBuffer(nil)

	// 1 byte Version
	err = buf.PushUInt8(c.Version)
	if err != nil {
		return nil, err
	}

	// 6 byte MilliTime
	err = buf.PushBinaryMarshallable(c.MilliTime)
	if err != nil {
		return nil, err
	}

	// 32 byte Entry Hash
	err = buf.PushBinaryMarshallable(c.EntryHash)
	if err != nil {
		return nil, err
	}

	// 1 byte number of Entry Credits
	err = buf.PushUInt8(c.Credits)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// Transaction hash of entry commit. (version through pub key hashed)
func (c *CommitEntry) MarshalBinaryTransaction() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "CommitEntry.MarshalBinaryTransaction err:%v", *pe)
		}
	}(&err)
	b, err := c.MarshalBinarySig()
	if err != nil {
		return nil, err
	}
	buf := primitives.NewBuffer(b)

	// 32 byte Public Key
	err = buf.PushBinaryMarshallable(c.ECPubKey)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (c *CommitEntry) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "CommitEntry.MarshalBinary err:%v", *pe)
		}
	}(&err)
	b, err := c.MarshalBinaryTransaction()
	if err != nil {
		return nil, err
	}
	buf := primitives.NewBuffer(b)

	// 64 byte Signature
	err = buf.PushBinaryMarshallable(c.Sig)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (c *CommitEntry) Sign(privateKey []byte) error {
	c.Init()
	sig, err := primitives.SignSignable(privateKey, c)
	if err != nil {
		return err
	}
	err = c.Sig.UnmarshalBinary(sig)
	if err != nil {
		return err
	}
	pub, err := primitives.PrivateKeyToPublicKey(privateKey)
	if err != nil {
		return err
	}
	err = c.ECPubKey.UnmarshalBinary(pub)
	if err != nil {
		return err
	}
	return nil
}

func (c *CommitEntry) ValidateSignatures() error {
	if c.ECPubKey == nil {
		return fmt.Errorf("No public key present")
	}
	if c.Sig == nil {
		return fmt.Errorf("No signature present")
	}
	data, err := c.MarshalBinarySig()
	if err != nil {
		return err
	}
	return primitives.VerifySignature(data, c.ECPubKey[:], c.Sig[:])
}

func (c *CommitEntry) ECID() byte {
	return constants.ECIDEntryCommit
}

func (c *CommitEntry) UnmarshalBinaryData(data []byte) ([]byte, error) {
	c.Init()
	buf := primitives.NewBuffer(data)
	var err error

	c.Version, err = buf.PopUInt8()
	if err != nil {
		return nil, err
	}

	// 6 byte MilliTime
	err = buf.PopBinaryMarshallable(c.MilliTime)
	if err != nil {
		return nil, err
	}

	// 32 byte Entry Hash
	err = buf.PopBinaryMarshallable(c.EntryHash)
	if err != nil {
		return nil, err
	}

	// 1 byte number of Entry Credits
	c.Credits, err = buf.PopUInt8()
	if err != nil {
		return nil, err
	}

	// 32 byte Public Key
	err = buf.PopBinaryMarshallable(c.ECPubKey)
	if err != nil {
		return nil, err
	}

	// 64 byte Signature
	err = buf.PopBinaryMarshallable(c.Sig)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (c *CommitEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = c.UnmarshalBinaryData(data)
	return
}

func (c *CommitEntry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(c)
}

func (c *CommitEntry) JSONString() (string, error) {
	return primitives.EncodeJSONString(c)
}

func (c *CommitEntry) GetVersion() uint8 {
	return c.Version
}

func (c *CommitEntry) GetMilliTime() *primitives.ByteSlice6 {
	return c.MilliTime
}

func (c *CommitEntry) GetCredits() uint8 {
	return c.Credits
}

func (c *CommitEntry) GetECPubKey() *primitives.ByteSlice32 {
	return c.ECPubKey
}

func (c *CommitEntry) GetSig() *primitives.ByteSlice64 {
	return c.Sig
}
