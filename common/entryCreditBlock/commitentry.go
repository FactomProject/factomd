// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"encoding/binary"
	"fmt"
	"time"

	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const (
	// CommitEntrySize = 1 + 6 + 32 + 1 + 32 + 64
	CommitEntrySize int = 136
)

type CommitEntry struct {
	Version   uint8
	MilliTime *primitives.ByteSlice6
	EntryHash interfaces.IHash
	Credits   uint8
	ECPubKey  *primitives.ByteSlice32
	Sig       *primitives.ByteSlice64
	SigHash   interfaces.IHash
}

var _ interfaces.Printable = (*CommitEntry)(nil)
var _ interfaces.BinaryMarshallable = (*CommitEntry)(nil)
var _ interfaces.ShortInterpretable = (*CommitEntry)(nil)
var _ interfaces.IECBlockEntry = (*CommitEntry)(nil)
var _ interfaces.ISignable = (*CommitEntry)(nil)

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

func (e *CommitEntry) String() string {
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf(" %-20s\n", "CommitEntry"))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "Version", e.Version))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "MilliTime", e.MilliTime))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "EntryHash", e.EntryHash.Bytes()[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "Credits", e.Credits))
	out.WriteString(fmt.Sprintf("   %-20s %x\n", "ECPubKey", e.ECPubKey[:3]))
	out.WriteString(fmt.Sprintf("   %-20s %d\n", "Sig", e.Sig[:3]))

	return (string)(out.DeepCopyBytes())
}

func (a *CommitEntry) GetEntryHash() interfaces.IHash {
	return a.EntryHash
}

func (a *CommitEntry) IsSameAs(b *CommitEntry) bool {
	if b == nil {
		return false
	}
	bin1, err := a.MarshalBinary()
	if err != nil {
		return false
	}
	bin2, err := b.MarshalBinary()
	if err != nil {
		return false
	}
	return primitives.AreBytesEqual(bin1, bin2)
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

func (e *CommitEntry) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

func (b *CommitEntry) IsInterpretable() bool {
	return false
}

func (b *CommitEntry) Interpret() string {
	return ""
}

// CommitMsg returns the binary marshaled message section of the CommitEntry
// that is covered by the CommitEntry.Sig.
func (c *CommitEntry) CommitMsg() []byte {
	p, err := c.MarshalBinary()
	if err != nil {
		return []byte{byte(0)}
	}
	return p[:len(p)-64-32]
}

// Return the timestamp
func (c *CommitEntry) GetTimestamp() interfaces.Timestamp {
	a := make([]byte, 2, 8)
	a = append(a, c.MilliTime[:]...)
	milli := uint64(binary.BigEndian.Uint64(a))
	return primitives.NewTimestampFromMilliseconds(milli)
}

// InTime checks the CommitEntry.MilliTime and returns true if the timestamp is
// whitin +/- 12 hours of the current time.
func (c *CommitEntry) InTime() bool {
	now := time.Now()
	sec := c.GetTimestamp().GetTimeSeconds()
	t := time.Unix(sec, 0)

	return t.After(now.Add(-constants.COMMIT_TIME_WINDOW*time.Hour)) && t.Before(now.Add(constants.COMMIT_TIME_WINDOW*time.Hour))
}

func (c *CommitEntry) IsValid() bool {
	//double check the credits in the commit
	if c.Credits < 1 || c.Version != 0 {
		return false
	}
	return ed.VerifyCanonical((*[32]byte)(c.ECPubKey), c.CommitMsg(), (*[64]byte)(c.Sig))
}

func (c *CommitEntry) GetHash() interfaces.IHash {
	h, _ := c.MarshalBinary()
	return primitives.Sha(h)
}

func (c *CommitEntry) GetSigHash() interfaces.IHash {
	if c.SigHash == nil {
		data := c.CommitMsg()
		c.SigHash = primitives.Sha(data)
	}
	return c.SigHash
}

func (c *CommitEntry) MarshalBinarySig() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	// 1 byte Version
	err := buf.PushUInt8(c.Version)
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
func (c *CommitEntry) MarshalBinaryTransaction() ([]byte, error) {
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

func (c *CommitEntry) MarshalBinary() ([]byte, error) {
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
	return ECIDEntryCommit
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

func (e *CommitEntry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommitEntry) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}
