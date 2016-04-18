package adminBlock_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestAdminBlockPreviousHash(t *testing.T) {
	fmt.Printf("\n---\nTestAdminBlockMarshalUnmarshal\n---\n")

	block := new(AdminBlock)
	data, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	_, err := block.UnmarshalBinaryData(data)
	if err != nil {
		t.Error(err)
	}

	fullHash, err := block.FullHash()
	if err != nil {
		t.Error(err)
	}

	partialHash, err := block.PartialHash()
	if err != nil {
		t.Error(err)
	}

	t.Logf("Current hashes - %s, %s", fullHash.String(), partialHash.String())

	if fullHash.String() != "0a9aa1efbe7d0e8d9c1d460d1c78e3e7b50f984e65a3f3ee7b73100a94189dbf" {
		t.Error("Invalid fullHash")
	}
	if partialHash.String() != "4fb409d5369fad6aa7768dc620f11cd219f9b885956b631ad050962ca934052e" {
		t.Error("Invalid partialHash")
	}
	/*
		block2, err := CreateAdminBlock(s, block, 5)
		if err != nil {
			t.Error(err)
		}

		fullHash2, err := block2.FullHash()
		if err != nil {
			t.Error(err)
		}

		partialHash2, err := block2.PartialHash()
		if err != nil {
			t.Error(err)
		}

		t.Logf("Second hashes - %s, %s", fullHash2.String(), partialHash2.String())
		t.Logf("Previous hash - %s", block2.Header.PrevFullHash.String())

		marshalled, err := block2.MarshalBinary()
		if err != nil {
			t.Error(err)
		}
		t.Logf("Marshalled - %X", marshalled)

		if block2.Header.PrevFullHash.String() != fullHash.String() {
			t.Error("PrevFullHash does not match ABHash")
		}
	*/
}

func TestAdminBlockMarshalUnmarshal(t *testing.T) {
	fmt.Printf("\n---\nTestAdminBlockMarshalUnmarshal\n---\n")

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
	fmt.Printf("\n---\nTestABlockHeaderMarshalUnmarshal\n---\n")

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

	if bytes.Compare(header.GetPrevFullHash().Bytes(), header2.GetPrevFullHash().Bytes()) != 0 {
		t.Error("PrevFullHashes are not identical")
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
	h, err := a.PartialHash()
	if err != nil {
		t.Error(err)
	}
	if h.String() != "b30ab81a8afdbe0be1627ef151bf7e263ce3d39d60b61464d81daa8320c28a4f" {
		t.Error("Invalid Hash")
	}
	h, err = a.FullHash()
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
	fmt.Printf("\n---\nTestInvalidABlockHeaderUnmarshal\n---\n")

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
	fmt.Printf("\n---\nTestInvalidAdminBlockUnmarshal\n---\n")

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
	header.PrevFullHash = hash
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
	header.PrevFullHash = hash
	header.DBHeight = 123

	header.HeaderExpansionSize = 0
	header.HeaderExpansionArea = []byte{}
	header.MessageCount = 234
	header.BodySize = 345

	return header
}
