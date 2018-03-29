// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package identity_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestSetSkeletonKey(t *testing.T) {
	im := NewIdentityManager()
	for i := 0; i < 1000; i++ {
		h := primitives.RandomHash()
		str := h.String()
		err := im.SetSkeletonKey(str)
		if err != nil {
			t.Errorf("%v", err)
		}
		auth := im.GetAuthority(primitives.NewZeroHash())
		str2 := auth.SigningKey.String()
		if str != str2 {
			t.Errorf("Invalid signing key - %v vs %v", str, str2)
		}

		str = str[1:]
		err = im.SetSkeletonKey(str)
		if err == nil {
			t.Errorf("No error returned")
		}
	}
}

func TestSetSkeletonKeyMainNet(t *testing.T) {
	im := NewIdentityManager()
	str := "0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a"
	err := im.SetSkeletonKeyMainNet()
	if err != nil {
		t.Errorf("%v", err)
	}
	auth := im.GetAuthority(primitives.NewZeroHash())
	str2 := auth.SigningKey.String()
	if str != str2 {
		t.Errorf("Invalid signing key - %v vs %v", str, str2)
	}
}
