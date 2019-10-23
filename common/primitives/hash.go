// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/util/atomic"

	llog "github.com/FactomProject/factomd/log"
)

// Hash is a convenient fixed []byte type created at the hash length
type Hash [constants.HASH_LENGTH]byte

// The various interfaces Hash will be a part of, and must implement below
var _ interfaces.Printable = (*Hash)(nil)
var _ interfaces.IHash = (*Hash)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*Hash)(nil)
var _ encoding.TextMarshaler = (*Hash)(nil)

// ZeroHash is a zero hash
var ZeroHash interfaces.IHash = NewHash(constants.ZERO_HASH)

var noRepeat map[string]int = make(map[string]int)

func LogNilHashBug(msg string) {
	whereAmI := atomic.WhereAmIString(2)
	noRepeat[whereAmI]++

	if noRepeat[whereAmI]%100 == 1 {
		fmt.Fprintf(os.Stderr, "%s. Called from %s\n", msg, whereAmI)
	}

}

// IsHashNil returns true if receiver is nil, or the hash is zero
func (h *Hash) IsHashNil() bool {
	return h == nil || reflect.ValueOf(h).IsNil()
}

// RandomHash returns a new random hash
func RandomHash() interfaces.IHash {
	h := random.RandByteSliceOfLen(constants.HASH_LENGTH)
	answer := NewHash(h)
	return answer
}

// Copy returns a copy of this Hash
func (h *Hash) Copy() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			LogNilHashBug("Hash.Copy() saw an interface that was nil")
		}
	}()
	nh := new(Hash)
	err := nh.SetBytes(h.Bytes())
	if err != nil {
		panic(err)
	}
	return nh
}

// New creates a new Hash (required for BinarymarshallableAndCopyable interface)
func (h *Hash) New() interfaces.BinaryMarshallableAndCopyable {
	return new(Hash)
}

// MarshalText marshals the Hash as text
func (h *Hash) MarshalText() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Hash.MarshalText err:%v", *pe)
		}
	}(&err)
	return []byte(hex.EncodeToString(h[:])), nil
}

// IsZero returns true iff Hash is zero
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

// UnmarshalText unmarshals the input array
func (h *Hash) UnmarshalText(b []byte) error {
	p, err := hex.DecodeString(string(b))
	if err != nil {
		return err
	}
	copy(h[:], p)
	return nil
}

// Fixed returns the fixed []byte array
func (h *Hash) Fixed() [constants.HASH_LENGTH]byte {
	// Might change the error produced by IHash in FD-398
	if h == nil {
		panic("nil Hash")
	}
	return *h
}

// PFixed returns a pointer to the fixed []byte array
func (h *Hash) PFixed() *[constants.HASH_LENGTH]byte {
	// Might change the error produced by IHash in FD-398
	return (*[constants.HASH_LENGTH]byte)(h)
}

// Bytes returns a copy of the internal []byte array
func (h *Hash) Bytes() (rval []byte) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "nil hash")
			llog.LogPrintf("recovery", "nil hash")
			debug.PrintStack()
			rval = constants.ZERO_HASH
		}
	}()
	return h.GetBytes()
}

// GetHash is unused, merely here to implement the IHash interface
func (Hash) GetHash() interfaces.IHash {
	return nil
}

// CreateHash returns a hash created from all input entities
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

// MarshalBinary returns a copy of the []byte array
func (h *Hash) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Hash.MarshalBinary err:%v", *pe)
		}
	}(&err)
	return h.Bytes(), nil
}

// UnmarshalBinaryData unmarshals the input array into the Hash
func (h *Hash) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
		}
	}()
	copy(h[:], p)
	newData = p[constants.HASH_LENGTH:]
	return
}

// UnmarshalBinary unmarshals the input array into the Hash
func (h *Hash) UnmarshalBinary(p []byte) (err error) {
	_, err = h.UnmarshalBinaryData(p)
	return
}

// GetBytes makes a copy of the hash in this hash.  Changes to the return value WILL NOT be
// reflected in the source hash.  You have to do a SetBytes to change the source
// value.
func (h *Hash) GetBytes() []byte {
	newHash := make([]byte, constants.HASH_LENGTH)
	copy(newHash, h[:])

	return newHash
}

// SetBytes sets the bytes which represent the hash.  An error is returned if
// the number of bytes passed in is not constants.HASH_LENGTH.
func (h *Hash) SetBytes(newHash []byte) error {
	nhlen := len(newHash)
	if nhlen != constants.HASH_LENGTH {
		return fmt.Errorf("invalid sha length of %v, want %v", nhlen, constants.HASH_LENGTH)
	}
	copy(h[:], newHash)
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

// Sha512Half creates a Sha512[:256] Hash from a byte array
func Sha512Half(p []byte) (h *Hash) {
	sha := sha512.New()
	sha.Write(p)

	h = new(Hash)
	copy(h[:], sha.Sum(nil)[:constants.HASH_LENGTH])
	return h
}

// String converts a hash into a hexidecimal (0-F) string
func (h *Hash) String() string {
	if h == nil {
		return hex.EncodeToString(nil)
	} else {
		return hex.EncodeToString(h[:])
	}
}

// ByteString returns the hash as a byte string
func (h *Hash) ByteString() string {
	return string(h[:])
}

// HexToHash converts the input hexidecimal (0-F) string into the internal []byte array
func HexToHash(hexStr string) (h interfaces.IHash, err error) {
	h = new(Hash)
	v, err := hex.DecodeString(hexStr)
	err = h.SetBytes(v)
	return h, err
}

// IsSameAs compares two Hashes and returns true iff they hashs are binary identical
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

// IsMinuteMarker checks if the Hash is the hash of a minute marker (the last byte indicates the minute number)
// A minute marker is determined by having all but the last value as zero
func (h *Hash) IsMinuteMarker() bool {
	if bytes.Equal(h[:constants.HASH_LENGTH-1], constants.ZERO_HASH[:constants.HASH_LENGTH-1]) {
		return true
	}

	return false
}

// ToMinute returns the last byte of the Hash
func (h *Hash) ToMinute() byte {
	return h[constants.HASH_LENGTH-1]
}

// JSONByte returns the json encoded []byte array of the Hash
func (h *Hash) JSONByte() ([]byte, error) {
	return EncodeJSON(h)
}

// JSONString returns the json encoded byte string of the Hash
func (h *Hash) JSONString() (string, error) {
	return EncodeJSONString(h)
}

// Loghashfixed logs the input hash
/****************************************************************
	DEBUG logging to keep full hash. Turned on from command line
 ****************************************************************/
func Loghashfixed(h [32]byte) {
	if !globals.Params.FullHashesLog {
		return
	}
	if globals.Hashlog == nil {
		f, err := os.Create("fullhashes.txt")
		globals.Hashlog = bufio.NewWriter(f)
		f.WriteString(time.Now().String() + "\n")
		if err != nil {
			panic(err)
		}
	}
	globals.HashMutex.Lock()
	defer globals.HashMutex.Unlock()
	if globals.Hashes == nil {
		globals.Hashes = make(map[[32]byte]bool)
	}
	_, exists := globals.Hashes[h]
	if !exists {
		//fmt.Fprintf(globals.Hashlog, "%x\n", h)
		var x int
		// turns out random is better than LRU because the leader/common chain hashes get used a lot and keep getting
		// tossed. Probably better to add special handling for leader and known chains ...

		if true {
			x = globals.HashNext // Use LRU
		} else {
			x = random.RandIntBetween(0, len(globals.HashesInOrder)) // use random replacement
		}
		if globals.HashesInOrder[x] != nil {
			fmt.Fprintf(globals.Hashlog, "delete [%4d] %x\n", x, *globals.HashesInOrder[x])
			delete(globals.Hashes, *globals.HashesInOrder[x]) // delete the oldest hash
			globals.HashesInOrder[x] = nil
		}
		fmt.Fprintf(globals.Hashlog, "add    [%4d] %x\n", x, h)
		globals.Hashes[h] = true                                               // add the new hash
		globals.HashesInOrder[x] = &h                                          // add it to the ordered list
		globals.HashNext = (globals.HashNext + 1) % len(globals.HashesInOrder) // wrap index at end of array
	}
}

// Loghash logs the input interface
func Loghash(h interfaces.IHash) {
	if h == nil {
		return
	}
	Loghashfixed(h.Fixed())
}

/**********************
 * Support functions
 **********************/

// Sha creates a Sha256 Hash from a byte array
func Sha(p []byte) interfaces.IHash {
	h := new(Hash)
	b := sha256.Sum256(p)
	h.SetBytes(b[:])
	Loghash(h)
	return h
}

// Shad returns a new hash created by double hashing the input: Double Sha256 Hash; sha256(sha256(data))
func Shad(data []byte) interfaces.IHash {
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	h := new(Hash)
	h.SetBytes(h2[:])
	return h
}

// NewZeroHash creates a new zero hash object
func NewZeroHash() interfaces.IHash {
	h := new(Hash)
	return h
}

// NewHash creates a new object for the input hash
func NewHash(b []byte) interfaces.IHash {
	h := new(Hash)
	h.SetBytes(b)
	return h
}

// DoubleSha returns a new hash created by double hashing the input: shad Double Sha256 Hash; sha256(sha256(data))
func DoubleSha(data []byte) []byte {
	h1 := sha256.Sum256(data)
	h2 := sha256.Sum256(h1[:])
	Loghashfixed(h2)
	return h2[:]
}

// NewShaHashFromStruct marshals the input struct into a json byte array, then double hashes the json array
func NewShaHashFromStruct(DataStruct interface{}) (interfaces.IHash, error) {
	jsonbytes, err := json.Marshal(DataStruct)
	if err != nil {
		return nil, err
	}

	return NewShaHash(DoubleSha(jsonbytes))
}
