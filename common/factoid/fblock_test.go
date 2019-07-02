// +build all

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

type TestFBlock struct {
	Raw   string
	KeyMR string
	Hash  string
}

func TestMarshalUnmarshal(t *testing.T) {
	ts := []TestFBlock{}

	t1 := TestFBlock{}
	t1.Raw = "000000000000000000000000000000000000000000000000000000000000000f16a82932aa64e6ad45b2749f2abb871fcf3353ab9d4e163c9bd90e5bbd745b59a164ccbb77a21904edc4f2bb753aa60635fb2b60279c06ae01aa211f375417362fb170f73c3961d4218ff806dd75e6e348ca1798a5fc7a99d443fbe2ff939d9900000000000a2be8000000010000000002000000c702014f8a7fcd1b00000002014f8a851657010001e397a1607d4f56c528ab09da5bbf7b37b0b453f43db303730e28e9ebe02657dff431d4f7dfaf840017ef7a21d1a616d65e6b73f3c6a7ad5c49340a6c2592872020ec60767ff00d7d01a5be79b6ada79c0af4d6b7f91234ff321f3b647ed01e02ccbbc0fe9dcc63293482f22455b9756ee4b4db411a5d00e31b689c1bd1abe1d1e887cf4c52e67fc51fe4d9594c24643a91009c6ea91701b5b6df240248c2f39453162b61d71b98270100000000000000000000"
	t1.KeyMR = "aa100f203f159e4369081bb366f6816b302387ec19a4f8b9c98495d97fbe3527"
	t1.Hash = "5810ed83155dfb7b6039323b8a5572cd03166a37d1c3e86d4538c99907a81757"
	ts = append(ts, t1)

	t2 := TestFBlock{}
	t2.Raw = "000000000000000000000000000000000000000000000000000000000000000f1f633b6b38d8982896097da393dedd15fa9765a30161e53056a6f53002d0e42f6f0b71ffcbfc04165b78fafb865d25a593271f5b8031ec967754d4655822574bd59c461bc0f45997d544658843c396e005a4f54b5547be22ff17e1031f7f2224000000000000c35000015fc200000000050000047602015c3b570f9c00000002015c3b5738a7010100ab99e440733a747dfa9f3325e541b24f500831d332c551c2d10a1b65064824854343d741aaf59500733a747dfa9f3325e541b24f500831d332c551c2d10a1b65064824854343d7410120f372dc9d5e2a0d9683ad83874508ae193f2b8d1c9735ce0ee49e29b6260b02fb98ddaecc0af37a69744a9d2919b66ac376d52210f705579669e11d7bd8b84a1998bd9c82ca16bb5ebaffb872112128e1749edb33b912bd144d6bdbdc175d0802015c3b579215020100818afffca610e7511e875844ad95f767384fd79396439c636e2e76c57424b05867456b7ab62fa7d610330fd717584445ac866dc2facd8b856e63bdb8b15b5ed46c0b053b2c6c5c5c3f818afffca610330fd717584445ac866dc2facd8b856e63bdb8b15b5ed46c0b053b2c6c5c5c3f0108f5380fafc0df6dec81132f24d8bbdc20bd321a677e84b1473f91e01bd4386799e9db891e03fea16d02ec3d2e649bfc3624409948973a938e0c4f2c8d57ca13674c6c45b99ab9ebc5d2bab7ba667f92bb1624f077525d6dfaaafd23edb7850f012c94f2bbe49899679c54482eba49bf1d024476845e478f9cce3238f612edd7616f95ca30231c35c6c5c96ed2603383099f648c16b504445e77ec94bc838e2a3e99787424970e6e9434eadd7c47c53103684702dec2e25aba7f2872df73b6ad0702015c3b5793cc020100dd8c94c700b8a8dc17e92a5d6f0accdf32cbad120838fea4fe8d2437ca1b9f0d46ca08a82ca7d610330fd717584445ac866dc2facd8b856e63bdb8b15b5ed46c0b053b2c6c5c5c3fdd8c94c700330fd717584445ac866dc2facd8b856e63bdb8b15b5ed46c0b053b2c6c5c5c3f01a36df6eee4bf43dca30bd01dc9b4ce2b32fcd44e5f2751fae79164fdabeef913c6c2d074a9865801160c1662560e2c9050269cb7931584bc184c6240758869623fb528aa7a7123ee6385073bcc0a67af4117544245b643a3ab08f66c4ad18f0d012c94f2bbe49899679c54482eba49bf1d024476845e478f9cce3238f612edd761965b5760236470938afa385555c5c398df23c08fbce69612a05393a736cd4c14478557019386b1c15230821b4e9db1d0404603c427e06eb05e101a1718011404000002015c3b5a24e102010081a5dadff81c9e2661b5bb8ade3deb2bb7e6b078e5c5e0398ea2b62c2c4cc3c4b3f1b3b5d7a2a7d610330fd717584445ac866dc2facd8b856e63bdb8b15b5ed46c0b053b2c6c5c5c3f81a5dadff81c330fd717584445ac866dc2facd8b856e63bdb8b15b5ed46c0b053b2c6c5c5c3f0141fb6c31249ece9e18b1f7d99285f7002e5cabeb7580fa200ed3630b3d5049feab6768b9e794516ab1cc49b816a680396b7f8b19e213d729c725f298a81e0b670baa8c31493f8ed691da450abb5acddf9a9e32286c4f1e20c07ce2840235880b012c94f2bbe49899679c54482eba49bf1d024476845e478f9cce3238f612edd761bcf6f84ceb25d91848adcf99115a51419dec155b7f09ae4b09aff19a864c2ab45e089ac6794231c2c37b1cef98875906174e8fee9a07d4362ce30fe2bc9073070000000000000000"
	t2.KeyMR = "ac2919000a514726e08b961a8b2443072cb37492a6c88eb81926d84ea189d2e8"
	t2.Hash = "35ac556392f934d702605eac3dac3138cdc134e3f188392afb550ee797d377f9"
	ts = append(ts, t2)

	for _, tBlock := range ts {
		rawStr := tBlock.Raw
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

		if f.DatabasePrimaryIndex().String() != tBlock.KeyMR {
			t.Errorf("Wrong PrimaryIndex - %v vs %v", f.DatabasePrimaryIndex().String(), tBlock.KeyMR)
		}
		if f.DatabaseSecondaryIndex().String() != tBlock.Hash {
			t.Errorf("Wrong SecondaryIndex - %v vs %v", f.DatabaseSecondaryIndex().String(), tBlock.Hash)
		}

		err = f.Validate()
		if err != nil {
			t.Errorf("%v", err)
		}
	}
}

func TestBadFBlockUnmarshal(t *testing.T) {

	t1 := TestFBlock{}

	//bad raw
	t1.Raw = "000000000000000000000000000000000000000000000000000000000000000f16a82932aa64e6ad45b2749f2abb871fcf3353ab9d4e163c9bd90e5bbd745b59a164ccbb77a21904edc4f2bb753aa60635fb2b60279c06ae01aa211f375417362fb170f73c3961d4218ff806dd75e6e348ca1798a5fc7a99d443fbe2ff939d9900000000000a2be80000000100ffffffff000000c702014f8a7fcd1b00000002014f8a851657010001e397a1607d4f56c528ab09da5bbf7b37b0b453f43db303730e28e9ebe02657dff431d4f7dfaf840017ef7a21d1a616d65e6b73f3c6a7ad5c49340a6c2592872020ec60767ff00d7d01a5be79b6ada79c0af4d6b7f91234ff321f3b647ed01e02ccbbc0fe9dcc63293482f22455b9756ee4b4db411a5d00e31b689c1bd1abe1d1e887cf4c52e67fc51fe4d9594c24643a91009c6ea91701b5b6df240248c2f39453162b61d71b98270100000000000000000000"
	// good raw
	// t1.Raw = "000000000000000000000000000000000000000000000000000000000000000f16a82932aa64e6ad45b2749f2abb871fcf3353ab9d4e163c9bd90e5bbd745b59a164ccbb77a21904edc4f2bb753aa60635fb2b60279c06ae01aa211f375417362fb170f73c3961d4218ff806dd75e6e348ca1798a5fc7a99d443fbe2ff939d9900000000000a2be8000000010000000002000000c702014f8a7fcd1b00000002014f8a851657010001e397a1607d4f56c528ab09da5bbf7b37b0b453f43db303730e28e9ebe02657dff431d4f7dfaf840017ef7a21d1a616d65e6b73f3c6a7ad5c49340a6c2592872020ec60767ff00d7d01a5be79b6ada79c0af4d6b7f91234ff321f3b647ed01e02ccbbc0fe9dcc63293482f22455b9756ee4b4db411a5d00e31b689c1bd1abe1d1e887cf4c52e67fc51fe4d9594c24643a91009c6ea91701b5b6df240248c2f39453162b61d71b98270100000000000000000000"

	rawStr := t1.Raw
	raw, err := hex.DecodeString(rawStr)
	if err != nil {
		t.Errorf("%v", err)
	}

	f2 := new(FBlock)
	// we should get an error from unmarshal for the bad data: "uint underflow"
	if err := f2.UnmarshalBinary(raw); err == nil {
		t.Error("FBlock should have errored on unmarshal", f2)
	} else {
		t.Log(err)
	}

}

func TestGetEntryHashes(t *testing.T) {
	f := GetDeterministicFBlock(t)
	hashes := f.GetEntryHashes()
	txs := f.Transactions

	if len(txs) == 0 {
		t.Errorf("No transactions found")
	}
	if len(hashes) != len(txs) {
		t.Errorf("Returned wrong amount of hashes")
		t.FailNow()
	}

	for i := range txs {
		if txs[i].GetHash().IsSameAs(hashes[i]) == false {
			t.Errorf("Hashes are not identical")
		}
	}
}

func TestGetEntrySigHashes(t *testing.T) {
	f := GetDeterministicFBlock(t)
	hashes := f.GetEntrySigHashes()
	txs := f.Transactions
	if len(txs) == 0 {
		t.Errorf("No transactions found")
	}

	if len(hashes) != len(txs) {
		t.Errorf("Returned wrong amount of hashes")
		t.FailNow()
	}

	for i := range txs {
		if txs[i].GetSigHash().IsSameAs(hashes[i]) == false {
			t.Errorf("Hashes are not identical")
		}
	}
}

func TestGetTransactionByHash(t *testing.T) {
	f := GetDeterministicFBlock(t)
	txs := f.Transactions

	if len(txs) == 0 {
		t.Errorf("No transactions found")
	}

	for _, v := range txs {
		tx := f.GetTransactionByHash(v.GetHash())
		if tx == nil {
			t.Errorf("Could not find transaction %v", v.GetHash())
		} else {
			if v.IsSameAs(tx) == false {
				t.Errorf("Transactions are not the same")
			}
		}
	}
}

func GetDeterministicFBlock(t *testing.T) *FBlock {
	rawStr := "000000000000000000000000000000000000000000000000000000000000000f16a82932aa64e6ad45b2749f2abb871fcf3353ab9d4e163c9bd90e5bbd745b59a164ccbb77a21904edc4f2bb753aa60635fb2b60279c06ae01aa211f375417362fb170f73c3961d4218ff806dd75e6e348ca1798a5fc7a99d443fbe2ff939d9900000000000a2be8000000010000000002000000c702014f8a7fcd1b00000002014f8a851657010001e397a1607d4f56c528ab09da5bbf7b37b0b453f43db303730e28e9ebe02657dff431d4f7dfaf840017ef7a21d1a616d65e6b73f3c6a7ad5c49340a6c2592872020ec60767ff00d7d01a5be79b6ada79c0af4d6b7f91234ff321f3b647ed01e02ccbbc0fe9dcc63293482f22455b9756ee4b4db411a5d00e31b689c1bd1abe1d1e887cf4c52e67fc51fe4d9594c24643a91009c6ea91701b5b6df240248c2f39453162b61d71b98270100000000000000000000"
	raw, err := hex.DecodeString(rawStr)
	if err != nil {
		t.Errorf("%v", err)
	}

	f, err := UnmarshalFBlock(raw)
	if err != nil {
		t.Errorf("%v", err)
	}
	return f.(*FBlock)
}

func TestMerkleTrees(t *testing.T) {
	f := GetDeterministicFBlock(t)

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
