// +build all

package securedb_test

import (
	"os"
	"testing"

	//	"github.com/FactomProject/factomd/common/primitives"
	//"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/database/securedb"
)

// Basic DB interactions are tested from the generic tester. This checks the encryption

func TestSecureDB(t *testing.T) {
	s, err := NewEncryptedDB("test.db", "Bolt", "rightPassword")
	if err != nil {
		t.Error(err)
	}
	s.Close()

	_, err = NewEncryptedDB("test.db", "Bolt", "wrongPassword")
	if err == nil {
		t.Error("Should error")
	}

	os.Remove("test.db")
}
