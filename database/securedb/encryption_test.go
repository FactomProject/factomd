package securedb_test

import (
	"crypto/sha256"
	"crypto/subtle"
	"testing"

	"github.com/PaulSnow/factom2d/common/primitives/random"
	. "github.com/PaulSnow/factom2d/database/securedb"
)

func TestGetKey(t *testing.T) {
	for i := 0; i < 15; i++ {
		pass := random.RandomString()
		salt := random.RandByteSliceOfLen(100)
		key := sha256.Sum256(salt)
		saltedKey, err := GetKey(pass, key[:])
		if err != nil {
			t.Error(err)
		}

		saltedKey2, err := GetKey(pass, key[:])
		if err != nil {
			t.Error(err)
		}

		if subtle.ConstantTimeCompare(saltedKey, saltedKey2) == 0 {
			t.Error("Different keys, but same password")
		}

		key = sha256.Sum256(random.RandByteSliceOfLen(100))
		saltedKey3, err := GetKey(pass, key[:])
		if err != nil {
			t.Error(err)
		}

		if subtle.ConstantTimeCompare(saltedKey, saltedKey3) == 1 {
			t.Error("Different salts, but same password")
		}
	}
}

func TestEncryption(t *testing.T) {
	for i := 0; i < 100; i++ {
		text := random.RandByteSliceOfLen(100)
		key := sha256.Sum256(text)
		saltedKey, _ := GetKey(random.RandomString(), key[:])

		ciphertext, err := Encrypt(text, saltedKey[:])
		if err != nil {
			t.Error(err)
		}

		for i := 0; i < 10; i++ {
			plaintext, err := Decrypt(ciphertext, saltedKey[:])
			if err != nil {
				t.Error(err)
			}

			if subtle.ConstantTimeCompare(plaintext, text) == 0 {
				t.Error("Encyption did not produce the same result")
			}
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
