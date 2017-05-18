// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type ECBlockBody struct {
	Entries []interfaces.IECBlockEntry
}

var _ = fmt.Print
var _ interfaces.Printable = (*ECBlockBody)(nil)
var _ interfaces.IECBlockBody = (*ECBlockBody)(nil)

func (a *ECBlockBody) IsSameAs(b interfaces.IECBlockBody) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}

	bEntries := b.GetEntries()
	if len(a.Entries) != len(bEntries) {
		return false
	}
	for i := range a.Entries {
		if a.Entries[i].IsSameAs(bEntries[i]) == false {
			return false
		}
	}

	return true
}

func (e *ECBlockBody) GetEntries() []interfaces.IECBlockEntry {
	return e.Entries
}

func (e *ECBlockBody) AddEntry(entry interfaces.IECBlockEntry) {
	e.Entries = append(e.Entries, entry)
}

func (e *ECBlockBody) SetEntries(entries []interfaces.IECBlockEntry) {
	e.Entries = entries
}

func (e *ECBlockBody) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ECBlockBody) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

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

func NewECBlockBody() interfaces.IECBlockBody {
	b := new(ECBlockBody)
	b.Entries = make([]interfaces.IECBlockEntry, 0)
	return b
}
