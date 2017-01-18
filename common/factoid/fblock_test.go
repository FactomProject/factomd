package factoid_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	. "github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestUnmarshalNilFBlock(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(FBlock)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	rawStr := "000000000000000000000000000000000000000000000000000000000000000f16a82932aa64e6ad45b2749f2abb871fcf3353ab9d4e163c9bd90e5bbd745b59a164ccbb77a21904edc4f2bb753aa60635fb2b60279c06ae01aa211f375417362fb170f73c3961d4218ff806dd75e6e348ca1798a5fc7a99d443fbe2ff939d9900000000000a2be8000000010000000002000000c702014f8a7fcd1b00000002014f8a851657010001e397a1607d4f56c528ab09da5bbf7b37b0b453f43db303730e28e9ebe02657dff431d4f7dfaf840017ef7a21d1a616d65e6b73f3c6a7ad5c49340a6c2592872020ec60767ff00d7d01a5be79b6ada79c0af4d6b7f91234ff321f3b647ed01e02ccbbc0fe9dcc63293482f22455b9756ee4b4db411a5d00e31b689c1bd1abe1d1e887cf4c52e67fc51fe4d9594c24643a91009c6ea91701b5b6df240248c2f39453162b61d71b98270100000000000000000000"
	raw, err := hex.DecodeString(rawStr)
	if err != nil {
		t.Errorf("%v", err)
	}

	f := new(FBlock)
	rest, err := f.UnmarshalBinaryData(raw)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(rest) > 0 {
		t.Errorf("Returned too much data - %x", rest)
	}

	b, err := f.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}

	if primitives.AreBytesEqual(raw, b) == false {
		t.Errorf("Marshalled bytes are not equal - %x vs %x", raw, b)
	}

	//f.CalculateHashes()

	if f.DatabasePrimaryIndex().String() != "aa100f203f159e4369081bb366f6816b302387ec19a4f8b9c98495d97fbe3527" {
		t.Errorf("Wrong PrimaryIndex - %v vs %v", f.DatabasePrimaryIndex().String(), "aa100f203f159e4369081bb366f6816b302387ec19a4f8b9c98495d97fbe3527")
	}
	if f.DatabaseSecondaryIndex().String() != "5810ed83155dfb7b6039323b8a5572cd03166a37d1c3e86d4538c99907a81757" {
		t.Errorf("Wrong SecondaryIndex - %v vs %v", f.DatabaseSecondaryIndex().String(), "5810ed83155dfb7b6039323b8a5572cd03166a37d1c3e86d4538c99907a81757")
	}

	err = f.Validate()
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestMerkleTrees(t *testing.T) {
	rawStr := "000000000000000000000000000000000000000000000000000000000000000f16a82932aa64e6ad45b2749f2abb871fcf3353ab9d4e163c9bd90e5bbd745b59a164ccbb77a21904edc4f2bb753aa60635fb2b60279c06ae01aa211f375417362fb170f73c3961d4218ff806dd75e6e348ca1798a5fc7a99d443fbe2ff939d9900000000000a2be8000000010000000002000000c702014f8a7fcd1b00000002014f8a851657010001e397a1607d4f56c528ab09da5bbf7b37b0b453f43db303730e28e9ebe02657dff431d4f7dfaf840017ef7a21d1a616d65e6b73f3c6a7ad5c49340a6c2592872020ec60767ff00d7d01a5be79b6ada79c0af4d6b7f91234ff321f3b647ed01e02ccbbc0fe9dcc63293482f22455b9756ee4b4db411a5d00e31b689c1bd1abe1d1e887cf4c52e67fc51fe4d9594c24643a91009c6ea91701b5b6df240248c2f39453162b61d71b98270100000000000000000000"
	raw, err := hex.DecodeString(rawStr)
	if err != nil {
		t.Errorf("%v", err)
	}

	f := new(FBlock)
	_, err = f.UnmarshalBinaryData(raw)
	if err != nil {
		t.Errorf("%v", err)
	}

	if f.GetKeyMR().String() != "aa100f203f159e4369081bb366f6816b302387ec19a4f8b9c98495d97fbe3527" {
		t.Errorf("Invalid GetKeyMR")
	}
	if f.GetLedgerKeyMR().String() != "5810ed83155dfb7b6039323b8a5572cd03166a37d1c3e86d4538c99907a81757" {
		t.Errorf("Invalid GetLedgerKeyMR")
	}
	if f.GetLedgerMR().String() != "c9935c01b47250da7be21f9838a45a797b5c1c2c8b12a347bcc7188c4ae9e0e8" {
		t.Errorf("Invalid GetLedgerMR")
	}
	if f.GetBodyMR().String() != "16a82932aa64e6ad45b2749f2abb871fcf3353ab9d4e163c9bd90e5bbd745b59" {
		t.Errorf("Invalid GetBodyMR")
	}
	/*
		t.Errorf("GetKeyMR - %v", f.GetKeyMR().String())
		t.Errorf("GetLedgerKeyMR - %v", f.GetLedgerKeyMR().String())
		t.Errorf("GetLedgerMR - %v", f.GetLedgerMR().String())
		t.Errorf("GetBodyMR - %v", f.GetBodyMR().String())
	*/
}

func TestFBlockDump(t *testing.T) {
	var i uint32
	i = 1
	f := NewFBlock(nil)
	f.SetDBHeight(i)
	str := f.String()
	line := findLine(str, "DBHeight")
	if strings.Contains(line, fmt.Sprintf("%v", i)) == false {
		t.Errorf("Did not find proper height for %v", i)
	}
	i = 10
	f = NewFBlock(nil)
	f.SetDBHeight(i)
	str = f.String()
	line = findLine(str, "DBHeight")
	if strings.Contains(line, fmt.Sprintf("%v", i)) == false {
		t.Errorf("Did not find proper height for %v", i)
	}
	i = 255
	f = NewFBlock(nil)
	f.SetDBHeight(i)
	str = f.String()
	line = findLine(str, "DBHeight")
	if strings.Contains(line, fmt.Sprintf("%v", i)) == false {
		t.Errorf("Did not find proper height for %v", i)
	}
	i = 0xFFFF
	f = NewFBlock(nil)
	f.SetDBHeight(i)
	str = f.String()
	line = findLine(str, "DBHeight")
	if strings.Contains(line, fmt.Sprintf("%v", i)) == false {
		t.Errorf("Did not find proper height for %v", i)
	}
}

func findLine(full, toFind string) string {
	strs := strings.Split(full, "\n")
	for _, v := range strs {
		if strings.Contains(v, toFind) {
			return v
		}
	}
	return ""
}

func TestExpandedDBlockHeader(t *testing.T) {
	block := NewFBlock(nil)
	j, err := block.JSONString()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !strings.Contains(j, `"chainid":"000000000000000000000000000000000000000000000000000000000000000f"`) {
		t.Error("Header does not contain ChainID")
		t.Logf("%v", j)
	}
}
