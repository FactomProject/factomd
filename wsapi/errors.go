// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"github.com/FactomProject/factomd/common/primitives"
)

/*
The error codes from and including -32768 to -32000 are reserved for pre-defined errors.
Any code within this range, but not defined explicitly below is reserved for future use.
The error codes are nearly the same as those suggested for XML-RPC at the following url:
http://xmlrpc-epi.sourceforge.net/specs/rfc.fault_codes.php

code				message						meaning
-32700				Parse error					Invalid JSON was received by the server.
												An error occurred on the server while parsing the JSON text.
-32600				Invalid Request				The JSON sent is not a valid Request object.
-32601				Method not found			The method does not exist / is not available.
-32602				Invalid params				Invalid method parameter(s).
-32603				Internal error				Internal JSON-RPC error.
-32000 to -32099	Server error				Reserved for implementation-defined server-errors.
*/

func NewParseError() *primitives.JSONError {
	return primitives.NewJSONError(-32700, "Parse error")
}
func NewInvalidRequestError() *primitives.JSONError {
	return primitives.NewJSONError(-32600, "Invalid Request")
}
func NewInvalidRequestError() *primitives.JSONError {
	return primitives.NewJSONError(-32601, "Method not found")
}
func NewInvalidRequestError() *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params")
}
func NewInvalidRequestError() *primitives.JSONError {
	return primitives.NewJSONError(-32603, "Internal error")
}