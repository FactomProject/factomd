package securedb_test

import (
	"testing"

	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/common/primitives/random"
	. "github.com/PaulSnow/factom2d/database/securedb"
)

func TestSecureDBMetaData(t *testing.T) {
	for i := 0; i < 100; i++ {
		var s primitives.ByteSlice
		m := new(SecureDBMetaData)
		s.Bytes = random.RandByteSlice()
		m.Salt = s

		var c primitives.ByteSlice
		c.Bytes = random.RandByteSlice()
		m.Challenge = c

		data, err := m.MarshalBinary()
		if err != nil {
			t.Error(err)
		}

		m2 := new(SecureDBMetaData)
		nd, err := m2.UnmarshalBinaryData(data)
		if err != nil {
			t.Error(err)
		}
		if len(nd) != 0 {
			t.Errorf("Should have 0 bytes left, found %d", len(nd))
		}

		if !m.IsSameAs(m2) {
			t.Errorf("Not same %x | %x", m.Challenge.Bytes, m2.Challenge.Bytes)
		}
	}
}
