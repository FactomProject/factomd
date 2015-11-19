// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"bytes"
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

func (e *ECBlockBody) GetEntries() ([]interfaces.IECBlockEntry) {
	return e.Entries
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

func (e *ECBlockBody) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *ECBlockBody) String() string {
	str, _ := e.JSONString()
	return str
}

/*******************************************************
 * Support Functions
 *******************************************************/

func NewECBlockBody() interfaces.IECBlockBody {
	b := new(ECBlockBody)
	b.Entries = make([]interfaces.IECBlockEntry, 0)
	return b
}



