// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity_test

import (
	"testing"

	"math/rand"

	"fmt"

	"bytes"

	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/identityEntries"
	"github.com/FactomProject/factomd/common/primitives/random"
)

//import . "github.com/FactomProject/factomd/common/identity"

func TestIdentityManagerMarshal(t *testing.T) {
	for i := 0; i < 100; i++ {
		im := identity.NewIdentityManager()
		for i := 0; i < rand.Intn(10); i++ {
			id := identity.RandomIdentity()
			im.Identities[id.IdentityChainID.Fixed()] = id
		}

		for i := 0; i < rand.Intn(10); i++ {
			id := identity.RandomAuthority()
			im.Authorities[id.AuthorityChainID.Fixed()] = id
		}
		for i := 0; i < rand.Intn(10); i++ {
			r := identityEntries.RandomRegisterFactomIdentityStructure()
			im.IdentityRegistrations[r.IdentityChainID.Fixed()] = r
		}

		data, err := im.MarshalBinary()
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		extra := rand.Intn(100)
		extraData := append(data, random.RandByteSliceOfLen(extra)...)

		im2 := identity.NewIdentityManager()
		newData, err := im2.UnmarshalBinaryData(extraData)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		if len(newData) != extra {
			t.Errorf(fmt.Sprintf("Extra %d data", len(newData)))
		}

		if !im2.IsSameAs(im) {
			t.Errorf("Not same")
		}

		data2, err := im2.MarshalBinary()
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		if bytes.Compare(data, data2) != 0 {
			t.Errorf("Bytes are different: \n%x \n%x", data, data2)
		}
	}
}

func TestIdentityManagerClone(t *testing.T) {
	a := identity.RandomIdentityManager()
	b := a.Clone()

	if !a.IsSameAs(b) {
		t.Error("Clone is not the same")
	}
}
