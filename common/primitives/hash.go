// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives/random"
)

type Hash [constants.HASH_LENGTH]byte

var _ interfaces.Printable = (*Hash)(nil)
var _ interfaces.IHash = (*Hash)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*Hash)(nil)
var _ encoding.TextMarshaler = (*Hash)(nil)

func RandomHash() interfaces.IHash {
	h := random.RandByteSliceOfLen(constants.HASH_LENGTH)
	answer := new(Hash)
	answer.SetBytes(h)
	return answer
}

func (c *Hash) Copy() interfaces.IHash {
	h := new(Hash)
	err := h.SetBytes(c.Bytes())
	if err != nil {
		panic(err)
	}
	return h
}

func (c *Hash) New() interfaces.BinaryMarshallableAndCopyable {
	return new(Hash)
}

func (h *Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

func (h *Hash) IsZero() bool {
	return h.String() == "0000000000000000000000000000000000000000000000000000000000000000"
}

// NewShaHashFromStr creates a ShaHash from a hash string.  The string should be
// the hexadecimal string of a byte-reversed hash, but any missing characters
// result in zero padding at the end of the ShaHash.
func NewShaHashFromStr(hash string) (*Hash, error) {
	h := new(Hash)
	err := h.UnmarshalText([]byte(hash))
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Hash) UnmarshalText(b []byte) error {
	p, err := hex.DecodeString(string(b))
	if err != nil {
		return err
	}
	copy(h[:], p)
	return nil
}

func (h *Hash) Fixed() [constants.HASH_LENGTH]byte {
  // Might change the error produced by IHash in FD-398
	if h == nil {
		panic("nil Hash")
	}
	return *h
}

func (h *Hash) Bytes() []byte {
	return h.GetBytes()
}

func (Hash) GetHash() interfaces.IHash {
	return nil
}

func CreateHash(entities ...interfaces.BinaryMarshallable) (h interfaces.IHash, err error) {
	sha := sha256.New()
	h = new(Hash)
	for _, entity := range entities {
		data, err := entity.MarshalBinary()
		if err != nil {
			return nil, err
		}
		sha.Write(data)
	}
	h.SetBytes(sha.Sum(nil))
	return
}

func (h *Hash) MarshalBinary() ([]byte, error) {
	return h.Bytes(), nil
}

func (h *Hash) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	copy(h[:], p)
	newData = p[constants.HASH_LENGTH:]
	return
}

func (h *Hash) UnmarshalBinary(p []byte) (err error) {
	_, err = h.UnmarshalBinaryData(p)
	return
}

// Make a copy of the hash in this hash.  Changes to the return value WILL NOT be
// reflected in the source hash.  You have to do a SetBytes to change the source
// value.
func (h *Hash) GetBytes() []byte {
	newHash := make([]byte, constants.HASH_LENGTH)
	copy(newHash, h[:])

	return newHash
}

// SetBytes sets the bytes which represent the hash.  An error is returned if
// the number of bytes passed in is not constants.HASH_LENGTH.
func (hash *Hash) SetBytes(newHash []byte) error {
	nhlen := len(newHash)
	if nhlen != constants.HASH_LENGTH {
		return fmt.Errorf("invalid sha length of %v, want %v", nhlen, constants.HASH_LENGTH)
	}
	copy(hash[:], newHash)
	return nil
}

// NewShaHash returns a new ShaHash from a byte slice.  An error is returned if
// the number of bytes passed in is not constants.HASH_LENGTH.
func NewShaHash(newHash []byte) (*Hash, error) {
	var sh Hash
	err := sh.SetBytes(newHash)
	if err != nil {
		return nil, err
	}
	return &sh, err
}

// Create a Sha512[:256] Hash from a byte array
func Sha512Half(p []byte) (h *Hash) {
	sha := sha512.New()
	sha.Write(p)

	h = new(Hash)
	copy(h[:], sha.Sum(nil)[:constants.HASH_LENGTH])
	return h
}

// Convert a hash into a string with hex encoding
func (h *Hash) String() string {
	if h == nil {
		return hex.EncodeToString(nil)
	} else {
		return hex.EncodeToString(h[:])
	}
}

func (h *Hash) ByteString() string {
	return string(h[:])
}

func HexToHash(hexStr string) (h interfaces.IHash, err error) {
	h = new(Hash)
	v, err := hex.DecodeString(hexStr)
	err = h.SetBytes(v)
	return h, err
}

// Compare two Hashes
func (a *Hash) IsSameAs(b interfaces.IHash) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}

	if bytes.Compare(a[:], b.Bytes()) == 0 {
		return true
	}

	return false
}

// Is the hash a minute marker (the last byte indicates the minute number)
func (h *Hash) IsMinuteMarker() bool {
	if bytes.Equal(h[:constants.HASH_LENGTH-1], constants.ZERO_HASH[:constants.HASH_LENGTH-1]) {
		return true
	}

	return false
}

func (h *Hash) ToMinute() byte {
	return h[constants.HASH_LENGTH-1]
}

func (e *Hash) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *Hash) JSONString() (string, error) {
	return EncodeJSONString(e)
}

/**********************
 * Support functions
 **********************/

// Create a Sha256 Hash from a byte array
func Sha(p []byte) interfaces.IHash {
	h := new(Hash)
	b := sha256.Sum256(p)
	h.SetBytes(b[:])
	return h
}

// Shad Double Sha256 Hash; sha256(sha256(data))
func Shad(data []byte) interfaces.IHash {
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	h := new(Hash)
	h.SetBytes(h2[:])
	return h
}

func NewZeroHash() interfaces.IHash {
	h := new(Hash)
	return h
}

func NewHash(b []byte) interfaces.IHash {
	h := new(Hash)
	h.SetBytes(b)
	return h
}

// shad Double Sha256 Hash; sha256(sha256(data))
func DoubleSha(data []byte) []byte {
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	return h2[:]
}

func NewShaHashFromStruct(DataStruct interface{}) (interfaces.IHash, error) {
	jsonbytes, err := json.Marshal(DataStruct)
	if err != nil {
		return nil, err
	}

	return NewShaHash(DoubleSha(jsonbytes))
}
