// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

const (
	// CommitChainSize = 1+6+32+32+32+1+32+64
	// These are the sizes of the members of the data structure below
	CommitChainSize int = 200
)

// CommitChain is a data structure which affects EC balances when committing a new user chain into the Factom blockchain
// see https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#entry_commit
type CommitChain struct {
	Version     uint8                   `json:"version"`     // The version of the CommitChain, currently 0
	MilliTime   *primitives.ByteSlice6  `json:"millitime"`   // The millisecond time stamp (0~=1970) this commit is created
	ChainIDHash interfaces.IHash        `json:"chainidhash"` // The chain id hash is the double hash of the chain id
	Weld        interfaces.IHash        `json:"weld"`        // The double hash of the concatonated (entry hash | chain id)
	EntryHash   interfaces.IHash        `json:"entryhash"`   // SHA512+256 descriptor of the Entry to be the first in the Chain
	Credits     uint8                   `json:"credits"`     // number of entry credits to deduct for this entry, must be 10 < Credits <= 20
	ECPubKey    *primitives.ByteSlice32 `json:"ecpubkey"`    // EC public key that will have balanced reduced
	Sig         *primitives.ByteSlice64 `json:"sig"`         // signature of the chain commit by the public key
}

var _ interfaces.Printable = (*CommitChain)(nil)
var _ interfaces.BinaryMarshallable = (*CommitChain)(nil)
var _ interfaces.ShortInterpretable = (*CommitChain)(nil)
var _ interfaces.IECBlockEntry = (*CommitChain)(nil)
var _ interfaces.ISignable = (*CommitChain)(nil)

// Init initializes all nil objects to their starting values/objects
func (c *CommitChain) Init() {
	if c.MilliTime == nil {
		c.MilliTime = new(primitives.ByteSlice6)
	}
	if c.ChainIDHash == nil {
		c.ChainIDHash = primitives.NewZeroHash()
	}
	if c.Weld == nil {
		c.Weld = primitives.NewZeroHash()
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
}

// IsSameAs only checks if everything in the item is identical.
// It does not catch if the private key holder has created a malleated version
// which is functionally identical in come cases from the protocol perspective,
// but would fail comparison here
func (c *CommitChain) IsSameAs(b interfaces.IECBlockEntry) bool {
	if c == nil || b == nil {
		if c == nil && b == nil {
			return true
		}
		return false
	}
	if c.ECID() != b.ECID() {
		return false
	}

	bb, ok := b.(*CommitChain)
	if ok == false {
		return false
	}

	if c.Version != bb.Version {
		return false
	}
	if c.MilliTime.IsSameAs(bb.MilliTime) == false {
		return false
	}
	if c.ChainIDHash.IsSameAs(bb.ChainIDHash) == false {
		return false
	}
	if c.Weld.IsSameAs(bb.Weld) == false {
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

// String returns this object as a string
func (c *CommitChain) String() string {
	c.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf(" %s\n", "CommitChain"))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "Version", c.Version))
	out.WriteString(fmt.Sprintf("   %-20s %s\n", "MilliTime", c.MilliTime))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "ChainIDHash", c.ChainIDHash.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "Weld", c.Weld.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "EntryHash", c.EntryHash.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "Credits", c.Credits))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "ECPubKey", c.ECPubKey[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "Sig", c.Sig[:3]))

	return (string)(out.DeepCopyBytes())
}

// NewCommitChain creates a newly initialized commit chain
func NewCommitChain() *CommitChain {
	c := new(CommitChain)
	c.Init()
	c.Version = 0
	c.Credits = 0
	return c
}

// GetEntryHash returns the entry hash
func (c *CommitChain) GetEntryHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "CommitChain.GetEntryHash") }()

	return c.EntryHash
}

// Hash marshals the object and computes the sha
func (c *CommitChain) Hash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "CommitChain.Hash") }()

	bin, err := c.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

// IsInterpretable always returns false
func (c *CommitChain) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (c *CommitChain) Interpret() string {
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

// GetTimestamp returns the timestamp in milliseconds
func (c *CommitChain) GetTimestamp() interfaces.Timestamp {
	c.Init()
	a := make([]byte, 2, 8)
	a = append(a, c.MilliTime[:]...)
	milli := uint64(binary.BigEndian.Uint64(a))
	return primitives.NewTimestampFromMilliseconds(milli)
}

// IsValid checks that the commit chain is valid:  11 < Credits <= 20, version==0, and valid signature
func (c *CommitChain) IsValid() bool {
	c.Init()
	//double check the credits in the commit
	if c.Credits < 11 || c.Version != 0 || c.Credits > 20 {
		return false
	}

	//if there were no errors in processing the signature, formatting or if didn't validate
	if nil == c.ValidateSignatures() {
		return true
	}
	return false
}

// GetHash marshals the entire object and computes the sha
func (c *CommitChain) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "CommitChain.GetHash") }()

	data, _ := c.MarshalBinary()
	return primitives.Sha(data)
}

// GetSigHash marshals the object covered by the signature, and computes its sha (version through entry credits hashed)
func (c *CommitChain) GetSigHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "CommitChain.GetSigHash") }()

	data := c.CommitMsg()
	return primitives.Sha(data)
}

// MarshalBinarySig marshals the object covered by the signature (version through entry credits marshalled)
// If this serialized data is hashed, it becomes the transaction hash of chain commit. (version through entry credits)
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

// MarshalBinaryTransaction partially marshals the object (version through pub key)
// NOTE: Contrary to what the name implies, this is not used to get a transaction hash, that
// seems to be done with the MarshalBinarySig function. Its unclear the origin of this function here.
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

// MarshalBinary marshals the entire object
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

// Sign signs the object with the input private key
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

// ValidateSignatures validates that the object is properly signed
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

// ECID returns the hard coded type ECIDChainCommit
func (c *CommitChain) ECID() byte {
	return constants.ECIDChainCommit
}

// UnmarshalBinaryData unmarshals the input data into this object
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

// UnmarshalBinary unmarshals the input data into this object
func (c *CommitChain) UnmarshalBinary(data []byte) (err error) {
	_, err = c.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (c *CommitChain) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(c)
}

// JSONString returns the json encoded string
func (c *CommitChain) JSONString() (string, error) {
	return primitives.EncodeJSONString(c)
}
