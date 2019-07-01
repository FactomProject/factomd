// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"bytes"
	"testing"

	. "github.com/FactomProject/factomd/common/primitives"
)

func TestJsonString(t *testing.T) {
	var methods = []string{"a", "b", "c", "random", "somenumbers0214", "@#%&", "blank"}
	for i, meth := range methods {
		r := NewJSON2Request(meth, i, nil)
		r2 := NewJSON2Request(meth, i, nil)

		if meth == "blank" {
			r = NewJSON2RequestBlank()
			r2 = NewJSON2RequestBlank()
		}

		b, err := r.JSONByte()
		if err != nil {
			t.Error(err)
		}

		b2, err := r2.JSONByte()
		if err != nil {
			t.Error(err)
		}

		if bytes.Compare(b, b2) != 0 {
			t.Error("Should be equal bytes")
		}

		s, err := r.JSONString()
		if err != nil {
			t.Error(err)
		}

		s2, err := r2.JSONString()
		if err != nil {
			t.Error(err)
		}

		if s != s2 {
			t.Error("Should be equal strings")
		}

		if r.String() != r2.String() {
			t.Error("Should be equal strings")
		}
	}
}

func TestBadUnmarshal(t *testing.T) {
	_, err := ParseJSON2Request("baddata")
	if err == nil {
		t.Error("Should error on bad data, but it did not")
	}

	_, err = ParseJSON2Request(`{"jsonrpc":"1.0","id":2,"method":"c"}`)
	if err == nil {
		t.Error("Should as JSONRPC must be 2.0")
	}
}

func TestJsonResp(t *testing.T) {
	var errors = []string{"a", "b", "c", "random", "somenumbers0214", "@#%&", "blank"}
	for i, er := range errors {
		r := NewJSON2Response()
		r2 := NewJSON2Response()
		r.AddError(i, er, "")
		r2.AddError(i, er, "")

		b, err := r.JSONByte()
		if err != nil {
			t.Error(err)
		}

		b2, err := r2.JSONByte()
		if err != nil {
			t.Error(err)
		}

		if bytes.Compare(b, b2) != 0 {
			t.Error("Should be equal bytes")
		}

		s, err := r.JSONString()
		if err != nil {
			t.Error(err)
		}

		s2, err := r2.JSONString()
		if err != nil {
			t.Error(err)
		}

		if s != s2 {
			t.Error("Should be equal strings")
		}

		if r.Error.Error() != r2.Error.Error() {
			t.Error("Should be equal strings")
		}
	}
}
