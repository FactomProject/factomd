package anchor_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	. "github.com/PaulSnow/factom2d/anchor"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
	. "github.com/PaulSnow/factom2d/testHelper"
)

// TestMarshalUnmarshalAnchorRecord unmarshals a test string into an AnchorRecord and marshals it back to a []byte
// and compares with the initial string cast to []byte, tests AnchorRecord V1
func TestMarshalUnmarshalAnchorRecord(t *testing.T) {
	record := `{"AnchorRecordVer":1,"DBHeight":5,"KeyMR":"980ab6d50d9fad574ad4df6dba06a8c02b1c67288ee5beab3fbfde2723f73ef6","RecordHeight":6,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"e2ac71c9c0fd8edc0be8c0ba7098b77fb7d90dcca755d5b9348116f3f9d9f951","BlockHeight":372576,"BlockHash":"000000000000000003059382ed4dd82b2086e99ec78d1b6e811ebb9d53d8656d","Offset":1144}}`
	ar, err := UnmarshalAnchorRecord([]byte(record))
	if err != nil {
		t.Error(err)
	}
	data, err := ar.Marshal()
	if err != nil {
		t.Error(err)
	}

	if primitives.AreBytesEqual([]byte(record), data) == false {
		t.Errorf("Anchors are not equal\n%s\nvs\n%s", record, string(data))
	}
}

// TestMarshalUnmarshalAnchorRecordV2 unmarshals a test string into an AnchorRecord and marshals it with a private key.
// Compares the marshaled string to the original string. Unmarshals the new string into new AnchorRecord and directly
// compares the two AnchorRecords. Tests version 2 AnchorRecords
func TestMarshalUnmarshalAnchorRecordV2(t *testing.T) {
	record := `{"AnchorRecordVer":2,"DBHeight":5,"KeyMR":"980ab6d50d9fad574ad4df6dba06a8c02b1c67288ee5beab3fbfde2723f73ef6","RecordHeight":6,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"e2ac71c9c0fd8edc0be8c0ba7098b77fb7d90dcca755d5b9348116f3f9d9f951","BlockHeight":372576,"BlockHash":"000000000000000003059382ed4dd82b2086e99ec78d1b6e811ebb9d53d8656d","Offset":1144}}`
	ar, err := UnmarshalAnchorRecord([]byte(record))
	if err != nil {
		t.Error(err)
	}

	priv := "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973"
	pk1, err := primitives.NewPrivateKeyFromHex(priv)
	if err != nil {
		t.Error(err)
	}

	data, sig, err := ar.MarshalAndSignV2(pk1)
	if err != nil {
		t.Error(err)
	}

	if primitives.AreBytesEqual([]byte(record), data) == false {
		t.Errorf("Anchors are not equal\n%s\nvs\n%s", record, string(data))
	}

	ar2, valid, err := UnmarshalAndValidateAnchorRecordV2(data, [][]byte{sig}, []interfaces.Verifier{pk1.Pub})
	if err != nil {
		t.Error(err)
	}
	if valid != true {
		t.Errorf("Anchor record is not valid")
	}
	if ar2 == nil {
		t.Errorf("No anchor record returned!")
	}
	if ar.IsSame(ar2) == false {
		t.Errorf("Returned anchor record does not match original")
	}
}

// TestValidateAnchorRecord creates a test signed record and unmarshals it into an AnchorRecord and checks for errors.
// Also creates an invalid record and verifies the code appropriately flags unmarshalled data as invalid. Tests version 1
// AnchorRecords
func TestValidateAnchorRecord(t *testing.T) {
	pub := new(primitives.PublicKey)
	err := pub.UnmarshalText([]byte("0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a"))
	if err != nil {
		t.Error(err)
	}
	signedRecord := `{"AnchorRecordVer":1,"DBHeight":46226,"KeyMR":"831287d96f9955ff438836b26219b8c937e01e668295e8334a5d140cdaacd2fd","RecordHeight":46226,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"9a47f131677e135645cebe74085cd980044eda2abbb05ac09ee79ba8640f368d","BlockHeight":421417,"BlockHash":"000000000000000003781bf1b36dbaa1a71f94e23b4b9dc13024379aaf244d89","Offset":599}}e5d40b8673fbeba9e4e5a4322a514ec28989acef55b6fcdf49eae400b6d302852c2cf6659dc1f3298ffae60f6605666efbcef3ac4947333515084e1269ec0807`

	ar, valid, err := UnmarshalAndValidateAnchorRecord([]byte(signedRecord), []interfaces.Verifier{pub})
	if err != nil {
		t.Error(err)
	}
	if valid != true {
		t.Errorf("Anchor record is not valid")
	}
	if ar == nil {
		t.Errorf("No anchor record returned!")
	}

	invalidRecord := `{"AnchorRecordVer":1,"DBHeight":46226,"KeyMR":"831287d96f9955ff438836b26219b8c937e01e668295e8334a5d140cdaacd2fd","RecordHeight":46226,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"9a47f131677e135645cebe74085cd980044eda2abbb05ac09ee79ba8640f368d","BlockHeight":421417,"BlockHash":"000000000000000003781bf1b36dbaa1a71f94e23b4b9dc13024379aaf244d89","Offset":599}}e5d40b8673fbeba9e4e5a4322a514ec28989acef55b6fcdf49eae400b6d302852c2cf6659dc1f3298ffae60f6605666efbcef3ac4947333515084e1269ec0806`

	ar, valid, err = UnmarshalAndValidateAnchorRecord([]byte(invalidRecord), []interfaces.Verifier{pub})
	if err != nil {
		t.Error(err)
	}
	if valid != false {
		t.Errorf("Anchor record is valid when it shouldn't be")
	}
	if ar != nil {
		t.Errorf("Anchor record returned when it shouldn't be!")
	}
}

// TestValidateAnchorRecordV2 the same steps as above, but uses a V2 function for unmarshaling for testing
// AnchorRecord version 2
func TestValidateAnchorRecordV2(t *testing.T) {
	pub := new(primitives.PublicKey)
	err := pub.UnmarshalText([]byte("0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a"))
	if err != nil {
		t.Error(err)
	}
	// This test is supposed to test v2 anchor records, but its using v1 in string below, however it fails if its changed to 2
	signedRecord := `{"AnchorRecordVer":1,"DBHeight":46226,"KeyMR":"831287d96f9955ff438836b26219b8c937e01e668295e8334a5d140cdaacd2fd","RecordHeight":46226,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"9a47f131677e135645cebe74085cd980044eda2abbb05ac09ee79ba8640f368d","BlockHeight":421417,"BlockHash":"000000000000000003781bf1b36dbaa1a71f94e23b4b9dc13024379aaf244d89","Offset":599}}`
	sig, err := hex.DecodeString("e5d40b8673fbeba9e4e5a4322a514ec28989acef55b6fcdf49eae400b6d302852c2cf6659dc1f3298ffae60f6605666efbcef3ac4947333515084e1269ec0807")
	if err != nil {
		t.Error(err)
	}

	ar, valid, err := UnmarshalAndValidateAnchorRecordV2([]byte(signedRecord), [][]byte{sig}, []interfaces.Verifier{pub})
	if err != nil {
		t.Error(err)
	}
	if valid != true {
		t.Errorf("Anchor record is not valid")
	}
	if ar == nil {
		t.Errorf("No anchor record returned!")
	}

	// This test is supposed to test v2 anchor records, but its using v1 in string below, however it fails if its changed to 2
	invalidRecord := `{"AnchorRecordVer":1,"DBHeight":46226,"KeyMR":"831287d96f9955ff438836b26219b8c937e01e668295e8334a5d140cdaacd2fd","RecordHeight":46226,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"9a47f131677e135645cebe74085cd980044eda2abbb05ac09ee79ba8640f368d","BlockHeight":421417,"BlockHash":"000000000000000003781bf1b36dbaa1a71f94e23b4b9dc13024379aaf244d89","Offset":599}}`
	invalidSig, err := hex.DecodeString("e5d40b8673fbeba9e4e5a4322a514ec28989acef55b6fcdf49eae400b6d302852c2cf6659dc1f3298ffae60f6605666efbcef3ac4947333515084e1269ec0806")

	ar, valid, err = UnmarshalAndValidateAnchorRecordV2([]byte(invalidRecord), [][]byte{invalidSig}, []interfaces.Verifier{pub})
	if err != nil {
		t.Error(err)
	}
	if valid != false {
		t.Errorf("Anchor record is valid when it shouldn't be")
	}
	if ar != nil {
		t.Errorf("Anchor record returned when it shouldn't be!")
	}
}

// TestCreateAndValidateAnchorRecordV1 sets up an AnchorRecord directly, marshals and signs it, and finally unmarshals it
// and checks for validity for AnchorRecord version 1
func TestCreateAndValidateAnchorRecordV1(t *testing.T) {
	dBlock := CreateTestDirectoryBlock(nil)
	height := dBlock.GetHeader().GetDBHeight()

	ar := new(AnchorRecord)
	ar.AnchorRecordVer = 1
	ar.DBHeight = height
	ar.KeyMR = dBlock.DatabasePrimaryIndex().String()
	ar.RecordHeight = ar.DBHeight
	ar.Bitcoin = new(BitcoinStruct)
	ar.Bitcoin.Address = "1HLoD9E4SDFFPDiYfNYnkBLQ85Y51J3Zb1"
	ar.Bitcoin.TXID = fmt.Sprintf("%x", IntToByteSlice(int(height)))
	ar.Bitcoin.BlockHeight = int32(height)
	ar.Bitcoin.BlockHash = fmt.Sprintf("%x", IntToByteSlice(255-int(height)))
	ar.Bitcoin.Offset = int32(height % 10)

	hex, err := ar.MarshalAndSign(NewPrimitivesPrivateKey(0))
	if err != nil {
		t.Errorf("%v", err)
	}

	pubs := []interfaces.Verifier{NewPrimitivesPrivateKey(0).Pub}
	ar2, ok, err := UnmarshalAndValidateAnchorRecord(hex, pubs)
	if err != nil {
		t.Errorf("%v", err)
	}
	if ok == false {
		t.Errorf("Invalid anchor signatures")
	}
	if ar2 == nil {
		t.Errorf("No anchor record unmarshalled.")
	}
}

// TestCreateAndValidateAnchoRecordV2 creates and sets an AnchorRecord directly, marshals and signs its with V2, and finally
// unmarshals it and checks for validity for AnchorRecord version 2
func TestCreateAndValidateAnchorRecordV2(t *testing.T) {
	dBlock := CreateTestDirectoryBlock(nil)
	height := dBlock.GetHeader().GetDBHeight()

	ar := new(AnchorRecord)
	ar.AnchorRecordVer = 2
	ar.DBHeight = height
	ar.KeyMR = dBlock.DatabasePrimaryIndex().String()
	ar.RecordHeight = ar.DBHeight

	ar.Bitcoin = new(BitcoinStruct)
	ar.Bitcoin.Address = "1HLoD9E4SDFFPDiYfNYnkBLQ85Y51J3Zb1"
	ar.Bitcoin.TXID = fmt.Sprintf("%x", IntToByteSlice(int(height)))
	ar.Bitcoin.BlockHeight = int32(height)
	ar.Bitcoin.BlockHash = fmt.Sprintf("%x", IntToByteSlice(255-int(height)))
	ar.Bitcoin.Offset = int32(height % 10)

	hex, exIDs, err := ar.MarshalAndSignV2(NewPrimitivesPrivateKey(0))
	if err != nil {
		t.Errorf("%v", err)
	}

	pubs := []interfaces.Verifier{NewPrimitivesPrivateKey(0).Pub}
	ar2, ok, err := UnmarshalAndValidateAnchorRecordV2(hex, [][]byte{exIDs}, pubs)
	if err != nil {
		t.Errorf("%v", err)
	}
	if ok == false {
		t.Errorf("Invalid anchor signatures")
	}
	if ar2 == nil {
		t.Errorf("No anchor record unmarshalled.")
	}
}
