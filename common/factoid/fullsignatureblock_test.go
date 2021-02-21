package factoid_test

import (
	"math/rand"
	"testing"

	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func testSig() interfaces.IFullSignature {
	sig := new(primitives.Signature)
	sig.Init()
	rand.Read(sig.Pub[:])
	rand.Read(sig.Sig[:])
	return sig
}

func TestFullSignatureBlock_AddGetSignature(t *testing.T) {
	block := factoid.NewFullSignatureBlock()
	other := factoid.NewFullSignatureBlock()
	sigs := make([]interfaces.IFullSignature, 32)
	for i := range sigs {
		sigs[i] = testSig()
		block.AddSignature(sigs[i])
		other.AddSignature(sigs[i])
	}

	if !block.IsSameAs(other) {
		t.Errorf("two equal blocks did not match")
	}

	allsigs := block.GetSignatures()

	if len(allsigs) != len(sigs) {
		t.Fatalf("not enough sigs in block. got = %d, want = %d", len(allsigs), len(sigs))
	}

	for i := range sigs {
		if !sigs[i].IsSameAs(block.GetSignature(i)) {
			t.Errorf("Signature index mismatch. index = %d, got = %v, want = %v", i, block.GetSignature(i), sigs[i])
		}

		if !sigs[i].IsSameAs(allsigs[i]) {
			t.Errorf("Signature mismatch. index = %d, got = %v, want = %v", i, block.GetSignature(i), sigs[i])
		}
	}

}
