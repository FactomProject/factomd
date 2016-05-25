// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"io"
	"time"

	ed "github.com/FactomProject/ed25519"
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
}

var _ interfaces.Printable = (*CommitEntry)(nil)
var _ interfaces.BinaryMarshallable = (*CommitEntry)(nil)
var _ interfaces.ShortInterpretable = (*CommitEntry)(nil)
var _ interfaces.IECBlockEntry = (*CommitEntry)(nil)
var _ interfaces.ISignable = (*CommitEntry)(nil)

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

// Return the timestamp in milliseconds.
func (c *CommitEntry) GetMilliTime() int64 {
	a := make([]byte, 2, 8)
	a = append(a, c.MilliTime[:]...)
	milli := int64(binary.BigEndian.Uint64(a))
	return milli
}

// InTime checks the CommitEntry.MilliTime and returns true if the timestamp is
// whitin +/- 12 hours of the current time.
func (c *CommitEntry) InTime() bool {
	now := time.Now()
	sec := c.GetMilliTime() / 1000
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
	data := c.CommitMsg()
	return primitives.Sha(data)
}

func (c *CommitEntry) MarshalBinarySig() ([]byte, error) {
	buf := new(primitives.Buffer)

	// 1 byte Version
	if err := binary.Write(buf, binary.BigEndian, c.Version); err != nil {
		return nil, err
	}

	// 6 byte MilliTime
	buf.Write(c.MilliTime[:])

	// 32 byte Entry Hash
	buf.Write(c.EntryHash.Bytes())

	// 1 byte number of Entry Credits
	if err := binary.Write(buf, binary.BigEndian, c.Credits); err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil

}

func (c *CommitEntry) MarshalBinary() ([]byte, error) {
	buf := new(primitives.Buffer)

	b, err := c.MarshalBinarySig()
	if err != nil {
		return nil, err
	}

	buf.Write(b)

	// 32 byte Public Key
	buf.Write(c.ECPubKey[:])

	// 64 byte Signature
	buf.Write(c.Sig[:])

	return buf.DeepCopyBytes(), nil
}

func (c *CommitEntry) Sign(privateKey []byte) error {
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

func (c *CommitEntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	buf := primitives.NewBuffer(data)
	hash := make([]byte, 32)

	var b byte
	var p []byte
	// 1 byte Version
	if b, err = buf.ReadByte(); err != nil {
		return
	} else {
		c.Version = uint8(b)
	}

	if buf.Len() < 6 {
		err = io.EOF
		return
	}

	// 6 byte MilliTime
	if p = buf.Next(6); p == nil {
		err = fmt.Errorf("Could not read MilliTime")
		return
	} else {
		copy(c.MilliTime[:], p)
	}

	// 32 byte Entry Hash
	if _, err = buf.Read(hash); err != nil {
		return
	} else if err = c.EntryHash.SetBytes(hash); err != nil {
		return
	}

	// 1 byte number of Entry Credits
	if b, err = buf.ReadByte(); err != nil {
		return
	} else {
		c.Credits = uint8(b)
	}

	if buf.Len() < 32 {
		err = io.EOF
		return
	}

	// 32 byte Public Key
	if p = buf.Next(32); p == nil {
		err = fmt.Errorf("Could not read ECPubKey")
		return
	} else {
		copy(c.ECPubKey[:], p)
	}

	if buf.Len() < 64 {
		err = io.EOF
		return
	}

	// 64 byte Signature
	if p = buf.Next(64); p == nil {
		err = fmt.Errorf("Could not read Sig")
		return
	} else {
		copy(c.Sig[:], p)
	}

	newData = buf.DeepCopyBytes()

	return
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

func (e *CommitEntry) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *CommitEntry) String() string {
	str, _ := e.JSONString()
	return str
}
