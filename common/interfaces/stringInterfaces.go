// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

//Interface for printing structures into JSON
type JSONable interface {
	JSONByte() ([]byte, error)
	JSONString() (string, error)
}

//Interface for both JSON and Spew
type Printable interface {
	JSONable
	String() string
}

//Interface for short, reoccuring data structures to interpret themselves into human-friendly form
type ShortInterpretable interface {
	IsInterpretable() bool //Whether the structure can interpret itself
	Interpret() string     //Turns the data encoded in the structure into human-friendly string
}

type IndividualObj struct {
	ERROR string `json:"err"`
}

type StructToReturnValues struct {
	TempBal int64  `json:"ack"`
	PermBal int64  `json:"saved"`
	Error   string `json:"err"`
}
