// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const (
	// CommitChainSize = 1+6+32+32+32+1+32+64
	CommitChainSize int = 200
)

type CommitChain struct {
	Version     uint8                   `json:"version"`
	MilliTime   *primitives.ByteSlice6  `json:"millitime"`
	ChainIDHash interfaces.IHash        `json:"chainidhash"`
	Weld        interfaces.IHash        `json:"weld"`
	EntryHash   interfaces.IHash        `json:"entryhash"`
	Credits     uint8                   `json:"credits"`
	ECPubKey    *primitives.ByteSlice32 `json:"ecpubkey"`
	Sig         *primitives.ByteSlice64 `json:"sig"`
}

var _ interfaces.Printable = (*CommitChain)(nil)
var _ interfaces.BinaryMarshallable = (*CommitChain)(nil)
var _ interfaces.ShortInterpretable = (*CommitChain)(nil)
var _ interfaces.IECBlockEntry = (*CommitChain)(nil)
var _ interfaces.ISignable = (*CommitChain)(nil)

func (e *CommitChain) Init() {
	if e.MilliTime == nil {
		e.MilliTime = new(primitives.ByteSlice6)
	}
	if e.ChainIDHash == nil {
		e.ChainIDHash = primitives.NewZeroHash()
	}
	if e.Weld == nil {
		e.Weld = primitives.NewZeroHash()
	}
	if e.EntryHash == nil {
		e.EntryHash = primitives.NewZeroHash()
	}
	if e.ECPubKey == nil {
		e.ECPubKey = new(primitives.ByteSlice32)
	}
	if e.Sig == nil {
		e.Sig = new(primitives.ByteSlice64)
	}
}

//this function only checks if everything in the item is identical.
// It does not catch if the private key holder has created a malleated version
//which is functionally identical in come cases from the protocol perspective,
//but would fail comparison here
func (a *CommitChain) IsSameAs(b interfaces.IECBlockEntry) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}
	if a.ECID() != b.ECID() {
		return false
	}

	bb, ok := b.(*CommitChain)
	if ok == false {
		return false
	}

	if a.Version != bb.Version {
		return false
	}
	if a.MilliTime.IsSameAs(bb.MilliTime) == false {
		return false
	}
	if a.ChainIDHash.IsSameAs(bb.ChainIDHash) == false {
		return false
	}
	if a.Weld.IsSameAs(bb.Weld) == false {
		return false
	}
	if a.EntryHash.IsSameAs(bb.EntryHash) == false {
		return false
	}
	if a.Credits != bb.Credits {
		return false
	}
	if a.ECPubKey.IsSameAs(bb.ECPubKey) == false {
		return false
	}
	if a.Sig.IsSameAs(bb.Sig) == false {
		return false
	}

	return true
}

func (e *CommitChain) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf(" %s\n", "CommitChain"))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "Version", e.Version))
	out.WriteString(fmt.Sprintf("   %-20s %s\n", "MilliTime", e.MilliTime))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "ChainIDHash", e.ChainIDHash.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "Weld", e.Weld.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "EntryHash", e.EntryHash.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "Credits", e.Credits))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "ECPubKey", e.ECPubKey[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "Sig", e.Sig[:3]))

	return (string)(out.DeepCopyBytes())
}

func NewCommitChain() *CommitChain {
	c := new(CommitChain)
	c.Version = 0
	c.MilliTime = new(primitives.ByteSlice6)
	c.ChainIDHash = primitives.NewZeroHash()
	c.Weld = primitives.NewZeroHash()
	c.EntryHash = primitives.NewZeroHash()
	c.Credits = 0
	c.ECPubKey = new(primitives.ByteSlice32)
	c.Sig = new(primitives.ByteSlice64)
	return c
}

func (a *CommitChain) GetEntryHash() interfaces.IHash {
	return a.EntryHash
}

func (e *CommitChain) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

func (b *CommitChain) IsInterpretable() bool {
	return false
}

func (b *CommitChain) Interpret() string {
	return ""
}

// CommitMsg returns the binary marshalled message section of the CommitEntry
// that is covered by the CommitEntry.Sig.
func (c *CommitChain) CommitMsg() []byte {
	p, err := c.MarshalBinarySig()
	if err != nil {
		return []byte{byte(0)}
	}
	return p
}

// Return the timestamp
func (c *CommitChain) GetTimestamp() interfaces.Timestamp {
	c.Init()
	a := make([]byte, 2, 8)
	a = append(a, c.MilliTime[:]...)
	milli := uint64(binary.BigEndian.Uint64(a))
	return primitives.NewTimestampFromMilliseconds(milli)
}

func (c *CommitChain) IsValid() bool {
	c.Init()
	//double check the credits in the commit
	if c.Credits < 11 || c.Version != 0 || c.Credits > 20 {
		return false
	}

	//if there were no errors in processing the signature, formatting or if didn't validate
	if nil == c.ValidateSignatures() {
		return true
	} else {
		return false
	}
}

func (c *CommitChain) GetHash() interfaces.IHash {
	data, _ := c.MarshalBinary()
	return primitives.Sha(data)
}

func (c *CommitChain) GetSigHash() interfaces.IHash {
	data := c.CommitMsg()
	return primitives.Sha(data)
}

func (c *CommitChain) MarshalBinarySig() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "CommitChain.MarshalBinarySig err:%v", *pe)
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

	// 32 byte double sha256 hash of the ChainID
	err = buf.PushBinaryMarshallable(c.ChainIDHash)
	if err != nil {
		return nil, err
	}

	// 32 byte Commit Weld sha256(sha256(Entry Hash + ChainID))
	err = buf.PushBinaryMarshallable(c.Weld)
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

// Transaction hash of chain commit. (version through pub key hashed)
func (c *CommitChain) MarshalBinaryTransaction() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "CommitChain.MarshalBinaryTransaction err:%v", *pe)
		}
	}(&err)
	c.Init()
	buf := new(primitives.Buffer)

	b, err := c.MarshalBinarySig()
	if err != nil {
		return nil, err
	}

	buf.Write(b)

	// 32 byte Public Key
	buf.Write(c.ECPubKey[:])

	return buf.DeepCopyBytes(), nil

}

func (c *CommitChain) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "CommitChain.MarshalBinary err:%v", *pe)
		}
	}(&err)
	c.Init()
	buf := new(primitives.Buffer)

	b, err := c.MarshalBinaryTransaction()
	if err != nil {
		return nil, err
	}

	buf.Write(b)

	// 32 byte Public Key
	//buf.Write(c.ECPubKey[:])

	// 64 byte Signature
	buf.Write(c.Sig[:])

	return buf.DeepCopyBytes(), nil
}

func (c *CommitChain) Sign(privateKey []byte) error {
	c.Init()
	sig, err := primitives.SignSignable(privateKey, c)
	if err != nil {
		return err
	}
	if c.Sig == nil {
		c.Sig = new(primitives.ByteSlice64)
	}
	err = c.Sig.UnmarshalBinary(sig)
	if err != nil {
		return err
	}
	pub, err := primitives.PrivateKeyToPublicKey(privateKey)
	if err != nil {
		return err
	}
	if c.ECPubKey == nil {
		c.ECPubKey = new(primitives.ByteSlice32)
	}
	err = c.ECPubKey.UnmarshalBinary(pub)
	if err != nil {
		return err
	}
	return nil
}

func (c *CommitChain) ValidateSignatures() error {
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

func (c *CommitChain) ECID() byte {
	return constants.ECIDChainCommit
}

func (c *CommitChain) UnmarshalBinaryData(data []byte) ([]byte, error) {
	c.Init()
	buf := primitives.NewBuffer(data)
	var err error

	// 1 byte Version
	c.Version, err = buf.PopUInt8()
	if err != nil {
		return nil, err
	}

	c.MilliTime = new(primitives.ByteSlice6)
	err = buf.PopBinaryMarshallable(c.MilliTime)
	if err != nil {
		return nil, err
	}

	// 32 byte ChainIDHash
	err = buf.PopBinaryMarshallable(c.ChainIDHash)
	if err != nil {
		return nil, err
	}

	// 32 byte Weld
	err = buf.PopBinaryMarshallable(c.Weld)
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
	c.ECPubKey = new(primitives.ByteSlice32)
	err = buf.PopBinaryMarshallable(c.ECPubKey)
	if err != nil {
		return nil, err
	}

	c.Sig = new(primitives.ByteSlice64)
	err = buf.PopBinaryMarshallable(c.Sig)
	if err != nil {
		return nil, err
	}

	err = c.ValidateSignatures()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (c *CommitChain) UnmarshalBinary(data []byte) (err error) {
	_, err = c.UnmarshalBinaryData(data)
	return
}

func (e *CommitChain) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommitChain) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}
