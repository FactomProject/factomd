package entryBlock_test

import (
	. "github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/primitives"
	"testing"
)

func TestEBlockMarshal(t *testing.T) {
	t.Logf("\n---\nTestEBlockMarshal\n---\n")

	// build an EBlock for testing
	eb := NewEBlock()
	eb.Header.ChainID.SetBytes(byteof(0x11))
	eb.Header.BodyMR.SetBytes(byteof(0x22))
	eb.Header.PrevKeyMR.SetBytes(byteof(0x33))
	eb.Header.PrevLedgerKeyMR.SetBytes(byteof(0x44))
	eb.Header.EBSequence = 5
	eb.Header.DBHeight = 6
	eb.Header.EntryCount = 7
	ha := primitives.NewZeroHash()
	ha.SetBytes(byteof(0xaa))
	hb := primitives.NewZeroHash()
	hb.SetBytes(byteof(0xbb))
	eb.Body.EBEntries = append(eb.Body.EBEntries, ha)
	eb.AddEndOfMinuteMarker(0xcc)
	eb.Body.EBEntries = append(eb.Body.EBEntries, hb)

	t.Log(eb)
	p, err := eb.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	eb2 := NewEBlock()
	if err := eb2.UnmarshalBinary(p); err != nil {
		t.Error(err)
	}
	t.Log(eb2)
	p2, err := eb2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if string(p) != string(p2) {
		t.Logf("eb1 = %x\n", p)
		t.Logf("eb2 = %x\n", p2)
		t.Fail()
	}
}

func byteof(b byte) []byte {
	r := make([]byte, 0, 32)
	for i := 0; i < 32; i++ {
		r = append(r, b)
	}
	return r
}
