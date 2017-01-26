package adminBlock_test

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
	"github.com/FactomProject/factomd/testHelper"
)

func TestAdminBlockUnmarshalComplexBlock(t *testing.T) {
	a := NewAdminBlock(nil)

	testVector := []interfaces.IABEntry{}

	testVector = append(testVector, new(EndOfMinuteEntry))
	testVector = append(testVector, new(DBSignatureEntry))
	testVector = append(testVector, new(RevealMatryoshkaHash))
	testVector = append(testVector, new(AddReplaceMatryoshkaHash))
	testVector = append(testVector, new(IncreaseServerCount))
	testVector = append(testVector, new(AddFederatedServer))
	testVector = append(testVector, new(AddAuditServer))
	testVector = append(testVector, new(RemoveFederatedServer))
	testVector = append(testVector, new(AddFederatedServerSigningKey))
	testVector = append(testVector, new(AddFederatedServerBitcoinAnchorKey))
	testVector = append(testVector, new(ServerFault))

	for _, v := range testVector {
		err := a.AddABEntry(v)
		if err != nil {
			t.Errorf("%v", err)
		}
	}

	b, err := a.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}

	a2 := new(AdminBlock)
	err = a2.UnmarshalBinary(b)
	if err != nil {
		t.Errorf("%v", err)
	}

	entries := a2.GetABEntries()

	for i, v := range testVector {
		if entries[i].Type() != v.Type() {
			t.Errorf("Invalid type for index %v - %v vs %v", i, entries[i].Type(), v.Type())
		}
	}
}

func TestAdminBlockGetHash(t *testing.T) {
	a := new(AdminBlock)
	h := a.GetHash()
	expected := "0a9aa1efbe7d0e8d9c1d460d1c78e3e7b50f984e65a3f3ee7b73100a94189dbf"
	if h.String() != expected {
		t.Errorf("Wrong hash returned - %v vs %v", h.String(), expected)
	}
}

func TestUnmarshalNilAdminBlock(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(AdminBlock)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestNilFunctions(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(AdminBlock)

	if a.UpdateState(nil) == nil {
		t.Errorf("No error returned")
	}
	if a.AddDBSig(nil, nil) == nil {
		t.Errorf("No error returned")
	}
	if a.AddFedServer(nil) == nil {
		t.Errorf("No error returned")
	}
	if a.AddAuditServer(nil) == nil {
		t.Errorf("No error returned")
	}
	if a.RemoveFederatedServer(nil) == nil {
		t.Errorf("No error returned")
	}
	if a.AddMatryoshkaHash(nil, nil) == nil {
		t.Errorf("No error returned")
	}
	if a.AddFederatedServerSigningKey(nil, [32]byte{}) == nil {
		t.Errorf("No error returned")
	}
	if a.AddFederatedServerBitcoinAnchorKey(nil, 0, 0, [20]byte{}) == nil {
		t.Errorf("No error returned")
	}
	if a.AddEntry(nil) == nil {
		t.Errorf("No error returned")
	}
	if a.AddServerFault(nil) == nil {
		t.Errorf("No error returned")
	}
}

func TestAddDBSig(t *testing.T) {
	testVector := []*DBSignatureEntry{}
	for i := 0; i < 1000; i++ {
		se := new(DBSignatureEntry)
		se.IdentityAdminChainID = primitives.RandomHash()
		_, _, sig := primitives.RandomSignatureSet()
		se.PrevDBSig = *sig.(*primitives.Signature)
		testVector = append(testVector, se)
	}
	ab := new(AdminBlock)
	for _, v := range testVector {
		err := ab.AddDBSig(v.IdentityAdminChainID, &v.PrevDBSig)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	for i := range testVector {
		if ab.ABEntries[i].(*DBSignatureEntry).IdentityAdminChainID.String() != testVector[i].IdentityAdminChainID.String() {
			t.Errorf("Invalid IdentityAdminChainID")
		}
		if ab.ABEntries[i].(*DBSignatureEntry).PrevDBSig.IsSameAs(&testVector[i].PrevDBSig) == false {
			t.Errorf("Invalid PrevDBSig")
		}
	}
}

func TestAddFedServer(t *testing.T) {
	testVector := []interfaces.IHash{}
	for i := 0; i < 1000; i++ {
		testVector = append(testVector, primitives.RandomHash())
	}
	ab := new(AdminBlock)
	ab.Init()
	for i, v := range testVector {
		ab.Header.SetDBHeight(uint32(i))
		err := ab.AddFedServer(v)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	for i := range testVector {
		if ab.ABEntries[i].(*AddFederatedServer).IdentityChainID.String() != testVector[i].String() {
			t.Errorf("Invalid IdentityChainID")
		}
		if ab.ABEntries[i].(*AddFederatedServer).DBHeight != uint32(i+1) {
			t.Errorf("Invalid DBHeight")
		}
	}
}

func TestAddAuditServer(t *testing.T) {
	testVector := []interfaces.IHash{}
	for i := 0; i < 1000; i++ {
		testVector = append(testVector, primitives.RandomHash())
	}
	ab := new(AdminBlock)
	ab.Init()
	for i, v := range testVector {
		ab.Header.SetDBHeight(uint32(i))
		err := ab.AddAuditServer(v)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	for i := range testVector {
		if ab.ABEntries[i].(*AddAuditServer).IdentityChainID.String() != testVector[i].String() {
			t.Errorf("Invalid IdentityChainID")
		}
		if ab.ABEntries[i].(*AddAuditServer).DBHeight != uint32(i+1) {
			t.Errorf("Invalid DBHeight")
		}
	}
}

func TestRemoveFederatedServer(t *testing.T) {
	testVector := []interfaces.IHash{}
	for i := 0; i < 1000; i++ {
		testVector = append(testVector, primitives.RandomHash())
	}
	ab := new(AdminBlock)
	ab.Init()
	for i, v := range testVector {
		ab.Header.SetDBHeight(uint32(i))
		err := ab.RemoveFederatedServer(v)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	for i := range testVector {
		if ab.ABEntries[i].(*RemoveFederatedServer).IdentityChainID.String() != testVector[i].String() {
			t.Errorf("Invalid IdentityChainID")
		}
		if ab.ABEntries[i].(*RemoveFederatedServer).DBHeight != uint32(i+1) {
			t.Errorf("Invalid DBHeight")
		}
	}
}

func TestAddMatryoshkaHash(t *testing.T) {
	testVector := []*AddReplaceMatryoshkaHash{}
	for i := 0; i < 1000; i++ {
		se := new(AddReplaceMatryoshkaHash)
		se.IdentityChainID = primitives.RandomHash()
		se.MHash = primitives.RandomHash()
		testVector = append(testVector, se)
	}
	ab := new(AdminBlock)
	for _, v := range testVector {
		err := ab.AddMatryoshkaHash(v.IdentityChainID, v.MHash)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	for i := range testVector {
		if ab.ABEntries[i].(*AddReplaceMatryoshkaHash).IdentityChainID.String() != testVector[i].IdentityChainID.String() {
			t.Errorf("Invalid IdentityChainID")
		}
		if ab.ABEntries[i].(*AddReplaceMatryoshkaHash).MHash.String() != testVector[i].MHash.String() {
			t.Errorf("Invalid MHash")
		}
	}
}

func TestAddFederatedServerSigningKey(t *testing.T) {
	testVector := []*AddFederatedServerSigningKey{}
	for i := 0; i < 1000; i++ {
		se := new(AddFederatedServerSigningKey)
		se.IdentityChainID = primitives.RandomHash()
		priv := primitives.RandomPrivateKey()
		se.PublicKey = *priv.(*primitives.PrivateKey).Pub
		testVector = append(testVector, se)
	}
	ab := new(AdminBlock)
	for _, v := range testVector {
		err := ab.AddFederatedServerSigningKey(v.IdentityChainID, v.PublicKey.Fixed())
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	for i := range testVector {
		if ab.ABEntries[i].(*AddFederatedServerSigningKey).IdentityChainID.String() != testVector[i].IdentityChainID.String() {
			t.Errorf("Invalid IdentityChainID")
		}
		if primitives.AreBytesEqual(ab.ABEntries[i].(*AddFederatedServerSigningKey).PublicKey[:], testVector[i].PublicKey[:]) == false {
			t.Errorf("Invalid PublicKey")
		}
	}
}

func TestAddFederatedServerBitcoinAnchorKey(t *testing.T) {
	testVector := []*AddFederatedServerBitcoinAnchorKey{}
	for i := 0; i < 1000; i++ {
		se := new(AddFederatedServerBitcoinAnchorKey)
		se.IdentityChainID = primitives.RandomHash()
		b := [20]byte{}
		copy(b[:], random.RandByteSliceOfLen(20))
		testVector = append(testVector, se)
	}
	ab := new(AdminBlock)
	for i, v := range testVector {
		fixed, err := v.ECDSAPublicKey.GetFixed()
		if err != nil {
			t.Errorf("%v", err)
		}

		err = ab.AddFederatedServerBitcoinAnchorKey(v.IdentityChainID, byte(i%256), byte(256-i%256), fixed)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	for i := range testVector {
		if ab.ABEntries[i].(*AddFederatedServerBitcoinAnchorKey).IdentityChainID.String() != testVector[i].IdentityChainID.String() {
			t.Errorf("Invalid IdentityChainID")
		}
		if primitives.AreBytesEqual(ab.ABEntries[i].(*AddFederatedServerBitcoinAnchorKey).ECDSAPublicKey[:], testVector[i].ECDSAPublicKey[:]) == false {
			t.Errorf("Invalid ECDSAPublicKey")
		}
		if ab.ABEntries[i].(*AddFederatedServerBitcoinAnchorKey).KeyPriority != byte(i%256) {
			t.Errorf("Invalid KeyPriority")
		}
		if ab.ABEntries[i].(*AddFederatedServerBitcoinAnchorKey).KeyType != byte(256-i%256) {
			t.Errorf("Invalid KeyType")
		}
	}
}

func TestAdminBlockPreviousHash(t *testing.T) {
	block := new(AdminBlock)
	data, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	_, err := block.UnmarshalBinaryData(data)
	if err != nil {
		t.Error(err)
	}

	backRefHash, err := block.BackReferenceHash()
	if err != nil {
		t.Error(err)
	}

	lookupHash, err := block.LookupHash()
	if err != nil {
		t.Error(err)
	}

	t.Logf("Current hashes - %s, %s", backRefHash.String(), lookupHash.String())

	if backRefHash.String() != "0a9aa1efbe7d0e8d9c1d460d1c78e3e7b50f984e65a3f3ee7b73100a94189dbf" {
		t.Error("Invalid backRefHash")
	}
	if lookupHash.String() != "4fb409d5369fad6aa7768dc620f11cd219f9b885956b631ad050962ca934052e" {
		t.Error("Invalid lookupHash")
	}
	/*
		block2, err := CreateAdminBlock(s, block, 5)
		if err != nil {
			t.Error(err)
		}

		backRefHash2, err := block2.BackReferenceHash()
		if err != nil {
			t.Error(err)
		}

		lookupHash2, err := block2.LookupHash()
		if err != nil {
			t.Error(err)
		}

		t.Logf("Second hashes - %s, %s", backRefHash2.String(), lookupHash2.String())
		t.Logf("Previous hash - %s", block2.Header.PrevBackRefHash.String())

		marshalled, err := block2.MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		t.Logf("Marshalled - %X", marshalled)

		if block2.Header.PrevBackRefHash.String() != backRefHash.String() {
			t.Error("PrevBackRefHash does not match ABHash")
		}
	*/
}

func TestAdminBlockHash(t *testing.T) {
	block := new(AdminBlock)
	data, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000A8D665EE36947529E660101ADF2D2A7D7CA0B045F7932E76F86409AE0CA9123B000000005000000000000000000")
	_, err := block.UnmarshalBinaryData(data)
	if err != nil {
		t.Error(err)
	}

	backRefHash, err := block.BackReferenceHash()
	if err != nil {
		t.Error(err)
	}

	lookupHash, err := block.LookupHash()
	if err != nil {
		t.Error(err)
	}

	t.Logf("Current hashes - %s, %s", backRefHash.String(), lookupHash.String())

	if backRefHash.String() != "9515e5108c89ef004ff4fa01c6511f98c8c11f5c2976c4816f8bcfcc551a134d" {
		t.Error("Invalid backRefHash")
	}
	if lookupHash.String() != "f10eefb55197e34f2875c1727c816fcf6564a44902b716a380f0961406ff92d5" {
		t.Error("Invalid lookupHash")
	}

	j, err := block.JSONString()
	if err != nil {
		t.Errorf("%v", err)
	}
	if strings.Contains(j, `"BackReferenceHash":"9515e5108c89ef004ff4fa01c6511f98c8c11f5c2976c4816f8bcfcc551a134d"`) == false {
		t.Errorf("JSON printout does not contain the backreference hash - %v", j)
	}
	if strings.Contains(j, `"LookupHash":"f10eefb55197e34f2875c1727c816fcf6564a44902b716a380f0961406ff92d5"`) == false {
		t.Errorf("JSON printout does not contain the lookup hash - %v", j)
	}
}

func TestAdminBlockMarshalUnmarshal(t *testing.T) {
	blocks := []interfaces.IAdminBlock{}
	blocks = append(blocks, createSmallTestAdminBlock())
	blocks = append(blocks, createTestAdminBlock())
	for b, block := range blocks {
		binary, err := block.MarshalBinary()
		if err != nil {
			t.Logf("Block %d", b)
			t.Error(err)
			t.FailNow()
		}
		block2 := new(AdminBlock)
		err = block2.UnmarshalBinary(binary)
		if err != nil {
			t.Logf("Block %d", b)
			t.Error(err)
			t.FailNow()
		}
		if len(block2.GetABEntries()) != len(block.GetABEntries()) {
			t.Logf("Block %d", b)
			t.Error("Invalid amount of ABEntries")
			t.FailNow()
		}
		for i := range block2.ABEntries {
			entryOne, err := block.GetABEntries()[i].MarshalBinary()
			if err != nil {
				t.Logf("Block %d", b)
				t.Error(err)
				t.FailNow()
			}
			entryTwo, err := block2.ABEntries[i].MarshalBinary()
			if err != nil {
				t.Logf("Block %d", b)
				t.Error(err)
				t.FailNow()
			}

			if bytes.Compare(entryOne, entryTwo) != 0 {
				t.Logf("Block %d", b)
				t.Logf("%X vs %X", entryOne, entryTwo)
				t.Error("ABEntries are not identical")
			}
		}
	}
}

func TestABlockHeaderMarshalUnmarshal(t *testing.T) {
	header := createTestAdminHeader()

	binary, err := header.MarshalBinary()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	header2 := new(ABlockHeader)
	err = header2.UnmarshalBinary(binary)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if bytes.Compare(header.GetAdminChainID().Bytes(), header2.GetAdminChainID().Bytes()) != 0 {
		t.Error("AdminChainIDs are not identical")
	}

	if bytes.Compare(header.GetPrevBackRefHash().Bytes(), header2.GetPrevBackRefHash().Bytes()) != 0 {
		t.Error("PrevBackRefHashes are not identical")
	}

	if header.GetDBHeight() != header2.GetDBHeight() {
		t.Error("DBHeights are not identical")
	}

	if header.HeaderExpansionSize != header2.HeaderExpansionSize {
		t.Error("HeaderExpansionSizes are not identical")
	}

	if bytes.Compare(header.HeaderExpansionArea, header2.HeaderExpansionArea) != 0 {
		t.Error("HeaderExpansionAreas are not identical")
	}

	if header.MessageCount != header2.MessageCount {
		t.Error("HeaderExpansionSizes are not identical")
	}

	if header.BodySize != header2.BodySize {
		t.Error("HeaderExpansionSizes are not identical")
	}

}

func TestUnmarshalABlock(t *testing.T) {
	raw := "000000000000000000000000000000000000000000000000000000000000000a0a9aa1efbe7d0e8d9c1d460d1c78e3e7b50f984e65a3f3ee7b73100a94189dbf000000010000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a83efbcbed19b5842e5aa06e66c41d8b61826d95d50c1cbc8bd5373f986c370547133462a9ffa0dcff025a6ad26747c95f1bdd88e2596fc8c6eaa8a2993c72c050002"
	b, err := hex.DecodeString(raw)
	if err != nil {
		t.Error(err)
	}
	a, err := UnmarshalABlock(b)
	if err != nil {
		t.Error(err)
	}
	h, err := a.LookupHash()
	if err != nil {
		t.Error(err)
	}
	if h.String() != "b30ab81a8afdbe0be1627ef151bf7e263ce3d39d60b61464d81daa8320c28a4f" {
		t.Error("Invalid Hash")
	}
	h, err = a.BackReferenceHash()
	if err != nil {
		t.Error(err)
	}
	if h.String() != "b2405450392038716e9b24804345f9ac0736792dba436c024268ed8100683894" {
		t.Error("Invalid KeyMR")
	}
}

var WeDidPanic bool

func CatchPanic() {
	if r := recover(); r != nil {
		WeDidPanic = true
	}
}

func TestInvalidABlockHeaderUnmarshal(t *testing.T) {
	WeDidPanic = false
	defer CatchPanic()

	header := new(ABlockHeader)
	err := header.UnmarshalBinary(nil)
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	if WeDidPanic == true {
		t.Error("We did panic and we shouldn't have")
		WeDidPanic = false
		defer CatchPanic()
	}

	header = new(ABlockHeader)
	err = header.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	if WeDidPanic == true {
		t.Error("We did panic and we shouldn't have")
		WeDidPanic = false
		defer CatchPanic()
	}

	header2 := createTestAdminHeader()

	binary, err := header2.MarshalBinary()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	header = new(ABlockHeader)
	err = header.UnmarshalBinary(binary[:len(binary)-1])
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	if WeDidPanic == true {
		t.Error("We did panic and we shouldn't have")
		WeDidPanic = false
		defer CatchPanic()
	}
}

func TestInvalidAdminBlockUnmarshal(t *testing.T) {
	WeDidPanic = false
	defer CatchPanic()

	block := new(AdminBlock)
	err := block.UnmarshalBinary(nil)
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	if WeDidPanic == true {
		t.Error("We did panic and we shouldn't have")
		WeDidPanic = false
		defer CatchPanic()
	}

	block = new(AdminBlock)
	err = block.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	if WeDidPanic == true {
		t.Error("We did panic and we shouldn't have")
		WeDidPanic = false
		defer CatchPanic()
	}

	block2 := createTestAdminBlock()

	binary, err := block2.MarshalBinary()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	block = new(AdminBlock)
	err = block.UnmarshalBinary(binary[:len(binary)-1])
	if err == nil {
		t.Error("We expected errors but we didn't get any")
	}
	if WeDidPanic == true {
		t.Error("We did panic and we shouldn't have")
		WeDidPanic = false
		defer CatchPanic()
	}
}

func TestExpandedABlockHeader(t *testing.T) {
	block := createTestAdminBlock()
	j, err := block.JSONString()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !strings.Contains(j, `"AdminChainID":"000000000000000000000000000000000000000000000000000000000000000a"`) {
		t.Error("Header does not contain AdminChainID")
	}
	if !strings.Contains(j, `"ChainID":"000000000000000000000000000000000000000000000000000000000000000a"`) {
		t.Error("Header does not contain ChainID")
	}
}

func TestAddServerFault(t *testing.T) {
	block := createTestAdminBlock().(*AdminBlock)

	for i := 0; i < 5; i++ {
		block.ABEntries = append(block.ABEntries, new(AddFederatedServer))
		block.ABEntries = append(block.ABEntries, new(DBSignatureEntry))
		block.ABEntries = append(block.ABEntries, new(EndOfMinuteEntry))
		block.ABEntries = append(block.ABEntries, new(IncreaseServerCount))
		block.ABEntries = append(block.ABEntries, new(AddFederatedServerBitcoinAnchorKey))
	}

	for i := 0; i < 10; i++ {
		sf := new(ServerFault)

		sf.Timestamp = primitives.NewTimestampFromMinutes(uint32(i * 2))
		sf.ServerID = testHelper.NewRepeatingHash(1)
		sf.AuditServerID = testHelper.NewRepeatingHash(2)

		sf.VMIndex = 5
		sf.DBHeight = 0x44556677
		sf.Height = 0x88990011

		block.AddServerFault(sf)
	}

	for i := 0; i < 10; i++ {
		sf := new(ServerFault)

		sf.Timestamp = primitives.NewTimestampFromMinutes(uint32(i*2 + 1))
		sf.ServerID = testHelper.NewRepeatingHash(1)
		sf.AuditServerID = testHelper.NewRepeatingHash(2)

		sf.VMIndex = 5
		sf.DBHeight = 0x44556677
		sf.Height = 0x88990011

		block.AddServerFault(sf)
	}

	for i := 0; i < 20; i++ {
		sf := new(ServerFault)

		sf.Timestamp = primitives.NewTimestampFromMinutes(uint32(i))
		sf.ServerID = testHelper.NewRepeatingHash(1)
		sf.AuditServerID = testHelper.NewRepeatingHash(2)

		sf.VMIndex = byte(i)
		sf.DBHeight = 0x44556677
		sf.Height = 0x88990011

		block.AddServerFault(sf)
	}

	if len(block.ABEntries) != 5*5+10+10+20 {
		t.Errorf("Wrong length of ABEntries - %v", len(block.ABEntries))
	}

	sfFound := false
	for i := range block.ABEntries {
		if block.ABEntries[i].Type() != constants.TYPE_SERVER_FAULT {
			if sfFound {
				t.Errorf("Non-SF entry between SF entries at position %v", i)
			}
			continue
		}
		if i == 0 {
			t.Errorf("SF entry is at position 0 when it shouldn't be")
			continue
		}
		if block.ABEntries[i-1].Type() != constants.TYPE_SERVER_FAULT {
			continue
		}

		prev := block.ABEntries[i-1].(*ServerFault)
		cur := block.ABEntries[i].(*ServerFault)

		if prev.Timestamp.GetTimeMilliUInt64() > cur.Timestamp.GetTimeMilliUInt64() {
			t.Errorf("Wrong order by Timestamp")
			continue
		}
		if prev.Timestamp.GetTimeMilliUInt64() < cur.Timestamp.GetTimeMilliUInt64() {
			continue
		}
		if prev.VMIndex > cur.VMIndex {
			t.Errorf("Wrong order by VMIndex")
			continue
		}
	}
}

func createTestAdminBlock() (block interfaces.IAdminBlock) {
	block = new(AdminBlock)
	block.SetHeader(createTestAdminHeader())
	/**
	p, _ := hex.DecodeString("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
	hash := primitives.Sha(p)
	sigBytes := make([]byte, 96)
	for i := 0; i < 5; i++ {
		for j := range sigBytes {                      // Don't know why this fails.cd
			sigBytes[j] = byte(i)
		}
		sig := primitives.UnmarshalBinarySignature(sigBytes)
		fmt.Println(hash, sig)
		entry := NewDBSignatureEntry(hash, sig)
		var _ = entry
		block.SetABEntries(append(block.GetABEntries(), nil))
	}
	**/
	block.GetHeader().SetMessageCount(uint32(len(block.GetABEntries())))
	return block
}

func createSmallTestAdminBlock() (block interfaces.IAdminBlock) {
	block = new(AdminBlock)
	block.SetHeader(createSmallTestAdminHeader())
	block.GetHeader().SetMessageCount(uint32(len(block.GetABEntries())))
	return block
}

func createTestAdminHeader() *ABlockHeader {
	header := new(ABlockHeader)

	p, _ := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	hash, _ := primitives.NewShaHash(p)
	header.PrevBackRefHash = hash
	header.DBHeight = 123

	header.HeaderExpansionSize = 5
	header.HeaderExpansionArea = []byte{0x00, 0x01, 0x02, 0x03, 0x04}
	header.MessageCount = 234
	header.BodySize = 345

	return header
}

func createSmallTestAdminHeader() *ABlockHeader {
	header := new(ABlockHeader)

	p, _ := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	hash, _ := primitives.NewShaHash(p)
	header.PrevBackRefHash = hash
	header.DBHeight = 123

	header.HeaderExpansionSize = 0
	header.HeaderExpansionArea = []byte{}
	header.MessageCount = 234
	header.BodySize = 345

	return header
}
