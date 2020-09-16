// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity_test

import (
	"testing"

	"math/rand"

	"fmt"

	"bytes"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/identityEntries"
	"github.com/FactomProject/factomd/common/primitives/random"
)

//import . "github.com/FactomProject/factomd/common/identity"

// TestIdentityManagerMarshal creates 100 new identity managers with random entries, and ensures they can be marshaled and unmarshaled properly
func TestIdentityManagerMarshal(t *testing.T) {
	for i := 0; i < 100; i++ {
		im := identity.NewIdentityManager()
		for i := 0; i < rand.Intn(10); i++ {
			id := identity.RandomIdentity()
			im.SetIdentity(id.IdentityChainID, id)

			id2 := im.GetIdentity(id.IdentityChainID)
			if id2 == nil {
				t.Errorf("Added Identity not found")
			}
			if id.IsSameAs(id2) == false {
				t.Errorf("Added Identity not the same as original")
			}
			im.RemoveIdentity(id.IdentityChainID)
			id3 := im.GetIdentity(id.IdentityChainID)
			if id3 != nil && id3.Status != constants.IDENTITY_SKELETON { // Skeleton ids are not deleted, so don't check them
				t.Errorf("Found identity when it should have been removed")
			}
			im.SetIdentity(id.IdentityChainID, id) // Reset the id after the above tests
		}

		for i := 0; i < rand.Intn(10); i++ {
			id := identity.RandomAuthority()
			im.SetAuthority(id.AuthorityChainID, id)
			id2 := im.GetAuthority(id.AuthorityChainID)
			if id2 == nil {
				t.Errorf("Added authority not found")
			}
			if id.IsSameAs(id2) == false {
				t.Errorf("Added authority not the same as original")
			}
			im.RemoveAuthority(id.AuthorityChainID)
			id3 := im.GetAuthority(id.AuthorityChainID)
			if id3 != nil {
				t.Errorf("Found identity when it should have been removed")
			}
			im.SetAuthority(id.AuthorityChainID, id) // Reset the id after the above tests
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

// TestIdentityManagerClone checks that the clone function returns an identical identity manager
func TestIdentityManagerClone(t *testing.T) {
	a := identity.RandomIdentityManager()
	b := a.Clone()

	if !a.IsSameAs(b) {
		t.Error("Clone is not the same")
	}
}
