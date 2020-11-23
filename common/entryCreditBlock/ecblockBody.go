// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// ECBlockBody contains all the entry information for the EC block. The entries will be one of 5 types:
// CommitChain, CommitEntry, IncreaseBalance, MinuteNumber, or ServerIndexNumber
type ECBlockBody struct {
	Entries []interfaces.IECBlockEntry `json:"entries"` // The EC block entries
}

var _ interfaces.Printable = (*ECBlockBody)(nil)
var _ interfaces.IECBlockBody = (*ECBlockBody)(nil)

// IsSameAs returns true iff the input object is the identical to this object
func (e *ECBlockBody) IsSameAs(b interfaces.IECBlockBody) bool {
	if e == nil || b == nil {
		if e == nil && b == nil {
			return true
		}
		return false
	}

	bEntries := b.GetEntries()
	if len(e.Entries) != len(bEntries) {
		return false
	}
	for i := range e.Entries {
		if e.Entries[i].IsSameAs(bEntries[i]) == false {
			return false
		}
	}

	return true
}

// GetEntries returns the list of entries
func (e *ECBlockBody) GetEntries() []interfaces.IECBlockEntry {
	return e.Entries
}

// AddEntry appends the input entry to the internal list of entries
func (e *ECBlockBody) AddEntry(entry interfaces.IECBlockEntry) {
	e.Entries = append(e.Entries, entry)
}

// SetEntries sets the internal entry list to the input list
func (e *ECBlockBody) SetEntries(entries []interfaces.IECBlockEntry) {
	e.Entries = entries
}

// JSONByte returns the json encoded byte array
func (e *ECBlockBody) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *ECBlockBody) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

// String returns this object as a string
func (e *ECBlockBody) String() string {
	var out primitives.Buffer
	for _, v := range e.Entries {
		out.WriteString(v.String())
	}
	return string(out.DeepCopyBytes())
}

/*******************************************************
 * Support Functions
 *******************************************************/

// NewECBlockBody creates a new, empty EC block body
func NewECBlockBody() interfaces.IECBlockBody {
	b := new(ECBlockBody)
	b.Entries = make([]interfaces.IECBlockEntry, 0)
	return b
}
