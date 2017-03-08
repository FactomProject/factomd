// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/primitives"
)

func TestServerMarshalUnmarshal(t *testing.T) {
	for i := 0; i < 1000; i++ {
		s := RandomServer()
		b, err := s.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		s2 := new(Server)
		rest, err := s2.UnmarshalBinaryData(b)
		if err != nil {
			t.Errorf("%v", err)
		}
		if len(rest) > 0 {
			t.Errorf("Returned too much data")
		}
		if s.IsSameAs(s2) == false {
			t.Errorf("Servers are not the same")
		}
	}
}
