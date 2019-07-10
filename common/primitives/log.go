// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
)

// Log prints the formated string with args (using fmt.Printf) with a prepended file and line number of
// the calling function: "<file>:<linenumber> - <fmt.Printf_output>\n"
func Log(format string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}
	fmt.Printf(file+":"+strconv.Itoa(line)+" - "+format+"\n", args...)
}

// LogJSONs converts input args to Json format and then prints the formated string with
// new Json args (using fmt.Printf) with a prepended file and line number of the calling
// function: "<file>:<linenumber> - <fmt.Printf_output>\n"
func LogJSONs(format string, args ...interface{}) {
	jsons := []interface{}{}
	for _, v := range args {
		j, _ := EncodeJSONString(v)
		jsons = append(jsons, j)
	}
	// Note: I believe below could just call Log(format,jsons) and achieve the same functionality and gaurantee
	// formating between Log() and LogJSONs(). However, the newly created unit test for LogJSONs() fails if I do
	// this. I suspect its because this function LogJSONs() doesn't behave has it should be, because these
	// functions aren't called anywhere that I can find, but I don't want to change the functionality.
	// Log(format, jsons) // This call can be uncommented and delete the below lines in the future
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}
	fmt.Printf(file+":"+strconv.Itoa(line)+" - "+format+"\n", jsons...)
}

// DecodeJSON unmarshals input []byte into input interface
func DecodeJSON(data []byte, v interface{}) error {
	err := json.Unmarshal(data, &v)
	return err
}

// EncodeJSON marshals data into Json format and returns a []byte
func EncodeJSON(data interface{}) ([]byte, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

// EncodeJSONString marshals data into Json format and returns a string
func EncodeJSONString(data interface{}) (string, error) {
	encoded, err := EncodeJSON(data)
	if err != nil {
		return "", err
	}
	return string(encoded), err
}

// DecodeJSONString unmarshals input string into input interface
func DecodeJSONString(data string, v interface{}) error {
	return DecodeJSON([]byte(data), v)
}

// EncodeJSONToBuffer marshals input data into Json format and writes it to the input buffer
func EncodeJSONToBuffer(data interface{}, b *bytes.Buffer) error {
	encoded, err := EncodeJSON(data)
	if err != nil {
		return err
	}
	_, err = b.Write(encoded)
	return err
}
