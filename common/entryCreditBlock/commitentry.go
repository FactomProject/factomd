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
	// CommitEntrySize = 1 + 6 + 32 + 1 + 32 + 64
	// These are the sizes of the members of the data structure below
	CommitEntrySize int = 136
)

// CommitEntry is a type of EC block entry which handles commits to the block chain. Related to the EC block because you
// must pay EC to commit something to the chain
type CommitEntry struct {
	Version   uint8                   `json:"version"`   // Version, must be 0
	MilliTime *primitives.ByteSlice6  `json:"millitime"` // Millisecond time stamp for this entry's creation (0~=1970)
	EntryHash interfaces.IHash        `json:"entryhash"` // SHA512+256 descriptor of the Entry to be paid for
	Credits   uint8                   `json:"credits"`   // number of entry credits to deduct for this entry, must be 0 < Credits <= 10
	ECPubKey  *primitives.ByteSlice32 `json:"ecpubkey"`  // EC public key that will have balanced reduced
	Sig       *primitives.ByteSlice64 `json:"sig"`       // signature of the entry commit by the public key
}

var _ interfaces.Printable = (*CommitEntry)(nil)
var _ interfaces.BinaryMarshallable = (*CommitEntry)(nil)
var _ interfaces.ShortInterpretable = (*CommitEntry)(nil)
var _ interfaces.IECBlockEntry = (*CommitEntry)(nil)
var _ interfaces.ISignable = (*CommitEntry)(nil)

// Init initializes all nil objects
func (e *CommitEntry) Init() {
	if e.MilliTime == nil {
		e.MilliTime = new(primitives.ByteSlice6)
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
	/*
		if e.SigHash == nil {
			e.SigHash = primitives.NewZeroHash()
		}
	*/
}

// IsSameAs only checks if everything in the item is identical.
// It does not catch if the private key holder has created a malleated version
// which is functionally identical in come cases from the protocol perspective,
// but would fail comparison here
func (e *CommitEntry) IsSameAs(b interfaces.IECBlockEntry) bool {
	if e == nil || b == nil {
		if e == nil && b == nil {
			return true
		}
		return false
	}
	if e.ECID() != b.ECID() {
		return false
	}

	bb, ok := b.(*CommitEntry)
	if ok == false {
		return false
	}

	if e.Version != bb.Version {
		return false
	}
	if e.MilliTime.IsSameAs(bb.MilliTime) == false {
		return false
	}
	if e.EntryHash.IsSameAs(bb.EntryHash) == false {
		return false
	}
	if e.Credits != bb.Credits {
		return false
	}
	if e.ECPubKey.IsSameAs(bb.ECPubKey) == false {
		return false
	}
	if e.Sig.IsSameAs(bb.Sig) == false {
		return false
	}

	return true
}

// String returns this object as a string
func (e *CommitEntry) String() string {
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
	return fmt.Sprintf("ehash[%x] Credits[%d] PublicKey[%x] Sig[%x]", e.EntryHash.Bytes()[:3], e.Credits, e.ECPubKey[:3], e.Sig[:3])
}

// GetEntryHash returns the entry's hash
func (e *CommitEntry) GetEntryHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "CommitEntry.GetEntryHash") }()

	return e.EntryHash
}

// NewCommitEntry returns a new commit entry
func NewCommitEntry() *CommitEntry {
	c := new(CommitEntry)
	c.Init()
	c.Version = 0
	c.Credits = 0
	return c
}

// Hash marshals the object and computes the sha
func (e *CommitEntry) Hash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "CommitEntry.Hash") }()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

// IsInterpretable always returns false
func (e *CommitEntry) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *CommitEntry) Interpret() string {
	return ""
}

// CommitMsg returns the binary marshalled message section of the CommitEntry
// that is covered by the CommitEntry.Sig.
func (e *CommitEntry) CommitMsg() []byte {
	p, err := e.MarshalBinarySig()
	if err != nil {
		return []byte{byte(0)}
	}
	return p
}

// GetTimestamp returns the timestamp in milliseconds
func (e *CommitEntry) GetTimestamp() interfaces.Timestamp {
	a := make([]byte, 2, 8)
	a = append(a, e.MilliTime[:]...)
	milli := uint64(binary.BigEndian.Uint64(a))
	return primitives.NewTimestampFromMilliseconds(milli)
}

// IsValid checks if the commit entry is valid: 0 < Credits <= 10, version==0, and valid signature
func (e *CommitEntry) IsValid() bool {
	//double check the credits in the commit
	if e.Credits < 1 || e.Version != 0 || e.Credits > 10 {
		return false
	}

	//if there were no errors in processing the signature, formatting or if didn't validate
	if nil == e.ValidateSignatures() {
		return true
	}
	return false
}

// GetHash marshals the object and computes the sha of the data
func (e *CommitEntry) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "CommitEntry.GetHash") }()

	h, _ := e.MarshalBinary()
	return primitives.Sha(h)
}

// GetSigHash computes the hash of the partially marshalled object: (version through entry credits hashed)
func (e *CommitEntry) GetSigHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "CommitEntry.GetSigHash") }()

	data := e.CommitMsg()
	return primitives.Sha(data)
}

// MarshalBinarySig marshals the object covered by the signature (version through entry credits)
// If this serialized data is hashed, it becomes the transaction hash of entry commit. (version through entry credits)
func (e *CommitEntry) MarshalBinarySig() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "CommitEntry.MarshalBinarySig err:%v", *pe)
		}
	}(&err)
	e.Init()
	buf := primitives.NewBuffer(nil)

	// 1 byte Version
	err = buf.PushUInt8(e.Version)
	if err != nil {
		return nil, err
	}

	// 6 byte MilliTime
	err = buf.PushBinaryMarshallable(e.MilliTime)
	if err != nil {
		return nil, err
	}

	// 32 byte Entry Hash
	err = buf.PushBinaryMarshallable(e.EntryHash)
	if err != nil {
		return nil, err
	}

	// 1 byte number of Entry Credits
	err = buf.PushUInt8(e.Credits)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// MarshalBinary marshals the entire object
func (e *CommitEntry) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "CommitEntry.MarshalBinary err:%v", *pe)
		}
	}(&err)
	b, err := e.MarshalBinarySig()
	if err != nil {
		return nil, err
	}
	buf := primitives.NewBuffer(b)

	// 32 byte Public Key
	err = buf.PushBinaryMarshallable(e.ECPubKey)
	if err != nil {
		return nil, err
	}

	// 64 byte Signature
	err = buf.PushBinaryMarshallable(e.Sig)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// Sign signs the object with the input private key
func (e *CommitEntry) Sign(privateKey []byte) error {
	e.Init()
	sig, err := primitives.SignSignable(privateKey, e)
	if err != nil {
		return err
	}
	err = e.Sig.UnmarshalBinary(sig)
	if err != nil {
		return err
	}
	pub, err := primitives.PrivateKeyToPublicKey(privateKey)
	if err != nil {
		return err
	}
	err = e.ECPubKey.UnmarshalBinary(pub)
	if err != nil {
		return err
	}
	return nil
}

// ValidateSignatures validates the object's signature
func (e *CommitEntry) ValidateSignatures() error {
	if e.ECPubKey == nil {
		return fmt.Errorf("No public key present")
	}
	if e.Sig == nil {
		return fmt.Errorf("No signature present")
	}
	data, err := e.MarshalBinarySig()
	if err != nil {
		return err
	}
	return primitives.VerifySignature(data, e.ECPubKey[:], e.Sig[:])
}

// ECID returns the entry credit id ECIDEntryCommit
func (e *CommitEntry) ECID() byte {
	return constants.ECIDEntryCommit
}

// UnmarshalBinaryData unmarshals the input data into this object
func (e *CommitEntry) UnmarshalBinaryData(data []byte) ([]byte, error) {
	e.Init()
	buf := primitives.NewBuffer(data)
	var err error

	e.Version, err = buf.PopUInt8()
	if err != nil {
		return nil, err
	}

	// 6 byte MilliTime
	err = buf.PopBinaryMarshallable(e.MilliTime)
	if err != nil {
		return nil, err
	}

	// 32 byte Entry Hash
	err = buf.PopBinaryMarshallable(e.EntryHash)
	if err != nil {
		return nil, err
	}

	// 1 byte number of Entry Credits
	e.Credits, err = buf.PopUInt8()
	if err != nil {
		return nil, err
	}

	// 32 byte Public Key
	err = buf.PopBinaryMarshallable(e.ECPubKey)
	if err != nil {
		return nil, err
	}

	// 64 byte Signature
	err = buf.PopBinaryMarshallable(e.Sig)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this object
func (e *CommitEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *CommitEntry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

// JSONString returns this object as a json encoded string
func (e *CommitEntry) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}
