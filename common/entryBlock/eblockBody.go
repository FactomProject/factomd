// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryBlock

import (
	"encoding/hex"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// EBlockBody is the series of Hashes that form the Entry Block Body.
type EBlockBody struct {
	EBEntries []interfaces.IHash `json:"ebentries"` // Array of entries from a single chain id associated with this entry block
}

var _ interfaces.Printable = (*EBlockBody)(nil)
var _ interfaces.IEBlockBody = (*EBlockBody)(nil)

// IsSameAs returns true iff th einput object is the same as this object
func (e *EBlockBody) IsSameAs(b interfaces.IEBlockBody) bool {
	if e == nil || b == nil {
		if e == nil && b == nil {
			return true
		}
		return false
	}

	bEBEntries := b.GetEBEntries()
	if len(e.EBEntries) != len(bEBEntries) {
		return false
	}
	for i := range e.EBEntries {
		if e.EBEntries[i].IsSameAs(bEBEntries[i]) == false {
			return false
		}
	}

	return true
}

// NewEBlockBody initializes an empty Entry Block Body.
func NewEBlockBody() *EBlockBody {
	e := new(EBlockBody)
	e.EBEntries = make([]interfaces.IHash, 0)
	return e
}

// MR calculates the Merkle Root of the Entry Block Body. See func
// primitives.BuildMerkleTreeStore(hashes []interfaces.IHash) (merkles []interfaces.IHash) in common/merkle.go.
func (e *EBlockBody) MR() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "EBlockBody.MR") }()

	mrs := primitives.BuildMerkleTreeStore(e.EBEntries)
	r := mrs[len(mrs)-1] // Remember, Merkle root is the last of the hashes
	return r
}

// JSONByte returns the json encoded byte array
func (e *EBlockBody) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded byte string
func (e *EBlockBody) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

// String returns this object as a string
func (e *EBlockBody) String() string {
	var out primitives.Buffer
	for _, eh := range e.EBEntries {
		out.WriteString(fmt.Sprintf("    %20s: %x\n", "Entry Hash", eh.Bytes()[:3]))
	}
	return (string)(out.DeepCopyBytes())
}

// GetEBEntries returns the hash array associated with the entry block
func (e *EBlockBody) GetEBEntries() []interfaces.IHash {
	return e.EBEntries[:]
}

// AddEBEntry creates a new Entry Block Entry from the provided Factom Entry
// and adds it to the Entry Block Body.
func (e *EBlockBody) AddEBEntry(entry interfaces.IHash) {
	e.EBEntries = append(e.EBEntries, entry)
}

// AddEndOfMinuteMarker adds the End of Minute to the Entry Block. The End of
// Minut byte becomes the last byte in a 32 byte slice that is added to the
// Entry Block Body as an Entry Block Entry.
func (e *EBlockBody) AddEndOfMinuteMarker(m byte) {
	// create a map of possible minute markers that may be found in the
	// EBlock Body
	mins := make(map[string]uint8)
	for i := byte(1); i <= 10; i++ {
		h := make([]byte, 32)
		h[len(h)-1] = i
		mins[hex.EncodeToString(h)] = i
	}

	// check if the previous entry is a minute marker and return without
	// writing if it is
	prevEntry := e.EBEntries[len(e.EBEntries)-1]
	if _, exist := mins[prevEntry.String()]; exist {
		return
	}

	h := make([]byte, 32)
	h[len(h)-1] = m
	hash := primitives.NewZeroHash()
	hash.SetBytes(h)

	e.AddEBEntry(hash)
}
