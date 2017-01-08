// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid_test

import (
	. "github.com/FactomProject/factomd/common/factoid"
	"testing"
)

func TestFactoid(t *testing.T) {
	if FactoidTx_VersionCheck(0) != true {
		t.Fail()
	}
	if FactoidTx_VersionCheck(1) != false {
		t.Fail()
	}

	if FactoidTx_LocktimeCheck(0) != true {
		t.Fail()
	}
	if FactoidTx_LocktimeCheck(1) != false {
		t.Fail()
	}

	if FactoidTx_RCDVersionCheck(0) != true {
		t.Fail()
	}
	if FactoidTx_RCDVersionCheck(1) != false {
		t.Fail()
	}

	if FactoidTx_RCDTypeCheck(0) != true {
		t.Fail()
	}
	if FactoidTx_RCDTypeCheck(1) != false {
		t.Fail()
	}
}
