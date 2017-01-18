package entryCreditBlock_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	ed "github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/go-spew/spew"
)

var _ = fmt.Sprint("testing")

func TestUnmarshalNilECBlock(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(ECBlock)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestStaticECBlockUnmarshal(t *testing.T) {
	ecb := NewECBlock()
	data, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000cbb3ff38bbb90032de6965587f46dcf37551ac26e15819303057c88999b2910b4f87cfc073df0e82cdc2ed0bb992d7ea956fd32b435b099fc35f4b0696948507a66fb49a15b68a2a0ce2382e6aa6970c835497c6074bec9794ccf84bb331ad1350000000100000000000000000b0000000000000058000001020103010401050106010701080417ef7a21d1a616d65e6b73f3c6a7ad5c49340a6c2592872020ec60767ff00d7dc3d09d10693eb867e2bd0a503746df370403c9451ae91a363046f2a68529c2fd00822c0109010a")
	rest, err := ecb.UnmarshalBinaryData(data)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(rest) > 0 {
		t.Error("Returned extra data")
	}
	h, err := ecb.HeaderHash()
	if err != nil {
		t.Errorf("%v", err)
	}
	expected := "c96a851d95db6d58cbcfdd63a8aaf93fc180fb8c003af5508667cc44fa31457d"
	if h.String() != expected {
		t.Errorf("Wrong hash - %v vs %v", h.String(), expected)
	}

	h, err = ecb.GetFullHash()
	if err != nil {
		t.Errorf("%v", err)
	}
	expected = "1eb3121d81cd8676f20c5fec2f4e0d7a892a2ab2f086506bf55735756098d9ba"
	if h.String() != expected {
		t.Errorf("Wrong hash - %v vs %v", h.String(), expected)
	}

}

func TestECBlockMarshal(t *testing.T) {
	ecb1 := createECBlock()

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

func TestECBlockHashingConsistency(t *testing.T) {
	ecb := createECBlock()
	h1, err := ecb.GetFullHash()
	if err != nil {
		t.Error(err)
	}
	k1, err := ecb.HeaderHash()
	if err != nil {
		t.Error(err)
	}
	h2, err := ecb.GetFullHash()
	if err != nil {
		t.Error(err)
	}
	k2, err := ecb.HeaderHash()
	if err != nil {
		t.Error(err)
	}
	if primitives.AreBytesEqual(h1.Bytes(), h2.Bytes()) == false {
		t.Error("ecb.Hash()es are not equal")
	}
	if primitives.AreBytesEqual(k1.Bytes(), k2.Bytes()) == false {
		t.Error("ecb.HeaderHash()es are not equal")
	}
}

func createECBlock() *ECBlock {
	ecb1 := NewECBlock().(*ECBlock)

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
		panic(err)
	} else {
		cc.ECPubKey = (*primitives.ByteSlice32)(pub)
		cc.Sig = (*primitives.ByteSlice64)(ed.Sign(privkey, cc.CommitMsg()))
	}

	// create a ECBlock for testing
	ecb1.Header.(*ECBlockHeader).BodyHash.SetBytes(byteof(0x22))
	ecb1.Header.(*ECBlockHeader).PrevHeaderHash.SetBytes(byteof(0x33))
	ecb1.Header.(*ECBlockHeader).PrevFullHash.SetBytes(byteof(0x44))
	ecb1.Header.(*ECBlockHeader).DBHeight = 10
	ecb1.Header.(*ECBlockHeader).HeaderExpansionArea = byteof(0x55)
	ecb1.Header.(*ECBlockHeader).ObjectCount = 0

	// add the CommitChain to the ECBlock
	ecb1.AddEntry(cc)

	m1 := NewMinuteNumber(0x01)
	ecb1.AddEntry(m1)

	// add a ServerIndexNumber
	si1 := NewServerIndexNumber()
	si1.ServerIndexNumber = 3
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

	m2 := NewMinuteNumber(0x02)
	ecb1.AddEntry(m2)

	return ecb1
}

func byteof(b byte) []byte {
	r := make([]byte, 0, 32)
	for i := 0; i < 32; i++ {
		r = append(r, b)
	}
	return r
}

func TestExpandedECBlockHeader(t *testing.T) {
	block := createECBlock()
	j, err := block.JSONString()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !strings.Contains(j, `"ChainID":"000000000000000000000000000000000000000000000000000000000000000c"`) {
		t.Error("Header does not contain ChainID")
	}
	if !strings.Contains(j, `"ECChainID":"000000000000000000000000000000000000000000000000000000000000000c"`) {
		t.Error("Header does not contain ECChainID")
	}
}
