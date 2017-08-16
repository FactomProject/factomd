// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/identity"
)

func TestIdentityMarshalUnmarshal(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id := RandomIdentity()
		h, err := id.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		id2 := new(Identity)
		err = id2.UnmarshalBinary(h)
		if err != nil {
			t.Errorf("%v", err)
		}
		if id.IsSameAs(id2) == false {
			t.Errorf("Identities are not the same")
		}
	}
}
