// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/identity"
)

// TesetHistoricKeyMarshalUnmarshal checks that 1000 random historic keys can be marshaled and unmarshaled properly
func TestHistoricKeyMarshalUnmarshal(t *testing.T) {
	for i := 0; i < 1000; i++ {
		hk := RandomHistoricKey()
		h, err := hk.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}
		hk2 := new(HistoricKey)
		err = hk2.UnmarshalBinary(h)
		if err != nil {
			t.Errorf("%v", err)
		}
		if hk.IsSameAs(hk2) == false {
			t.Errorf("Historic keys are not identical")
		}
	}
}
