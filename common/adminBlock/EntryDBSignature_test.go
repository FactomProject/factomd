package adminBlock_test

import (
	"encoding/hex"
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/state"
)

func TestDBSignatureEntryGetHash(t *testing.T) {
	a := new(DBSignatureEntry)
	h := a.Hash()
	expected := "b84147b0eeb997d0942e214ce03c7889e5653f276830838b91e2dfea9528d46d"
	if h.String() != expected {
		t.Errorf("Wrong hash returned - %v vs %v", h.String(), expected)
	}
}

func TestDBSignatureEntryTypeIDCheck(t *testing.T) {
	a := new(DBSignatureEntry)
	b, err := a.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	if b[0] != a.Type() {
		t.Errorf("Invalid byte marshalled")
	}
	a2 := new(DBSignatureEntry)
	err = a2.UnmarshalBinary(b)
	if err != nil {
		t.Errorf("%v", err)
	}

	b[0] = (b[0] + 1) % 255
	err = a2.UnmarshalBinary(b)
	if err == nil {
		t.Errorf("No error caught")
	}
}

func TestUnmarshalNilDBSignatureEntry(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(DBSignatureEntry)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestDBSEMisc(t *testing.T) {
	dbse := new(DBSignatureEntry)
	if dbse.IsInterpretable() != false {
		t.Fail()
	}
	if dbse.Interpret() != "" {
		t.Fail()
	}
}

func TestDBSEGenesisBlock(t *testing.T) {
	str := "0100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a83efbcbed19b5842e5aa06e66c41d8b61826d95d50c1cbc8bd5373f986c370547133462a9ffa0dcff025a6ad26747c95f1bdd88e2596fc8c6eaa8a2993c72c05"
	h, _ := hex.DecodeString(str)
	dbse := new(DBSignatureEntry)
	err := dbse.UnmarshalBinary(h)
	if err != nil {
		t.Errorf("%v", err)
	}

	dBlock, _, _, _ := state.GenerateGenesisBlocks(constants.MAIN_NETWORK_ID, nil)
	if dBlock == nil {
		t.Errorf("DBlock is nil")
		t.FailNow()
	}

	bin, _ := dBlock.GetHeader().MarshalBinary()

	t.Logf("%x", bin)
	if dbse.PrevDBSig.Verify(bin) == false {
		t.Errorf("Invalid signature")
	}
}

func TestDBsigMisc(t *testing.T) {
	a := new(DBSignatureEntry)
	if a.String() != "    E:         DB Signature --   IdentityChainID     0000       PubKey 00000000    Signature 3030303030303030" {
		t.Error("Unexpected string:", a.String())
	}
	as, err := a.JSONString()
	if err != nil {
		t.Error(err)
	}
	if as != "{\"adminidtype\":1,\"identityadminchainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"prevdbsig\":{\"pub\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"sig\":\"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\"}}" {
		t.Error("Unexpected JSON string:", as)
	}
	ab, err := a.JSONByte()
	if err != nil {
		t.Error(err)
	}
	if string(ab) != "{\"adminidtype\":1,\"identityadminchainid\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"prevdbsig\":{\"pub\":\"0000000000000000000000000000000000000000000000000000000000000000\",\"sig\":\"00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000\"}}" {
		t.Error("Unexpected JSON bytes:", string(ab))
	}

	if a.IsInterpretable() {
		t.Error("IsInterpretable should return false")
	}
	if a.Interpret() != "" {
		t.Error("Interpret should return empty string")
	}
}
