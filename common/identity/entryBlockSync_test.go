// +build all

package identity_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/identity"
)

func TestUnmarshalBadEntryBlockSync(t *testing.T) {
	ebs := NewEntryBlockSync()

	p, err := ebs.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	// wright bad block count into EntryBlockSync
	p[99] = 0xff

	ebs2 := new(EntryBlockSync)
	err = ebs2.UnmarshalBinary(p)
	if err == nil {
		t.Error("EntryBlockSync should have errored on unmarshal", ebs2)
	} else {
		t.Log(err)
	}
}
