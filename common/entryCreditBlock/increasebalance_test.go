package entryCreditBlock_test

import (
	"testing"

	. "github.com/PaulSnow/factom2d/common/entryCreditBlock"
	"github.com/PaulSnow/factom2d/common/primitives"
)

// TestIncreaseBalanceMarshalUnmarshal checks that an increase balance object can be marshalled and unmarshalled correctly
func TestIncreaseBalanceMarshalUnmarshal(t *testing.T) {
	ib1 := NewIncreaseBalance()
	pub := new([32]byte)
	copy(pub[:], byteof(0xaa))
	ib1.ECPubKey = (*primitives.ByteSlice32)(pub)
	ib1.TXID.SetBytes(byteof(0xbb))
	ib1.NumEC = uint64(13)
	p, err := ib1.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	ib2 := NewIncreaseBalance()
	ib2.UnmarshalBinary(p)

	q, err := ib2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if string(p) != string(q) {
		t.Errorf("ib1 = %x\n", p)
		t.Errorf("ib2 = %x\n", q)
	}
}

// TestInvalidIncreaseBalanceUnmarshal checks that unmarshalling nil and the empty interface throw errors
func TestInvalidIncreaseBalanceUnmarshal(t *testing.T) {
	ib := NewIncreaseBalance()
	_, err := ib.UnmarshalBinaryData(nil)
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	ib = NewIncreaseBalance()
	_, err = ib.UnmarshalBinaryData([]byte{})
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}

	ib = NewIncreaseBalance()
	err = ib.UnmarshalBinary(nil)
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	ib = NewIncreaseBalance()
	err = ib.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
}
