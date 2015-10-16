// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package confirmation

import (
)

type IConfirmation Interface {
	BinaryMarshallable
	Printable
	
	Type()			int			// Confirmation Type
	DBHeight()		int			// Directory Block Height
	ChainID[]		[]byte		// ChainID of the sending server
	ListHeight()	int			// Height in the Process List
	Value()			[]byte		// Value.  Different for each confirmation type
	SerialHash()	[]byte		// Serial Hash so far
	Signature()		[]byte		// Signature
}