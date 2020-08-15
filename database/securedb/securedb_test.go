package securedb_test

import (
	"os"
	"testing"

	//	"github.com/PaulSnow/factom2d/common/primitives"
	//"github.com/PaulSnow/factom2d/common/primitives/random"
	. "github.com/PaulSnow/factom2d/database/securedb"
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
