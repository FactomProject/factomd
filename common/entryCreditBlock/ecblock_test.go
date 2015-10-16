package entryCreditBlock_test

import (
	"crypto/rand"
	"fmt"
	"testing"

	ed "github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/davecgh/go-spew/spew"
)

var _ = fmt.Sprint("testing")

func TestECBlockMarshal(t *testing.T) {
	ecb1 := NewECBlock()

	// build a CommitChain for testing
	cc := NewCommitChain()
	cc.Version = 0
	cc.MilliTime = (*primitives.ByteSlice6)(&[6]byte{1, 1, 1, 1, 1, 1})
	cc.ChainIDHash.SetBytes(byteof(0xaa))
	cc.Weld.SetBytes(byteof(0xbb))
	cc.EntryHash.SetBytes(byteof(0xcc))
	cc.Credits = 11

	// make a key and sign the msg
	if pub, privkey, err := ed.GenerateKey(rand.Reader); err != nil {
		t.Error(err)
	} else {
		cc.ECPubKey = (*primitives.ByteSlice32)(pub)
		cc.Sig = (*primitives.ByteSlice64)(ed.Sign(privkey, cc.CommitMsg()))
	}

	// create a ECBlock for testing
	ecb1.Header.ECChainID.SetBytes(byteof(0x11))
	ecb1.Header.BodyHash.SetBytes(byteof(0x22))
	ecb1.Header.PrevHeaderHash.SetBytes(byteof(0x33))
	ecb1.Header.PrevLedgerKeyMR.SetBytes(byteof(0x44))
	ecb1.Header.DBHeight = 10
	ecb1.Header.HeaderExpansionArea = byteof(0x55)
	ecb1.Header.ObjectCount = 0

	// add the CommitChain to the ECBlock
	ecb1.AddEntry(cc)

	m1 := NewMinuteNumber()
	m1.Number = 0x01
	ecb1.AddEntry(m1)

	// add a ServerIndexNumber
	si1 := NewServerIndexNumber()
	si1.Number = 3
	ecb1.AddEntry(si1)

	// create an IncreaseBalance for testing
	ib := NewIncreaseBalance()
	pub := new(primitives.ByteSlice32)
	copy(pub[:], byteof(0xaa))
	ib.ECPubKey = pub
	ib.TXID.SetBytes(byteof(0xbb))
	ib.NumEC = uint64(13)
	// add the IncreaseBalance
	ecb1.AddEntry(ib)

	m2 := NewMinuteNumber()
	m2.Number = 0x02
	ecb1.AddEntry(m2)

	ecb2 := NewECBlock()
	if p, err := ecb1.MarshalBinary(); err != nil {
		t.Error(err)
	} else {
		if err := ecb2.UnmarshalBinary(p); err != nil {
			t.Error(err)
		}
		t.Log(spew.Sdump(ecb1))
		t.Log(spew.Sdump(ecb2))
		if q, err := ecb2.MarshalBinary(); err != nil {
			t.Error(err)
		} else if string(p) != string(q) {
			t.Errorf("ecb1 = %x\n", p)
			t.Errorf("ecb2 = %x\n", q)
		}
	}
}

func byteof(b byte) []byte {
	r := make([]byte, 0, 32)
	for i := 0; i < 32; i++ {
		r = append(r, b)
	}
	return r
}
