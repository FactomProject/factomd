// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import (
	"bytes"
	"encoding/json"
)

//Interface for printing structures into JSON
type JSONable interface {
	JSONByte() ([]byte, error)
	JSONString() (string, error)
	JSONBuffer(b *bytes.Buffer) error
}

//Interface for both JSON and Spew
type Printable interface {
	JSONable
	//String() string
}

//Interface for short, reoccuring data structures to interpret themselves into human-friendly form
type ShortInterpretable interface {
	IsInterpretable() bool //Whether the structure can interpret itself
	Interpret() string     //Turns the data encoded int he structure into human-friendly string
}

func DecodeJSON(data []byte, v interface{}) error {
	err := json.Unmarshal(data, &v)
	return err
}

func EncodeJSON(data interface{}) ([]byte, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func EncodeJSONString(data interface{}) (string, error) {
	encoded, err := EncodeJSON(data)
	if err != nil {
		return "", err
	}
	return string(encoded), err
}

func DecodeJSONString(data string, v interface{}) error {
	return DecodeJSON([]byte(data), v)
}

func EncodeJSONToBuffer(data interface{}, b *bytes.Buffer) error {
	encoded, err := EncodeJSON(data)
	if err != nil {
		return err
	}
	_, err = b.Write(encoded)
	return err
}
