// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"encoding/json"
	"fmt"
)

// JSON2Request is an JSON RPC request object used in factomd according to the specification for JSON RPC 2.0
// See specification: https://www.jsonrpc.org/specification
// Remote Proceedure Call (RPC) is a protocol for executing subroutines on a remote
type JSON2Request struct {
	JSONRPC string      `json:"jsonrpc"` // version string which MUST be "2.0" (version 1.0 didn't have this field)
	ID      interface{} `json:"id"`      // Unique client defined ID associated with this request. It may be a number,
	// string, or nil. It is used by the remote to formulate a response object
	// containing the same ID back to the client. If nil, remote treats request object
	// as a notification, and that the client does not expect a response object back
	Params interface{} `json:"params,omitempty"` // input parameters for the subroutine called on the remote - may be omitted
	Method string      `json:"method,omitempty"` // name of the subroutine to be called on the remote
}

// JSONByte encodes object to a json []byte
func (j *JSON2Request) JSONByte() ([]byte, error) {
	return EncodeJSON(j)
}

// JSONString encodes object to a json string with error return
func (j *JSON2Request) JSONString() (string, error) {
	return EncodeJSONString(j)
}

// String encodes object to a json string with no error return
func (j *JSON2Request) String() string {
	str, _ := j.JSONString()
	return str
}

// NewJSON2RequestBlank creates a new blank JSON2Request object with only the JSONRPC field filled (because it has 1 acceptable value)
func NewJSON2RequestBlank() *JSON2Request {
	j := new(JSON2Request)
	j.JSONRPC = "2.0"
	return j
}

// NewJSON2Request creates a new JSON2Request with the input parameters
func NewJSON2Request(method string, id, params interface{}) *JSON2Request {
	j := new(JSON2Request)
	j.JSONRPC = "2.0"
	j.ID = id
	j.Params = params
	j.Method = method
	return j
}

// ParseJSON2Request unmarshals a string into a JSON2Request object
func ParseJSON2Request(request string) (*JSON2Request, error) {
	j := new(JSON2Request)
	err := json.Unmarshal([]byte(request), j)
	if err != nil {
		return nil, err
	}
	if j.JSONRPC != "2.0" {
		return nil, fmt.Errorf("Invalid JSON RPC version - `%v`, should be `2.0`", j.JSONRPC)
	}
	return j, nil
}

// JSON2Response is an JSON RPC response object used in factomd according to the specification for JSON RPC 2.0
// See specification: https://www.jsonrpc.org/specification
// Remote Proceedure Call (RPC) is a protocol for executing subroutines on a remote
type JSON2Response struct {
	JSONRPC string      `json:"jsonrpc"` // version string which MUST be "2.0" (version 1.0 didn't have this field)
	ID      interface{} `json:"id"`      // Unique client defined ID associated with incoming request. It may be a number,
	// string, or nil. It is used by the remote to formulate a this response object
	// containing the same ID back to the client. If nil, remote treats request object
	// as a notification, and that the client does not expect a response object back
	Error  *JSONError  `json:"error,omitempty"`  // Must be present if called subroutine had an error (mutually exclusive with Result)
	Result interface{} `json:"result,omitempty"` // Must be present if called subroutine succeeded (mutually exclusive with Error)
}

// JSONByte encodes object to a json []byte
func (j *JSON2Response) JSONByte() ([]byte, error) {
	return EncodeJSON(j)
}

// JSONString encodes object to a json string with error return
func (j *JSON2Response) JSONString() (string, error) {
	return EncodeJSONString(j)
}

// String encodes object to a json string with no error return
func (j *JSON2Response) String() string {
	str, _ := j.JSONString()
	return str
}

// NewJSON2Response returns a new JSON2Response initialized with defaults
func NewJSON2Response() *JSON2Response {
	j := new(JSON2Response)
	j.JSONRPC = "2.0"
	return j
}

// AddError sets the member Error with a new JSONError formed from the inputs
func (j *JSON2Response) AddError(code int, message string, data interface{}) {
	err := NewJSONError(code, message, data)
	j.Error = err
}

// JSONError is an JSON RPC error object used in factomd according to the specification for JSON RPC 2.0
// See specification: https://www.jsonrpc.org/specification
// Remote Proceedure Call (RPC) is a protocol for executing subroutines on a remote
type JSONError struct {
	Code    int         `json:"code"`           // The error code associated with the error type
	Message string      `json:"message"`        // The error message as a concise single sentence
	Data    interface{} `json:"data,omitempty"` // Optional data object containing additional information about the error
}

// NewJSONError returns a new JSONError struct filled with the inputs
func NewJSONError(code int, message string, data interface{}) *JSONError {
	j := new(JSONError)
	j.Code = code
	j.Message = message
	j.Data = data
	return j
}

// Error returns the member 'Message' of JSONError struct. If member 'Data' is also a string type,
// the return string is concatenated with a colon ":" and the string contents of 'Data'
func (j *JSONError) Error() string {
	str, ok := j.Data.(string)
	if ok == false {
		return j.Message
	}
	return j.Message + ": " + str
}
