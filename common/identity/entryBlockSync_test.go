package identity_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/identity"
)

// TestUnmarshalBadEntryBlockSync creates a new entry block sync, marshals the object, corrupts it and tries to unmarshal the corrupted object
// The test ensures unmarshaling catches this error
func TestUnmarshalBadEntryBlockSync(t *testing.T) {
	ebs := NewEntryBlockSync()

	p, err := ebs.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	// write bad block count into EntryBlockSync
	p[99] = 0xff

	ebs2 := new(EntryBlockSync)
	err = ebs2.UnmarshalBinary(p)
	if err == nil {
		t.Error("EntryBlockSync should have errored on unmarshal", ebs2)
	} else {
		t.Log(err)
	}
}
