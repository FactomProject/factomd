// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package simplecoin

import (
    "bytes"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
)

type IHash interface {
    IBlock                                          // Implements IBlock
    
    Bytes() []byte                                  // Return the byte slice for this Hash
    SetBytes([]byte) error                          // Set the bytes 
    String() string                                 // Convert to a String ???
    IsSameAs(IHash) bool                            // Compare two Hashes
    CreateHash(a ...IBlock) (IHash, error)          // Create a serial Hash from arguments
    HexToHash(hexStr string) (IHash, error)         // Convert a Hex string to a Hash
}

type Hash struct {
    IHash
    bytes [ADDRESS_LENGTH]byte 
}

func (h Hash) NewBlock() IBlock {
    h2 := new(Hash)
    return h2
}

// Sum these Hashes
func (hash Hash) CreateHash(entities ...IBlock) (h2 IHash, err error) {
    sha := sha256.New()
    h := new(Hash)
    for _, entity := range entities {
        data, err := entity.MarshalBinary()
        if err != nil {
            return nil, err
        }
        sha.Write(data)
    }
    copy(h.bytes[:],sha.Sum(nil))
    return
}

func (h Hash) MarshalBinary() ([]byte, error) {
    var buf bytes.Buffer
    buf.Write(h.bytes[:])
    return buf.Bytes(), nil
}

func (h Hash) UnmarshalBinary(p []byte) error {
    h.bytes = *new([32]byte)	
    copy(h.bytes[:], p)
    return nil
}

// Make a copy of the bytes in this hash.  Changes to the return value WILL NOT be 
// reflected in the source hash.  You have to do a SetBytes to change the source
// value.
func (h Hash) Bytes() []byte {
    newHash := make([]byte, ADDRESS_LENGTH)
    copy(newHash, h.bytes[:])
    
    return newHash
}

// SetBytes sets the bytes which represent the hash.  An error is returned if
// the number of bytes passed in is not ADDRESS_LENGTH.
func (hash *Hash) SetBytes(newHash []byte) error {
    nhlen := len(newHash)
    if nhlen != ADDRESS_LENGTH {
        return fmt.Errorf("invalid sha length of %v, want %v", nhlen, ADDRESS_LENGTH)
    }
    
    hash.bytes = *new([32]byte)
    copy(hash.bytes[:], newHash)
    return nil
}


// Create a Sha256 Hash from a byte array
func Sha(p []byte) (h2 IHash) {
    sha := sha256.New()
    sha.Write(p)
    
    h := new(Hash)
    h.SetBytes(sha.Sum(nil))
    return h
}

// Convert a hash into a string with hex encoding
func (h Hash) String() string {
    return hex.EncodeToString(h.bytes[:])
}


func (hash Hash)HexToHash(hexStr string) (h2 IHash, err error) {
    h := new(Hash)
    byt, err := hex.DecodeString(hexStr)
    copy(h.bytes[:],byt)
    return h, err
}

// Compare two Hashes
func (a Hash) IsSameAs(b IHash) bool {
    if  b == nil {
        return false
    }
    
    if bytes.Compare(a.bytes[:], b.Bytes()) == 0 {
        return true
    }
    
    return false
}
