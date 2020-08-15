// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"github.com/PaulSnow/factom2d/common/primitives"
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
	return primitives.NewJSONError(-32700, "Parse error", nil)
}
func NewInvalidRequestError() *primitives.JSONError {
	return primitives.NewJSONError(-32600, "Invalid Request", nil)
}
func NewMethodNotFoundError() *primitives.JSONError {
	return primitives.NewJSONError(-32601, "Method not found", nil)
}
func NewInvalidParamsError() *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params", nil)
}
func NewInternalError() *primitives.JSONError {
	return primitives.NewJSONError(-32603, "Internal error", nil)
}

/*******************************************************************/
func NewCustomInternalError(data interface{}) *primitives.JSONError {
	return primitives.NewJSONError(-32603, "Internal error", data)
}
func NewCustomInvalidParamsError(data interface{}) *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params", data)
}

/*******************************************************************/

func NewInvalidAddressError() *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params", "Invalid Address")
}
func NewUnableToDecodeTransactionError() *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params", "Unable to decode the transaction")
}
func NewInvalidTransactionError() *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params", "Invalid Transaction")
}
func NewInvalidHashError() *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params", "Invalid Hash")
}
func NewInvalidEntryError() *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params", "Invalid Entry")
}
func NewInvalidCommitChainError() *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params", "Invalid Commit Chain")
}
func NewInvalidCommitEntryError() *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params", "Invalid Commit Entry")
}
func NewInvalidDataPassedError() *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params", "Invalid data passed")
}
func NewInvalidHeightError() *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params", "Invalid Block Height passed")
}
func NewInvertedHeightError() *primitives.JSONError {
	return primitives.NewJSONError(-32602, "Invalid params", "Starting block height must be less than ending block height")
}
func NewInternalDatabaseError() *primitives.JSONError {
	return primitives.NewJSONError(-32603, "Internal error", "database error")
}

//http://www.jsonrpc.org/specification : -32000 to -32099 error codes are reserved for implementation-defined server-errors.
func NewBlockNotFoundError() *primitives.JSONError {
	return primitives.NewJSONError(-32008, "Block not found", nil)
}
func NewEntryNotFoundError() *primitives.JSONError {
	return primitives.NewJSONError(-32008, "Entry not found", nil)
}
func NewObjectNotFoundError() *primitives.JSONError {
	return primitives.NewJSONError(-32008, "Object not found", nil)
}
func NewMissingChainHeadError() *primitives.JSONError {
	return primitives.NewJSONError(-32009, "Missing Chain Head", nil)
}
func NewReceiptError() *primitives.JSONError {
	return primitives.NewJSONError(-32010, "Receipt creation error", nil)
}
func NewRepeatCommitError(data interface{}) *primitives.JSONError {
	return primitives.NewJSONError(-32011, "Repeated Commit", data)
}
