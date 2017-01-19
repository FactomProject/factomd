// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"encoding/json"
	"fmt"
)

type JSON2Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Params  interface{} `json:"params,omitempty"`
	Method  string      `json:"method,omitempty"`
}

func (e *JSON2Request) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *JSON2Request) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (e *JSON2Request) String() string {
	str, _ := e.JSONString()
	return str
}

func NewJSON2RequestBlank() *JSON2Request {
	j := new(JSON2Request)
	j.JSONRPC = "2.0"
	return j
}

func NewJSON2Request(method string, id, params interface{}) *JSON2Request {
	j := new(JSON2Request)
	j.JSONRPC = "2.0"
	j.ID = id
	j.Params = params
	j.Method = method
	return j
}

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

type JSON2Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Error   *JSONError  `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

func (e *JSON2Response) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *JSON2Response) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (e *JSON2Response) String() string {
	str, _ := e.JSONString()
	return str
}

func NewJSON2Response() *JSON2Response {
	j := new(JSON2Response)
	j.JSONRPC = "2.0"
	return j
}

func (j *JSON2Response) AddError(code int, message string, data interface{}) {
	e := NewJSONError(code, message, data)
	j.Error = e
}

type JSONError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewJSONError(code int, message string, data interface{}) *JSONError {
	j := new(JSONError)
	j.Code = code
	j.Message = message
	j.Data = data
	return j
}

func (j *JSONError) Error() string {
	str, ok := j.Data.(string)
	if ok == false {
		return j.Message
	}
	return j.Message + ": " + str
}
