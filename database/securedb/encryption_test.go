package securedb_test

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/database/securedb"
)

func TestEncryption(t *testing.T) {
	for i := 0; i < 100; i++ {
		text := random.RandByteSliceOfLen(100)
		key := sha256.Sum256(text)
		saltedKey := GetKey(random.RandomString(), key[:])

		ciphertext, err := Encrypt(text, saltedKey[:])
		if err != nil {
			t.Error(err)
		}

		plaintext, err := Decrypt(ciphertext, saltedKey[:])
		if err != nil {
			t.Error(err)
		}

		if bytes.Compare(plaintext, text) != 0 {
			t.Error("Encyption did not produce the same result")
		}
	}
}

func TestEncryptionBadKey(t *testing.T) {
	for i := 0; i < 100; i++ {
		k := make([]byte, i)
		_, e1 := Encrypt([]byte{0x00}, k)
		_, e2 := Decrypt([]byte{0x00}, k)
		if i != 32 {
			if e1 == nil || e2 == nil {
				t.Error("Should error")
			}
		} else {
			if !(e1 == nil || e2 == nil) {
				t.Error("Should not error")
			}
		}
	}
}
