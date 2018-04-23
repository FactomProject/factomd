package adminBlock_test

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"sort"

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
		t.Error("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Error("Error is nil when it shouldn't be")
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
		t.Error("No error returned")
	}
	if a.AddDBSig(nil, nil) == nil {
		t.Error("No error returned")
	}
	if a.AddFedServer(nil) == nil {
		t.Error("No error returned")
	}
	if a.AddAuditServer(nil) == nil {
		t.Error("No error returned")
	}
	if a.RemoveFederatedServer(nil) == nil {
		t.Error("No error returned")
	}
	if a.AddMatryoshkaHash(nil, nil) == nil {
		t.Error("No error returned")
	}
	if a.AddFederatedServerSigningKey(nil, [32]byte{}) == nil {
		t.Error("No error returned")
	}
	if a.AddFederatedServerBitcoinAnchorKey(nil, 0, 0, [20]byte{}) == nil {
		t.Error("No error returned")
	}
	if a.AddEntry(nil) == nil {
		t.Error("No error returned")
	}
	if a.AddServerFault(nil) == nil {
		t.Error("No error returned")
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
			t.Error("Invalid IdentityAdminChainID")
		}
		if ab.ABEntries[i].(*DBSignatureEntry).PrevDBSig.IsSameAs(&testVector[i].PrevDBSig) == false {
			t.Error("Invalid PrevDBSig")
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
			t.Error("Invalid IdentityChainID")
		}
		if ab.ABEntries[i].(*AddFederatedServer).DBHeight != uint32(i+1) {
			t.Error("Invalid DBHeight")
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
			t.Error("Invalid IdentityChainID")
		}
		if ab.ABEntries[i].(*AddAuditServer).DBHeight != uint32(i+1) {
			t.Error("Invalid DBHeight")
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
			t.Error("Invalid IdentityChainID")
		}
		if ab.ABEntries[i].(*RemoveFederatedServer).DBHeight != uint32(i+1) {
			t.Error("Invalid DBHeight")
		}
	}
}

func TestAddMatryoshkaHash(t *testing.T) {
	testVector := []interfaces.IIdentityABEntry{}
	for i := 0; i < 1000; i++ {
		se := new(AddReplaceMatryoshkaHash)
		se.IdentityChainID = primitives.RandomHash()
		se.MHash = primitives.RandomHash()
		testVector = append(testVector, se)
	}
	ab := new(AdminBlock)
	for _, v := range testVector {
		av := v.(*AddReplaceMatryoshkaHash)
		err := ab.AddMatryoshkaHash(av.IdentityChainID, av.MHash)
		if err != nil {
			t.Errorf("%v", err)
		}
	}

	sort.Sort(interfaces.IIdentityABEntrySort(testVector))

	ab.InsertIdentityABEntries()
	for i := range testVector {
		av := testVector[i].(*AddReplaceMatryoshkaHash)
		if ab.ABEntries[i].(*AddReplaceMatryoshkaHash).IdentityChainID.String() != av.IdentityChainID.String() {
			t.Error("Invalid IdentityChainID")
		}
		if ab.ABEntries[i].(*AddReplaceMatryoshkaHash).MHash.String() != av.MHash.String() {
			t.Error("Invalid MHash")
		}
	}
}

func TestAddFederatedServerSigningKey(t *testing.T) {
	testVector := []interfaces.IIdentityABEntry{}
	for i := 0; i < 1000; i++ {
		se := new(AddFederatedServerSigningKey)
		se.IdentityChainID = primitives.RandomHash()
		priv := primitives.RandomPrivateKey()
		se.PublicKey = *priv.Pub
		testVector = append(testVector, se)
	}
	ab := new(AdminBlock)
	for _, v := range testVector {
		av := v.(*AddFederatedServerSigningKey)
		err := ab.AddFederatedServerSigningKey(av.IdentityChainID, av.PublicKey.Fixed())
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	sort.Sort(interfaces.IIdentityABEntrySort(testVector))
	ab.InsertIdentityABEntries()
	for i := range testVector {
		av := testVector[i].(*AddFederatedServerSigningKey)

		if ab.ABEntries[i].(*AddFederatedServerSigningKey).IdentityChainID.String() != av.IdentityChainID.String() {
			t.Error("Invalid IdentityChainID")
		}
		if primitives.AreBytesEqual(ab.ABEntries[i].(*AddFederatedServerSigningKey).PublicKey[:], av.PublicKey[:]) == false {
			t.Error("Invalid PublicKey")
		}
	}
}

func TestAddFederatedServerBitcoinAnchorKey(t *testing.T) {
	testVector := []interfaces.IIdentityABEntry{}
	for i := 0; i < 1000; i++ {
		se := new(AddFederatedServerBitcoinAnchorKey)
		se.IdentityChainID = primitives.RandomHash()
		b := [20]byte{}
		copy(b[:], random.RandByteSliceOfLen(20))
		testVector = append(testVector, se)
	}
	ab := new(AdminBlock)
	for i, v := range testVector {
		av := v.(*AddFederatedServerBitcoinAnchorKey)
		fixed, err := av.ECDSAPublicKey.GetFixed()
		if err != nil {
			t.Errorf("%v", err)
		}

		err = ab.AddFederatedServerBitcoinAnchorKey(av.IdentityChainID, byte(i%256), byte(256-i%256), fixed)
		if err != nil {
			t.Errorf("%v", err)
		}
	}

	sort.Sort(interfaces.IIdentityABEntrySort(testVector))
	ab.InsertIdentityABEntries()
	for i := range testVector {
		av := testVector[i].(*AddFederatedServerBitcoinAnchorKey)

		if ab.ABEntries[i].(*AddFederatedServerBitcoinAnchorKey).IdentityChainID.String() != av.IdentityChainID.String() {
			t.Error("Invalid IdentityChainID")
		}
		if primitives.AreBytesEqual(ab.ABEntries[i].(*AddFederatedServerBitcoinAnchorKey).ECDSAPublicKey[:], av.ECDSAPublicKey[:]) == false {
			t.Error("Invalid ECDSAPublicKey")
		}
		if ab.ABEntries[i].(*AddFederatedServerBitcoinAnchorKey).KeyPriority != byte(i%256) {
			t.Error("Invalid KeyPriority")
		}
		if ab.ABEntries[i].(*AddFederatedServerBitcoinAnchorKey).KeyType != byte(256-i%256) {
			t.Error("Invalid KeyType")
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
	if strings.Contains(j, `"backreferencehash":"9515e5108c89ef004ff4fa01c6511f98c8c11f5c2976c4816f8bcfcc551a134d"`) == false {
		t.Errorf("JSON printout does not contain the backreference hash - %v", j)
	}
	if strings.Contains(j, `"lookuphash":"f10eefb55197e34f2875c1727c816fcf6564a44902b716a380f0961406ff92d5"`) == false {
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
		str1, err1 := block.JSONString()
		str2, err2 := block2.JSONString()
		if err1 != nil || err2 != nil || str1 != str2 {
			t.Errorf("JSON doesn't match. %d", b)
		}
		b1, err1 := block.JSONByte()
		b2, err2 := block.JSONByte()
		if err1 != nil || err2 != nil || bytes.Compare(b1, b2) != 0 {
			t.Errorf("JSON Byte doesn't match. %d", b)
		}
		if block.String() != block2.String() {
			t.Errorf("String representation doesn't match %d", b)
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

type TestABlock struct {
	Raw   string
	KeyMR string
	Hash  string
}

func TestUnmarshalABlock(t *testing.T) {
	ts := []TestABlock{}

	t1 := TestABlock{}
	t1.Raw = "000000000000000000000000000000000000000000000000000000000000000a0a9aa1efbe7d0e8d9c1d460d1c78e3e7b50f984e65a3f3ee7b73100a94189dbf000000010000000002000000830100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a83efbcbed19b5842e5aa06e66c41d8b61826d95d50c1cbc8bd5373f986c370547133462a9ffa0dcff025a6ad26747c95f1bdd88e2596fc8c6eaa8a2993c72c050002"
	t1.KeyMR = "b30ab81a8afdbe0be1627ef151bf7e263ce3d39d60b61464d81daa8320c28a4f"
	t1.Hash = "b2405450392038716e9b24804345f9ac0736792dba436c024268ed8100683894"
	ts = append(ts, t1)

	t2 := TestABlock{}
	t2.Raw = "000000000000000000000000000000000000000000000000000000000000000a4a994fbc66747f91e28576e715de510bb12937ee68885ffc0a70ad3b03b375600001131100000000490000107706888888655866a003faabd999c7b0a7c908af17d63fd2ac2951dc99e1ad2a14f40001131208888888655866a003faabd999c7b0a7c908af17d63fd2ac2951dc99e1ad2a14f400f8139f98fadc948b254d0dea29c55fab7fa14f1fd97ef78ef7bb99d2d82bd6f10001131203888888655866a003faabd999c7b0a7c908af17d63fd2ac2951dc99e1ad2a14f486e2f9073dfafb461888955166c12c6b1d9aa98504af1cccb08f0ad53fbbb66609888888655866a003faabd999c7b0a7c908af17d63fd2ac2951dc99e1ad2a14f400005bf09c36ebb93643acf41e716261357583ee728106888888dda15d7ad44c3286d66cc4f82e6fc07ed88de4d13ac9a182199593cac10001131208888888dda15d7ad44c3286d66cc4f82e6fc07ed88de4d13ac9a182199593cac10007f339e556ee999cc7e33500753ea0933381b09f5c2bca26e224d716e61a88620001131203888888dda15d7ad44c3286d66cc4f82e6fc07ed88de4d13ac9a182199593cac17c45e29fd0c7e09428e7ea60ed5042e8a0d6a091cc576e255eb10b7e899d3c0309888888dda15d7ad44c3286d66cc4f82e6fc07ed88de4d13ac9a182199593cac100006788c85b7963c8527900a2a2ad2c24d15f347d89068888889585051d7117d217a55a366d56826eda35c951f02428b976524dbfc7f900011312088888889585051d7117d217a55a366d56826eda35c951f02428b976524dbfc7f9002001c69d076a5bf43335d41f49ad7626f1d79d8e1dfe9d9f9c8cc9a0d99efd5b00011312038888889585051d7117d217a55a366d56826eda35c951f02428b976524dbfc7f9914ab0fd1905f3ef19e54f94dd3caee1055793eb8cd5ce7f982cd15ea393bcd7098888889585051d7117d217a55a366d56826eda35c951f02428b976524dbfc7f90000c568a1206e29c7c8fed15aee12515833434b4eb406888888bf5e39211db27b2d2b1b57606b4d68cf57e908971949a233d8eb7341560001131208888888bf5e39211db27b2d2b1b57606b4d68cf57e908971949a233d8eb73415600646f6bf2eaa80a803f1ffd3286945c4d6ddfdf5974177a52141c6906153f52370001131203888888bf5e39211db27b2d2b1b57606b4d68cf57e908971949a233d8eb734156002762ccf5948b8e1c29a9c3df4748cf3efe6567eb3046e6361f353079e5534409888888bf5e39211db27b2d2b1b57606b4d68cf57e908971949a233d8eb73415600007c6b5121835d148932c75ce773208ffc17a4144f068888886ff14cef50365b785eb3cefab5bc30175d022be06ed412391a8264537600011312088888886ff14cef50365b785eb3cefab5bc30175d022be06ed412391a82645376000d6a22b9bf17851c830189fb324ba7d1ea8d6a15eea3adf671109825a133214700011312038888886ff14cef50365b785eb3cefab5bc30175d022be06ed412391a82645376fe21f1320ff7eaaab9ceb9551833078ab79b5b0dfe86097a88ca26d74e48b354098888886ff14cef50365b785eb3cefab5bc30175d022be06ed412391a826453760000b4db03e03da3555f630aef3900897e67247c847706888888b1255ea1cc0b3ab3b9425d3239643ae1f3c00ec634690eda784f05bda70001131208888888b1255ea1cc0b3ab3b9425d3239643ae1f3c00ec634690eda784f05bda700e3b88b704533612f69b5d6390737481694d7d8acb71e532cac3e8dd2d11ca6910001131203888888b1255ea1cc0b3ab3b9425d3239643ae1f3c00ec634690eda784f05bda71cbf54effa547cf89751e3a02d8980ea7e9325e591ff8f1d360bbe323da8fa5a09888888b1255ea1cc0b3ab3b9425d3239643ae1f3c00ec634690eda784f05bda70000dcb4dcd7e5a518854eadd0ec48955101d9fbac350688888841ac82c501a300def3e95d724b4b5e31f729f3b6d9d9736dca0f0edc34000113120888888841ac82c501a300def3e95d724b4b5e31f729f3b6d9d9736dca0f0edc3400667a53519cab0365d1a1ac625b6cd64d86695e8ae38d280ea6d3dbe8191acf34000113120388888841ac82c501a300def3e95d724b4b5e31f729f3b6d9d9736dca0f0edc3452103541ebcd32f5a55dc3c5037fd6396bbe3d65d22f8c06026a9ad97440d8cd0988888841ac82c501a300def3e95d724b4b5e31f729f3b6d9d9736dca0f0edc3400005ba2689c372fdf712e477a83059a5da313e07bf0068888882fa588e8ad6e73555a9b9ff3d84b468601b81328ec09d91051369d737300011312088888882fa588e8ad6e73555a9b9ff3d84b468601b81328ec09d91051369d737300296d08be4a741d6c328ab47d80a55590dceef6550066a0a76e4816a3f51eefee00011312038888882fa588e8ad6e73555a9b9ff3d84b468601b81328ec09d91051369d7373a5f91355b6c8a1a9b38d378434886caea05cc73e544416ec4c9b7f219f23c497098888882fa588e8ad6e73555a9b9ff3d84b468601b81328ec09d91051369d73730000850fd39e1841b29c12f4ace379380a467489dba806888888b4eecb6868615e1875120e855529b4e372e2887cdec7185b46abfcfb350001131208888888b4eecb6868615e1875120e855529b4e372e2887cdec7185b46abfcfb3500c2bbab9d274415765eae5c3ee3b94ff3c38dd5c9b02c8f842e2770a6de0b50680001131203888888b4eecb6868615e1875120e855529b4e372e2887cdec7185b46abfcfb3586400145400bf22a717d1bd4fc7f15e5de2872d21e815bc0a4916c15de2e6eb709888888b4eecb6868615e1875120e855529b4e372e2887cdec7185b46abfcfb350000e0e135c1ee0c2131b2dac5fcb353863ac21fff62068888884a0acbf1a23e3291b99681b80a91ca51914d64e39de65645868e0b471400011312088888884a0acbf1a23e3291b99681b80a91ca51914d64e39de65645868e0b47140093f6aca96b011fc31fd655fee9556b459509308eaaa63c02e9ebff8f384c72e000011312038888884a0acbf1a23e3291b99681b80a91ca51914d64e39de65645868e0b471435b100ead1d81fe3a3e6b1a656c127b14a2ef9d520adec6ea0d7b9d1d5488268098888884a0acbf1a23e3291b99681b80a91ca51914d64e39de65645868e0b4714000058e737d93cb52102d78ee7b918bd33a4412f901e06888888f05308313f6e8f5619cacfb32e0dcba25b4741de9c0fc3b127e8ba2a6b0001131208888888f05308313f6e8f5619cacfb32e0dcba25b4741de9c0fc3b127e8ba2a6b006ceeb261cc19b14f6c89bb0bd937f195ffc9e6adaa5618e432752b01a00792c70001131203888888f05308313f6e8f5619cacfb32e0dcba25b4741de9c0fc3b127e8ba2a6b8fcab189bbb2f97249d05b0b31adeaef23b7aaca326673e16fc901022f8285c809888888f05308313f6e8f5619cacfb32e0dcba25b4741de9c0fc3b127e8ba2a6b000057b3621913fd321c4c4f07cef3468bf04b0baf59068888881541fc5bc1bcc0597e71eed5df7de8d47c8eb97f867d16ebb20781f38b00011312088888881541fc5bc1bcc0597e71eed5df7de8d47c8eb97f867d16ebb20781f38b0034ffc2a7f6e35e503fd2d4259113d4d9b131e8e56d63a1c277ab5064d58d982600011312038888881541fc5bc1bcc0597e71eed5df7de8d47c8eb97f867d16ebb20781f38b1ce468172d6408643a8931838a935733f6fa97d02a8b44a741a1376da8829152098888881541fc5bc1bcc0597e71eed5df7de8d47c8eb97f867d16ebb20781f38b0000010c53bd5e4a863cf8e7df48f567e3f2e492aba906888888c1fd1cf7ca3e0c4e2e9a6462aa8ac4e537563ee54ff41eb6b617a1ec370001131208888888c1fd1cf7ca3e0c4e2e9a6462aa8ac4e537563ee54ff41eb6b617a1ec3700b9a4837383cf11d818f1c1931f5586f840967fe0931d9b733394f75bf39fcd170001131203888888c1fd1cf7ca3e0c4e2e9a6462aa8ac4e537563ee54ff41eb6b617a1ec3796fa0827f28ced76f18e42b8ef836d96c5c5adde4b8c98a406ad00610998562809888888c1fd1cf7ca3e0c4e2e9a6462aa8ac4e537563ee54ff41eb6b617a1ec3700001f605e0d687dbb731e6961cdf8c30e24195889d006888888b2ddad8c24fdf3033bdf6bd9c393ffe074b6f5d5741c84afea27c1656e0001131208888888b2ddad8c24fdf3033bdf6bd9c393ffe074b6f5d5741c84afea27c1656e00b11d2c22e96af34946810c816ada60a7027ed3d7c98aac72283ed348fc58cf730001131203888888b2ddad8c24fdf3033bdf6bd9c393ffe074b6f5d5741c84afea27c1656e74055ead8eb83d34515c66bb7824dfda3659e1193dd31f6f38eed6e2cdc4e59209888888b2ddad8c24fdf3033bdf6bd9c393ffe074b6f5d5741c84afea27c1656e000072f4aa05adc0b5284602bd744858106c618b932e06888888a8da713519881065d90f73f498b36d956e3390c5a6c06747922395075f0001131208888888a8da713519881065d90f73f498b36d956e3390c5a6c06747922395075f00ffb9efd4d490535e3b5041622354f5c440524b0d1976582e0c9ba6cb1649279b0001131203888888a8da713519881065d90f73f498b36d956e3390c5a6c06747922395075f36108e2fd7ba67a25886c14408db1bc2a1d0098a23f2b64e4734ff80b772def009888888a8da713519881065d90f73f498b36d956e3390c5a6c06747922395075f00003d5ffebea388ce494cd7d24ff03165117561ef9006888888a5ce32a3a257c1ff29033b6c01dd20239e7c68ebaf06d690e4ae2b7e830001131208888888a5ce32a3a257c1ff29033b6c01dd20239e7c68ebaf06d690e4ae2b7e830013d42208f7a7699c7976dc19424872268e503779850fb72aecae4b5341dd40c70001131203888888a5ce32a3a257c1ff29033b6c01dd20239e7c68ebaf06d690e4ae2b7e83611fb3b711629ee6964f6e6d7a7a389ab275b4b14c8eafaaa72930f2b9c1230309888888a5ce32a3a257c1ff29033b6c01dd20239e7c68ebaf06d690e4ae2b7e830000412945af7b4ec2ff17285b22631be19f3201d572068888886043746fe47dcf55952b20d8aee6ae842024010fd3f42bc0076a502f4200011312088888886043746fe47dcf55952b20d8aee6ae842024010fd3f42bc0076a502f4200847ef7a9d15df05940a97030a7b783fad54622bdb81f5698f948b94e127eb6e500011312038888886043746fe47dcf55952b20d8aee6ae842024010fd3f42bc0076a502f42b566f30f2013dc3cf7960268da70efb76534ce710f270c1b3ae08781f9faae1b098888886043746fe47dcf55952b20d8aee6ae842024010fd3f42bc0076a502f420000e2a977f66a529d3746727f390c429298f6daef680688888870cf06fb3ed94af3ba917dbe9fab73391915c57812bd6606e6ced76d53000113120888888870cf06fb3ed94af3ba917dbe9fab73391915c57812bd6606e6ced76d53005413e626ce80d90276b5b2388d13f4a4dce2faffce6bb76b9290fcd11dd700dc000113120388888870cf06fb3ed94af3ba917dbe9fab73391915c57812bd6606e6ced76d53151253cf6f9ad8db3f1bd7116a6ec894851fff4268ad1c14fe3ce8f3933a9b080988888870cf06fb3ed94af3ba917dbe9fab73391915c57812bd6606e6ced76d53000080b560002d85154fa1c255531c232f84b4293c860100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a9c798bf48022895c046a4e780fc43a4f112d90d79889a335ea88336e1cb01e47862bb4504a70cdd3055c9718b2b68166142229f66d772cccb2cac5e5b8dd6c09"
	t2.KeyMR = "748a13e79aa35130ea193141ee7849b5cc7ffcceb1aa77d58cb62c129170ca79"
	t2.Hash = "4f4ba20e4d8e62dd10827b20523f084ed3d5a90164bd06b95557109820ae0416"
	ts = append(ts, t2)

	t3 := TestABlock{}
	t3.Raw = "000000000000000000000000000000000000000000000000000000000000000a9ac526b050fbebf858a8e755e23655d3ccb53c256dac855304c74989de15d1810001131300000000290000095f05888888655866a003faabd999c7b0a7c908af17d63fd2ac2951dc99e1ad2a14f40001131408888888655866a003faabd999c7b0a7c908af17d63fd2ac2951dc99e1ad2a14f400f8139f98fadc948b254d0dea29c55fab7fa14f1fd97ef78ef7bb99d2d82bd6f10001131403888888655866a003faabd999c7b0a7c908af17d63fd2ac2951dc99e1ad2a14f486e2f9073dfafb461888955166c12c6b1d9aa98504af1cccb08f0ad53fbbb66609888888655866a003faabd999c7b0a7c908af17d63fd2ac2951dc99e1ad2a14f400005bf09c36ebb93643acf41e716261357583ee728105888888dda15d7ad44c3286d66cc4f82e6fc07ed88de4d13ac9a182199593cac10001131408888888dda15d7ad44c3286d66cc4f82e6fc07ed88de4d13ac9a182199593cac10007f339e556ee999cc7e33500753ea0933381b09f5c2bca26e224d716e61a88620001131403888888dda15d7ad44c3286d66cc4f82e6fc07ed88de4d13ac9a182199593cac17c45e29fd0c7e09428e7ea60ed5042e8a0d6a091cc576e255eb10b7e899d3c0309888888dda15d7ad44c3286d66cc4f82e6fc07ed88de4d13ac9a182199593cac100006788c85b7963c8527900a2a2ad2c24d15f347d89058888889585051d7117d217a55a366d56826eda35c951f02428b976524dbfc7f900011314088888889585051d7117d217a55a366d56826eda35c951f02428b976524dbfc7f9002001c69d076a5bf43335d41f49ad7626f1d79d8e1dfe9d9f9c8cc9a0d99efd5b00011314038888889585051d7117d217a55a366d56826eda35c951f02428b976524dbfc7f9914ab0fd1905f3ef19e54f94dd3caee1055793eb8cd5ce7f982cd15ea393bcd7098888889585051d7117d217a55a366d56826eda35c951f02428b976524dbfc7f90000c568a1206e29c7c8fed15aee12515833434b4eb405888888bf5e39211db27b2d2b1b57606b4d68cf57e908971949a233d8eb7341560001131408888888bf5e39211db27b2d2b1b57606b4d68cf57e908971949a233d8eb73415600646f6bf2eaa80a803f1ffd3286945c4d6ddfdf5974177a52141c6906153f52370001131403888888bf5e39211db27b2d2b1b57606b4d68cf57e908971949a233d8eb734156002762ccf5948b8e1c29a9c3df4748cf3efe6567eb3046e6361f353079e5534409888888bf5e39211db27b2d2b1b57606b4d68cf57e908971949a233d8eb73415600007c6b5121835d148932c75ce773208ffc17a4144f058888886ff14cef50365b785eb3cefab5bc30175d022be06ed412391a8264537600011314088888886ff14cef50365b785eb3cefab5bc30175d022be06ed412391a82645376000d6a22b9bf17851c830189fb324ba7d1ea8d6a15eea3adf671109825a133214700011314038888886ff14cef50365b785eb3cefab5bc30175d022be06ed412391a82645376fe21f1320ff7eaaab9ceb9551833078ab79b5b0dfe86097a88ca26d74e48b354098888886ff14cef50365b785eb3cefab5bc30175d022be06ed412391a826453760000b4db03e03da3555f630aef3900897e67247c847705888888b1255ea1cc0b3ab3b9425d3239643ae1f3c00ec634690eda784f05bda70001131408888888b1255ea1cc0b3ab3b9425d3239643ae1f3c00ec634690eda784f05bda700e3b88b704533612f69b5d6390737481694d7d8acb71e532cac3e8dd2d11ca6910001131403888888b1255ea1cc0b3ab3b9425d3239643ae1f3c00ec634690eda784f05bda71cbf54effa547cf89751e3a02d8980ea7e9325e591ff8f1d360bbe323da8fa5a09888888b1255ea1cc0b3ab3b9425d3239643ae1f3c00ec634690eda784f05bda70000dcb4dcd7e5a518854eadd0ec48955101d9fbac350588888841ac82c501a300def3e95d724b4b5e31f729f3b6d9d9736dca0f0edc34000113140888888841ac82c501a300def3e95d724b4b5e31f729f3b6d9d9736dca0f0edc3400667a53519cab0365d1a1ac625b6cd64d86695e8ae38d280ea6d3dbe8191acf34000113140388888841ac82c501a300def3e95d724b4b5e31f729f3b6d9d9736dca0f0edc3452103541ebcd32f5a55dc3c5037fd6396bbe3d65d22f8c06026a9ad97440d8cd0988888841ac82c501a300def3e95d724b4b5e31f729f3b6d9d9736dca0f0edc3400005ba2689c372fdf712e477a83059a5da313e07bf0058888882fa588e8ad6e73555a9b9ff3d84b468601b81328ec09d91051369d737300011314088888882fa588e8ad6e73555a9b9ff3d84b468601b81328ec09d91051369d737300296d08be4a741d6c328ab47d80a55590dceef6550066a0a76e4816a3f51eefee00011314038888882fa588e8ad6e73555a9b9ff3d84b468601b81328ec09d91051369d7373a5f91355b6c8a1a9b38d378434886caea05cc73e544416ec4c9b7f219f23c497098888882fa588e8ad6e73555a9b9ff3d84b468601b81328ec09d91051369d73730000850fd39e1841b29c12f4ace379380a467489dba805888888b4eecb6868615e1875120e855529b4e372e2887cdec7185b46abfcfb350001131408888888b4eecb6868615e1875120e855529b4e372e2887cdec7185b46abfcfb3500c2bbab9d274415765eae5c3ee3b94ff3c38dd5c9b02c8f842e2770a6de0b50680001131403888888b4eecb6868615e1875120e855529b4e372e2887cdec7185b46abfcfb3586400145400bf22a717d1bd4fc7f15e5de2872d21e815bc0a4916c15de2e6eb709888888b4eecb6868615e1875120e855529b4e372e2887cdec7185b46abfcfb350000e0e135c1ee0c2131b2dac5fcb353863ac21fff62058888884a0acbf1a23e3291b99681b80a91ca51914d64e39de65645868e0b471400011314088888884a0acbf1a23e3291b99681b80a91ca51914d64e39de65645868e0b47140093f6aca96b011fc31fd655fee9556b459509308eaaa63c02e9ebff8f384c72e000011314038888884a0acbf1a23e3291b99681b80a91ca51914d64e39de65645868e0b471435b100ead1d81fe3a3e6b1a656c127b14a2ef9d520adec6ea0d7b9d1d5488268098888884a0acbf1a23e3291b99681b80a91ca51914d64e39de65645868e0b4714000058e737d93cb52102d78ee7b918bd33a4412f901e0100000000000000000000000000000000000000000000000000000000000000000426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613ab50980c62e82c822a5801de348ebbc8072510654737f45420adf19a273cbc50657fe854181e0597a71ecce36ae3763c7f0bcfea270d441212a15ebf6f313e707"
	t3.KeyMR = "c4994ca612791460f4687d68cc351bdb183636d2f5300dbf3b8e58811171b39c"
	t3.Hash = "336c9f4c143be396afb2fb112e18777da000883e576d2c9801c51a0f1d7cb7bf"
	ts = append(ts, t3)

	for _, tBlock := range ts {
		raw := tBlock.Raw
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
		if h.String() != tBlock.KeyMR {
			t.Error("Invalid Hash")
		}
		h, err = a.BackReferenceHash()
		if err != nil {
			t.Error(err)
		}
		if h.String() != tBlock.Hash {
			t.Error("Invalid KeyMR")
		}
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
	if !strings.Contains(j, `"adminchainid":"000000000000000000000000000000000000000000000000000000000000000a"`) {
		t.Error("Header does not contain AdminChainID")
	}
	if !strings.Contains(j, `"chainid":"000000000000000000000000000000000000000000000000000000000000000a"`) {
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
			t.Error("SF entry is at position 0 when it shouldn't be")
			continue
		}
		if block.ABEntries[i-1].Type() != constants.TYPE_SERVER_FAULT {
			continue
		}

		prev := block.ABEntries[i-1].(*ServerFault)
		cur := block.ABEntries[i].(*ServerFault)

		if prev.Timestamp.GetTimeMilliUInt64() > cur.Timestamp.GetTimeMilliUInt64() {
			t.Error("Wrong order by Timestamp")
			continue
		}
		if prev.Timestamp.GetTimeMilliUInt64() < cur.Timestamp.GetTimeMilliUInt64() {
			continue
		}
		if prev.VMIndex > cur.VMIndex {
			t.Error("Wrong order by VMIndex")
			continue
		}
	}
}

func createTestAdminBlock() (block interfaces.IAdminBlock) {
	block = new(AdminBlock)
	block.(*AdminBlock).Init()
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
	block.(*AdminBlock).Init()
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

func TestABlockVec(t *testing.T) {
	vec1 := "000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000000009c00000000090000035c018888881570f89283f3a516b6e5ed240f43f5ad7cb05132378c4a006abe7c2b93803b318b23ec15de43db470200c1afb5d1b6156184e247ed035a8f0b6879155b17eb2403f90fe408aa2d83e748a70809299ac7b22b3e92d261b36552d0f7169ccd5044adcdd74e299b2c21bbb586626c20289267ff625d6bb1ee525dbba7490f018888888da6ed14ec63e623cab6917c66b954b361d530770b3f5f5188f87f173811cae6d21e92d9ac0ee83e00f89a3aabde7e3c6f90824339281cfeb93c1377cdba244108175cf299f4db909821adaa42e3346a8b199df30ea391b570bf1cdcbcbb4de04a80de9b8e91885038089bed9046b9c919c90b2da3c0800f89826e3b0d01888888aeaac80d825ac9675cf3a6591916883bd9947e16ab752d39164d80a60815688e940b854d71411dd8dead29843932fc79c9c99cfb69ca6888b29cd13237c7f20041a9d8cd306036253e778a6f833ef75c526aaaf996e6fe8472cc5b7b6b2c58f3695b828550ec247f1c337567a2ab8d184c5c37dd4cf7ae5fb6c98045050138bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a7119b2b8a031ff1ed1a3ce913ab91faf98a3ced5718cf6ce839aef61197c56a7babdb9a09d4f88f22ffa14da4120cd8e95302cb9ac10256b518c2c92b3637a070d4120888888f0b7e308974afc34b2c7f703f25ed2699cb05f818e84e8745644896c551ac603582c31ba6f72a329cea154637d3b372ad538c43a52d3bf3c269cee00240d41208888887020255b631764fc0fd524dac89ae96db264c572391a3f19fcf0f8991e126cf7a9192bae19b3f6920c5f4e0508b87f4d852958641c9b262d5f1727fd520d41208888888da6ed14ec63e623cab6917c66b954b361d530770b3f5f5188f87f17382dd83759f2b90e2083c90399ff8cc5b272b9633b0782bb1daa7a6c11f1b0393f0d4120888888aeaac80d825ac9675cf3a6591916883bd9947e16ab752d39164d80a6089c503dc3d7729b328ad5fc78b43e88ba06bc22736a8c4dc34e3a16a2e2e557e50b4a82a8a9d800641a03e954418966623b4cd6d1477da2cda11ef13d7ed9efb3e1364b1f8c707a82a7fedc001056f38d2e49abf8520f2d447c7b4f8c046add2d2244be5f95a973b6d423e428"
	vec2 := "000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000000009c00000000090000035c018888881570f89283f3a516b6e5ed240f43f5ad7cb05132378c4a006abe7c2b93803b318b23ec15de43db470200c1afb5d1b6156184e247ed035a8f0b6879155b17eb2403f90fe408aa2d83e748a70809299ac7b22b3e92d261b36552d0f7169ccd5044adcdd74e299b2c21bbb586626c20289267ff625d6bb1ee525dbba7490f018888888da6ed14ec63e623cab6917c66b954b361d530770b3f5f5188f87f173811cae6d21e92d9ac0ee83e00f89a3aabde7e3c6f90824339281cfeb93c1377cdba244108175cf299f4db909821adaa42e3346a8b199df30ea391b570bf1cdcbcbb4de04a80de9b8e91885038089bed9046b9c919c90b2da3c0800f89826e3b0d01888888aeaac80d825ac9675cf3a6591916883bd9947e16ab752d39164d80a60815688e940b854d71411dd8dead29843932fc79c9c99cfb69ca6888b29cd13237c7f20041a9d8cd306036253e778a6f833ef75c526aaaf996e6fe8472cc5b7b6b2c58f3695b828550ec247f1c337567a2ab8d184c5c37dd4cf7ae5fb6c98045050138bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a7119b2b8a031ff1ed1a3ce913ab91faf98a3ced5718cf6ce839aef61197c56a7babdb9a09d4f88f22ffa14da4120cd8e95302cb9ac10256b518c2c92b3637a070d41208888888da6ed14ec63e623cab6917c66b954b361d530770b3f5f5188f87f17382dd83759f2b90e2083c90399ff8cc5b272b9633b0782bb1daa7a6c11f1b0393f0d4120888888aeaac80d825ac9675cf3a6591916883bd9947e16ab752d39164d80a6089c503dc3d7729b328ad5fc78b43e88ba06bc22736a8c4dc34e3a16a2e2e557e50d4120888888f0b7e308974afc34b2c7f703f25ed2699cb05f818e84e8745644896c551ac603582c31ba6f72a329cea154637d3b372ad538c43a52d3bf3c269cee00240d41208888887020255b631764fc0fd524dac89ae96db264c572391a3f19fcf0f8991e126cf7a9192bae19b3f6920c5f4e0508b87f4d852958641c9b262d5f1727fd520b4a82a8a9d800641a03e954418966623b4cd6d1477da2cda11ef13d7ed9efb3e1364b1f8c707a82a7fedc001056f38d2e49abf8520f2d447c7b4f8c046add2d2244be5f95a973b6d423e428"

	if vec1 == vec2 {
		t.Error("WHY!")
	}

	a1 := NewAdminBlock(nil)
	a2 := NewAdminBlock(nil)

	d1, _ := hex.DecodeString(vec1)
	d2, _ := hex.DecodeString(vec2)

	nd, err := a1.UnmarshalBinaryData(d1)
	if err != nil {
		t.Error(err)
	}
	if len(nd) > 0 {
		t.Errorf("Left over %d bytes", len(nd))
	}
	nd, err = a2.UnmarshalBinaryData(d2)
	if err != nil {
		t.Error(err)
	}
	if len(nd) > 0 {
		t.Errorf("Left over %d bytes", len(nd))
	}

	if !a1.IsSameAs(a2) {
		t.Error("Not Same")
	}

}
